// Copyright (c) 2019 Uber Technologies, Inc.
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

package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/m3db/m3/src/dbnode/encoding"
	"github.com/m3db/m3/src/dbnode/generated/proto/annotation"
	"github.com/m3db/m3/src/dbnode/ts"
	"github.com/m3db/m3/src/query/generated/proto/prompb"
	"github.com/m3db/m3/src/query/models"
	"github.com/m3db/m3/src/query/storage/m3/consolidators"
	xerrors "github.com/m3db/m3/src/x/errors"
	xsync "github.com/m3db/m3/src/x/sync"
	xtime "github.com/m3db/m3/src/x/time"
)

const initRawFetchAllocSize = 32

func iteratorToPromResult(
	iter encoding.SeriesIterator,
	tags models.Tags,
	promOptions PromOptions,
) (*prompb.TimeSeries, error) {
	var (
		resolution = xtime.UnixNano(promOptions.Resolution)

		firstDP           = true
		handleResets      = false
		lastDPEmitted     = true
		annotationPayload annotation.Payload

		cumulativeSum float64
		prevDP        ts.Datapoint

		samples = make([]prompb.Sample, 0, initRawFetchAllocSize)
	)
	fmt.Println("iteratorToPromResult")
	for iter.Next() {
		dp, _, annotationData := iter.Current()

		if firstDP && len(annotationData) > 0 && resolution > 0 {
			if err := annotationPayload.Unmarshal(annotationData); err != nil {
				return nil, err
			}
			handleResets = annotationPayload.HandleValueResets
			fmt.Printf("handleResets set to %t on %s, resolution %s\n", handleResets, dp.TimestampNanos.ToTime(), resolution)
		}

		firstDP = false

		if handleResets {
			lastDPEmitted = false
			if dp.TimestampNanos%resolution != prevDP.TimestampNanos%resolution && !firstDP {
				samples = append(samples, prompb.Sample{
					Timestamp: TimeToPromTimestamp(prevDP.TimestampNanos),
					Value:     cumulativeSum,
				})
				lastDPEmitted = true
			}

			if dp.Value <= prevDP.Value { // counter reset
				cumulativeSum += dp.Value
			} else {
				cumulativeSum += dp.Value - prevDP.Value
			}

			prevDP = dp

		} else {
			samples = append(samples, prompb.Sample{
				Timestamp: TimeToPromTimestamp(dp.TimestampNanos),
				Value:     dp.Value,
			})
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	if handleResets && !lastDPEmitted {
		samples = append(samples, prompb.Sample{
			Timestamp: TimeToPromTimestamp(prevDP.TimestampNanos),
			Value:     cumulativeSum,
		})
	}

	return &prompb.TimeSeries{
		Labels:  TagsToPromLabels(tags),
		Samples: samples,
	}, nil
}

// Fall back to sequential decompression if unable to decompress concurrently.
func toPromSequentially(
	fetchResult consolidators.SeriesFetchResult,
	tagOptions models.TagOptions,
	promOptions PromOptions,
) (PromResult, error) {
	count := fetchResult.Count()
	seriesList := make([]*prompb.TimeSeries, 0, count)
	for i := 0; i < count; i++ {
		iter, tags, err := fetchResult.IterTagsAtIndex(i, tagOptions)
		if err != nil {
			return PromResult{}, err
		}

		series, err := iteratorToPromResult(iter, tags, promOptions)
		if err != nil {
			return PromResult{}, err
		}

		if len(series.GetSamples()) > 0 {
			seriesList = append(seriesList, series)
		}
	}

	return PromResult{
		PromResult: &prompb.QueryResult{
			Timeseries: seriesList,
		},
	}, nil
}

func toPromConcurrently(
	ctx context.Context,
	fetchResult consolidators.SeriesFetchResult,
	readWorkerPool xsync.PooledWorkerPool,
	tagOptions models.TagOptions,
	promOptions PromOptions,
) (PromResult, error) {
	count := fetchResult.Count()
	var (
		seriesList = make([]*prompb.TimeSeries, count)

		wg       sync.WaitGroup
		multiErr xerrors.MultiError
		mu       sync.Mutex
	)

	fastWorkerPool := readWorkerPool.FastContextCheck(100)
	for i := 0; i < count; i++ {
		i := i
		iter, tags, err := fetchResult.IterTagsAtIndex(i, tagOptions)
		if err != nil {
			return PromResult{}, err
		}

		wg.Add(1)
		available := fastWorkerPool.GoWithContext(ctx, func() {
			defer wg.Done()
			series, err := iteratorToPromResult(iter, tags, promOptions)
			if err != nil {
				mu.Lock()
				multiErr = multiErr.Add(err)
				mu.Unlock()
			}

			seriesList[i] = series
		})
		if !available {
			wg.Done()
			mu.Lock()
			multiErr = multiErr.Add(ctx.Err())
			mu.Unlock()
			break
		}
	}

	wg.Wait()
	if err := multiErr.LastError(); err != nil {
		return PromResult{}, err
	}

	// Filter out empty series inplace.
	filteredList := seriesList[:0]
	for _, s := range seriesList {
		if len(s.GetSamples()) > 0 {
			filteredList = append(filteredList, s)
		}
	}

	return PromResult{
		PromResult: &prompb.QueryResult{
			Timeseries: filteredList,
		},
	}, nil
}

func seriesIteratorsToPromResult(
	ctx context.Context,
	fetchResult consolidators.SeriesFetchResult,
	readWorkerPool xsync.PooledWorkerPool,
	tagOptions models.TagOptions,
	promOptions PromOptions,
) (PromResult, error) {
	if readWorkerPool == nil {
		return toPromSequentially(fetchResult, tagOptions, promOptions)
	}

	return toPromConcurrently(ctx, fetchResult, readWorkerPool, tagOptions, promOptions)
}

// SeriesIteratorsToPromResult converts raw series iterators directly to a
// Prometheus-compatible result.
func SeriesIteratorsToPromResult(
	ctx context.Context,
	fetchResult consolidators.SeriesFetchResult,
	readWorkerPool xsync.PooledWorkerPool,
	tagOptions models.TagOptions,
	promOptions PromOptions,
) (PromResult, error) {
	defer fetchResult.Close()
	if err := fetchResult.Verify(); err != nil {
		return PromResult{}, err
	}

	promResult, err := seriesIteratorsToPromResult(ctx, fetchResult,
		readWorkerPool, tagOptions, promOptions)
	promResult.Metadata = fetchResult.Metadata

	return promResult, err
}
