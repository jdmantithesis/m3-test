// Copyright (c) 2020 Uber Technologies, Inc.
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

package index

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueryOptions(t *testing.T) {
	opts := QueryOptions{
		DocsLimit:          10,
		SeriesLimit:        20,
		IndexChecksumQuery: false,
	}

	assert.False(t, opts.SeriesLimitExceeded(19))
	assert.True(t, opts.SeriesLimitExceeded(20))

	assert.False(t, opts.DocsLimitExceeded(9))
	assert.True(t, opts.DocsLimitExceeded(10))

	assert.True(t, opts.LimitsExceeded(19, 10))
	assert.True(t, opts.LimitsExceeded(20, 9))
	assert.False(t, opts.LimitsExceeded(19, 9))

	assert.False(t, opts.exhaustive(19, 10))
	assert.False(t, opts.exhaustive(20, 9))
	assert.True(t, opts.exhaustive(19, 9))

	assert.Equal(t, "storage/index.block.Query", opts.queryTracepoint())
	assert.Equal(t, "storage.dbNamespace.QueryIDs", opts.NSTracepoint())
	assert.Equal(t, "storage.nsIndex.Query", opts.NSIdxTracepoint())
}

func TestIndexChecksumQueryOptions(t *testing.T) {
	opts := QueryOptions{
		IndexChecksumQuery: true,
	}

	assert.False(t, opts.SeriesLimitExceeded(19))
	assert.False(t, opts.SeriesLimitExceeded(20))

	assert.False(t, opts.DocsLimitExceeded(9))
	assert.False(t, opts.DocsLimitExceeded(10))

	assert.False(t, opts.LimitsExceeded(19, 10))
	assert.False(t, opts.LimitsExceeded(20, 9))
	assert.False(t, opts.LimitsExceeded(19, 9))

	assert.True(t, opts.exhaustive(19, 10))
	assert.True(t, opts.exhaustive(20, 9))
	assert.True(t, opts.exhaustive(19, 9))

	assert.Equal(t, "storage/index.block.IndexChecksum", opts.queryTracepoint())
	assert.Equal(t, "storage.dbNamespace.IndexChecksum", opts.NSTracepoint())
	assert.Equal(t, "storage.nsIndex.IndexChecksum", opts.NSIdxTracepoint())
}

func TestIndexChecksumSnapToTime(t *testing.T) {
	now := time.Now()
	opts := QueryOptions{
		StartInclusive: now,
	}

	blockSize := time.Hour * 2
	assert.Equal(t, now, opts.StartInclusive)
	assert.Equal(t, time.Time{}, opts.EndExclusive)
	opts.SnapToNearestDataBlock(blockSize)
	assert.Equal(t, now.Truncate(blockSize), opts.StartInclusive)
	assert.Equal(t, now.Truncate(blockSize).Add(blockSize), opts.EndExclusive)
}