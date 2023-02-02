// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package client

import (
	"encoding/json"
	"fmt"
	"github.com/m3db/m3/src/cluster/shard"
	"github.com/m3db/m3/src/dbnode/topology"
	"github.com/m3db/m3/src/x/checked"
	xerrors "github.com/m3db/m3/src/x/errors"
	"github.com/m3db/m3/src/x/ident"
	"github.com/m3db/m3/src/x/pool"
	"github.com/m3db/m3/src/x/serialize"
	"log"
	"strconv"
	"sync"
)

// writeOp represents a generic write operation
type writeOp interface {
	op

	ShardID() uint32

	SetCompletionFn(fn completionFn)

	Close()
}

type writeState struct {
	sync.Cond
	sync.Mutex
	refCounter

	consistencyLevel                                  topology.ConsistencyLevel
	shardsLeavingCountTowardsConsistency              bool
	shardsLeavingAndInitiazingCountTowardsConsistency bool
	hostSucessMap                                     map[string]bool
	topoMap                                           topology.Map
	op                                                writeOp
	nsID                                              ident.ID
	tsID                                              ident.ID
	tagEncoder                                        serialize.TagEncoder
	annotation                                        checked.Bytes
	majority, pending                                 int32
	success                                           int32
	errors                                            []error

	queues         []hostQueue
	tagEncoderPool serialize.TagEncoderPool
	pool           *writeStatePool
}

func newWriteState(
	encoderPool serialize.TagEncoderPool,
	pool *writeStatePool,
) *writeState {
	w := &writeState{
		pool:           pool,
		tagEncoderPool: encoderPool,
	}
	w.destructorFn = w.close
	w.L = w
	return w
}

func (w *writeState) close() {
	w.op.Close()

	w.nsID.Finalize()
	w.tsID.Finalize()

	if w.annotation != nil {
		w.annotation.DecRef()
		w.annotation.Finalize()
	}

	if enc := w.tagEncoder; enc != nil {
		enc.Finalize()
	}

	w.op, w.majority, w.pending, w.success = nil, 0, 0, 0
	w.nsID, w.tsID, w.tagEncoder, w.annotation = nil, nil, nil, nil

	for i := range w.errors {
		w.errors[i] = nil
	}
	w.errors = w.errors[:0]

	for i := range w.queues {
		w.queues[i] = nil
	}
	w.queues = w.queues[:0]

	if w.pool == nil {
		return
	}
	w.pool.Put(w)
}

func (w *writeState) completionFn(result interface{}, err error) {
	hostID := result.(topology.Host).ID()
	// NB(bl) panic on invalid result, it indicates a bug in the code

	w.Lock()
	w.pending--

	var wErr error

	if err != nil {
		if IsBadRequestError(err) {
			// Wrap with invalid params and non-retryable so it is
			// not retried.
			err = xerrors.NewInvalidParamsError(err)
			err = xerrors.NewNonRetryableError(err)
		}
		wErr = xerrors.NewRenamedError(err, fmt.Errorf("error writing to host %s: %v", hostID, err))
	} else if hostShardSet, ok := w.topoMap.LookupHostShardSet(hostID); !ok {
		errStr := "missing host shard in writeState completionFn: %s"
		wErr = xerrors.NewRetryableError(fmt.Errorf(errStr, hostID))
	} else if shardState, err := hostShardSet.ShardSet().LookupStateByID(w.op.ShardID()); err != nil {
		errStr := "missing shard %d in host %s"
		wErr = xerrors.NewRetryableError(fmt.Errorf(errStr, w.op.ShardID(), hostID))
	} else {
		available := shardState == shard.Available
		leaving := shardState == shard.Leaving

		// TODO: shardsLeavingCountTowardsConsistency and leavingAndShardsLeavingCountTowardsConsistency both cannot be true
		leavingAndShardsLeavingCountTowardsConsistency := leaving &&
			w.shardsLeavingCountTowardsConsistency
		// NB(bl): Only count writes to available shards towards success.
		// NB(r): If shard is leaving and configured to allow writes to leaving
		// shards to count towards consistency then allow that to count
		// to success.

		log.Printf("Replace node: flag value" + strconv.FormatBool(w.shardsLeavingAndInitiazingCountTowardsConsistency))

		if !available && !leavingAndShardsLeavingCountTowardsConsistency && !w.shardsLeavingAndInitiazingCountTowardsConsistency {
			var errStr string
			switch shardState {
			case shard.Initializing:
				errStr = "shard %d in host %s is not available (initializing)"
			case shard.Leaving:
				errStr = "shard %d in host %s not available (leaving)"
			default:
				errStr = "shard %d in host %s not available (unknown state)"
			}
			wErr = xerrors.NewRetryableError(fmt.Errorf(errStr, w.op.ShardID(), hostID))
		} else if !available && w.shardsLeavingAndInitiazingCountTowardsConsistency {
			var errStr string
			switch shardState {
			case shard.Initializing:
				pairedHostID, ok := w.topoMap.LookupParentHost(hostID, w.op.ShardID())
				if !ok {
					errStr = "shard %d in host %s has no leaving shard"
					wErr = xerrors.NewRetryableError(fmt.Errorf(errStr, w.op.ShardID(), hostID))
				} else {
					log.Printf("Replace node: paired host:" + pairedHostID)
					log.Printf("Replace node: host ID:" + hostID)
					log.Printf("Replace node: shard ID: %d", w.op.ShardID())
					if w.hostSucessMap[pairedHostID] {
						w.success++
						log.Printf("Replace node: success value %d", w.success)
					}
					w.hostSucessMap[hostID] = true
				}
			case shard.Leaving:
				pairedHostID, ok := w.topoMap.LookupChildHost(hostID, w.op.ShardID())
				if !ok {
					errStr = "shard %d in host %s has no initializing shard"
					wErr = xerrors.NewRetryableError(fmt.Errorf(errStr, w.op.ShardID(), hostID))
				} else {
					log.Printf("Replace node: paired host:" + pairedHostID)
					log.Printf("Replace node: host ID:" + hostID)
					log.Printf("Replace node: shard ID: %d", w.op.ShardID())
					if w.hostSucessMap[pairedHostID] {
						w.success++
						log.Printf("Replace node: success value %d", w.success)
					}
					w.hostSucessMap[hostID] = true
				}
			default:
				errStr = "shard %d in host %s not available (unknown state)"
			}
		} else {
			w.success++
		}
	}

	if wErr != nil {
		w.errors = append(w.errors, wErr)
	}

	switch w.consistencyLevel {
	case topology.ConsistencyLevelOne:
		if w.success > 0 || w.pending == 0 {
			w.Signal()
		}
	case topology.ConsistencyLevelMajority:
		if w.success >= w.majority || w.pending == 0 {
			log.Printf("Replace node: got majority for shard: %d", w.op.ShardID())
			map1, _ := json.Marshal(w.hostSucessMap)
			fmt.Println("w.hostSucessMap" + string(map1))
			w.Signal()
		}
	case topology.ConsistencyLevelAll:
		if w.pending == 0 {
			w.Signal()
		}
	}

	w.Unlock()
	w.decRef()
}

type writeStatePool struct {
	pool           pool.ObjectPool
	tagEncoderPool serialize.TagEncoderPool
}

func newWriteStatePool(
	tagEncoderPool serialize.TagEncoderPool,
	opts pool.ObjectPoolOptions,
) *writeStatePool {
	p := pool.NewObjectPool(opts)
	return &writeStatePool{
		pool:           p,
		tagEncoderPool: tagEncoderPool,
	}
}

func (p *writeStatePool) Init() {
	p.pool.Init(func() interface{} {
		return newWriteState(p.tagEncoderPool, p)
	})
}

func (p *writeStatePool) Get() *writeState {
	return p.pool.Get().(*writeState)
}

func (p *writeStatePool) Put(w *writeState) {
	p.pool.Put(w)
}
