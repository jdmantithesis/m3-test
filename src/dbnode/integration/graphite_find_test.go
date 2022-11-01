//go:build integration
// +build integration

// Copyright (c) 2021 Uber Technologies, Inc.
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

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	// nolint: gci
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/m3db/m3/src/dbnode/integration/generate"
	"github.com/m3db/m3/src/dbnode/namespace"
	"github.com/m3db/m3/src/dbnode/retention"
	graphitehandler "github.com/m3db/m3/src/query/api/v1/handler/graphite"
	"github.com/m3db/m3/src/query/graphite/graphite"
	xclock "github.com/m3db/m3/src/x/clock"
	"github.com/m3db/m3/src/x/ident"
	xhttp "github.com/m3db/m3/src/x/net/http"
	xsync "github.com/m3db/m3/src/x/sync"
	xtest "github.com/m3db/m3/src/x/test"
	xtime "github.com/m3db/m3/src/x/time"
)

type testGraphiteFindDatasetSize uint

const (
	smallDatasetSize testGraphiteFindDatasetSize = iota
	largeDatasetSize
)

type testGraphiteFindOptions struct {
	checkConcurrency int
	datasetSize      testGraphiteFindDatasetSize
	findPathIndexing bool
}

func TestGraphiteFindSequential(t *testing.T) {
	// NB(rob): We need to investigate why using high concurrency (and hence
	// need to use small dataset size since otherwise verification takes
	// forever) encounters errors running on CI.
	for _, findPathIndexing := range []bool{false, true} {
		name := fmt.Sprintf("findPathIndexing=%v", findPathIndexing)
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			testGraphiteFind(t, testGraphiteFindOptions{
				checkConcurrency: 1,
				datasetSize:      smallDatasetSize,
				findPathIndexing: findPathIndexing,
			})
			t.Logf("test name=%s took %v", name, time.Since(start))
		})
	}
}

func TestGraphiteFindParallel(t *testing.T) {
	// Skip until investigation of why check concurrency encounters errors on CI.
	t.SkipNow()
	for _, findPathIndexing := range []bool{false, true} {
		name := fmt.Sprintf("findPathIndexing=%v", findPathIndexing)
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			testGraphiteFind(t, testGraphiteFindOptions{
				checkConcurrency: runtime.NumCPU(),
				datasetSize:      largeDatasetSize,
				findPathIndexing: findPathIndexing,
			})
			t.Logf("test name=%s took %v", name, time.Since(start))
		})
	}
}

func testGraphiteFind(tt *testing.T, testOpts testGraphiteFindOptions) {
	if testing.Short() {
		tt.SkipNow() // Just skip if we're doing a short run
	}

	// Make sure that parallel assertions fail test immediately
	// by using a TestingT that panics when FailNow is called.
	t := xtest.FailNowPanicsTestingT(tt)

	queryConfigYAML := fmt.Sprintf(`
listenAddress: 127.0.0.1:7201

logging:
  level: info

metrics:
  scope:
    prefix: "coordinator"
  prometheus:
    handlerPath: /metrics
    listenAddress: "127.0.0.1:0"
  sanitization: prometheus
  samplingRate: 1.0

local:
  namespaces:
    - namespace: testns
      type: unaggregated
      retention: 12h
    - namespace: testns
      type: aggregated
      retention: 12h
      resolution: 1m

carbon:
  findPathIndexingEnabled: %v
`, testOpts.findPathIndexing)

	var (
		blockSize       = 2 * time.Hour
		retentionPeriod = 6 * blockSize
		rOpts           = retention.NewOptions().
				SetRetentionPeriod(retentionPeriod).
				SetBlockSize(blockSize).
				SetBufferPast(blockSize - time.Minute).
				SetBufferFuture(blockSize - time.Minute)
		idxOpts = namespace.NewIndexOptions().
			SetEnabled(true).
			SetBlockSize(blockSize)
		nOpts = namespace.NewOptions().
			SetRetentionOptions(rOpts).
			SetIndexOptions(idxOpts)
	)
	ns, err := namespace.NewMetadata(ident.StringID("testns"), nOpts)
	require.NoError(t, err)

	opts := NewTestOptions(tt).
		SetNamespaces([]namespace.Metadata{ns})

	// Test setup.
	setup, err := NewTestSetup(tt, opts, nil)
	require.NoError(t, err)
	defer setup.Close()

	// Make sure DB node is using path indexing if set for test.
	storageOpts := setup.StorageOpts()
	indexOpts := storageOpts.IndexOptions()
	segmentBuilderOpts := indexOpts.SegmentBuilderOptions().SetGraphitePathIndexingEnabled(testOpts.findPathIndexing)
	storageOpts = storageOpts.SetIndexOptions(indexOpts.SetSegmentBuilderOptions(segmentBuilderOpts))
	setup.SetStorageOpts(storageOpts)

	log := setup.StorageOpts().InstrumentOptions().Logger().
		With(zap.String("ns", ns.ID().String()))

	require.NoError(t, setup.InitializeBootstrappers(InitializeBootstrappersOptions{
		WithFileSystem: true,
	}))

	// Write test data.
	now := setup.NowFn()()

	// Create graphite node tree for tests.
	var (
		// nolint: gosec
		randConstSeedSrc = rand.NewSource(123456789)
		// nolint: gosec
		randGen            = rand.New(randConstSeedSrc)
		rootNode           = &graphiteNode{}
		buildNodes         func(node *graphiteNode, level int)
		generateSeries     []generate.Series
		levels             int
		entriesPerLevelMin int
		entriesPerLevelMax int
	)
	switch testOpts.datasetSize {
	case smallDatasetSize:
		levels = 4
		entriesPerLevelMin = 5
		entriesPerLevelMax = 7
	case largeDatasetSize:
		// Ideally we'd always use a large dataset size, however you do need
		// high concurrency to validate this entire dataset and CI can't seem
		// to handle high concurrency without encountering errors.
		levels = 5
		entriesPerLevelMin = 6
		entriesPerLevelMax = 9
	default:
		require.FailNow(t, fmt.Sprintf("invalid test dataset size set: %d", testOpts.datasetSize))
	}

	buildNodes = func(node *graphiteNode, level int) {
		entries := entriesPerLevelMin +
			randGen.Intn(entriesPerLevelMax-entriesPerLevelMin)
		for entry := 0; entry < entries; entry++ {
			name := fmt.Sprintf("lvl%02d_entry%02d", level, entry)

			// Create a directory node and spawn more underneath.
			if nextLevel := level + 1; nextLevel <= levels {
				childDir := node.child(name+"_dir", graphiteNodeChildOptions{
					isLeaf: false,
				})
				buildNodes(childDir, nextLevel)
			}

			// Create a leaf node.
			childLeaf := node.child(name+"_leaf", graphiteNodeChildOptions{
				isLeaf: true,
			})

			// Create series to generate data for the leaf node.
			tags := make([]ident.Tag, 0, len(childLeaf.pathParts))
			for i, pathPartValue := range childLeaf.pathParts {
				tags = append(tags, ident.Tag{
					Name:  graphite.TagNameID(i),
					Value: ident.StringID(pathPartValue),
				})
			}
			series := generate.Series{
				ID:   ident.StringID(strings.Join(childLeaf.pathParts, ".")),
				Tags: ident.NewTags(tags...),
			}
			generateSeries = append(generateSeries, series)
		}
	}

	// Build tree.
	log.Info("building graphite data set series")
	buildNodes(rootNode, 0)

	// Generate and write test data.
	log.Info("generating graphite data set datapoints",
		zap.Int("seriesSize", len(generateSeries)))
	generateBlocks := make([]generate.BlockConfig, 0, len(generateSeries))
	for _, series := range generateSeries {
		generateBlocks = append(generateBlocks, []generate.BlockConfig{
			{
				IDs:       []string{series.ID.String()},
				Tags:      series.Tags,
				NumPoints: 1,
				Start:     now,
			},
		}...)
	}
	seriesMaps := generate.BlocksByStart(generateBlocks)

	// Start the server with filesystem bootstrapper.
	log.Info("starting server")
	require.NoError(t, setup.StartServer())
	log.Info("server is now up")

	var toWrite uint32
	for _, blockSeries := range seriesMaps {
		for _, series := range blockSeries {
			toWrite += uint32(len(series.Data))
		}
	}
	log.Info("writing graphite data via client",
		zap.Int("seriesMapSize", len(seriesMaps)),
		zap.Uint32("datapointsSize", toWrite),
	)

	writeWorkerPool := xsync.NewWorkerPool(1024)
	writeWorkerPool.Init()

	var (
		writeWG          sync.WaitGroup
		writeDoneCh      = make(chan struct{}, 1)
		numTotalSuccess  = atomic.NewUint32(0)
		numTotalErrors   = atomic.NewUint32(0)
		start            = time.Now()
		numSeriesEnqueue int
	)
	go func() {
		for {
			select {
			case <-writeDoneCh:
				return
			case <-time.After(5 * time.Second):
				log.Info("written datapoints progress",
					zap.Uint32("success", numTotalSuccess.Load()),
					zap.Uint32("total", toWrite),
				)
			}
		}
	}()

	session, err := setup.M3DBClient().DefaultSession()
	require.NoError(t, err)
	for _, blockSeries := range seriesMaps {
		for _, series := range blockSeries {
			numSeriesEnqueue++
			for _, value := range series.Data {
				series, value := series, value
				writeWG.Add(1)
				writeWorkerPool.Go(func() {
					defer writeWG.Done()
					err := session.WriteTagged(
						ns.ID(),
						series.ID,
						ident.NewTagsIterator(series.Tags),
						value.Timestamp,
						value.Value,
						xtime.Second,
						value.Annotation,
					)
					if err != nil {
						numTotalErrors.Inc()
						assert.NoError(t, err)
						return
					}
					numTotalSuccess.Inc()
				})
			}
		}
	}

	writeWG.Wait()
	close(writeDoneCh)

	// Check no write errors.
	require.Equal(t, int(0), int(numTotalErrors.Load()))

	log.Info("test data written",
		zap.Duration("took", time.Since(start)),
		zap.Int("written", int(numTotalSuccess.Load())))

	log.Info("data indexing verify start")

	// Wait for at least all things to be enqueued for indexing.
	expectStatPrefix := "dbindex.index-attempt+namespace=testns,"
	expectStatProcess := expectStatPrefix + "stage=process"
	expectNumIndex := numSeriesEnqueue
	indexProcess := xclock.WaitUntil(func() bool {
		counters := setup.Scope().Snapshot().Counters()
		counter, ok := counters[expectStatProcess]
		if !ok {
			return false
		}
		return int(counter.Value()) == expectNumIndex
	}, 10*time.Second)

	counters := setup.Scope().Snapshot().Counters()
	counter, ok := counters[expectStatProcess]

	var value int
	if ok {
		value = int(counter.Value())
	}
	if !indexProcess {
		logCounterValues(setup.Scope(), log)
	}
	require.True(t, indexProcess,
		fmt.Sprintf("expected to index %d but processed %d", expectNumIndex, value))

	// Stop the server.
	defer func() {
		require.NoError(t, setup.StopServer())
		log.Info("server is now down")
	}()

	// Start the query server
	log.Info("starting query server")
	require.NoError(t, setup.StartQuery(queryConfigYAML))
	log.Info("started query server", zap.String("addr", setup.QueryAddress()))

	// Stop the query server.
	defer func() {
		require.NoError(t, setup.StopQuery())
		log.Info("query server is now down")
	}()

	// Check each level of the tree can answer expected queries.
	type checkResult struct {
		leavesVerified int
	}
	type checkFailure struct {
		expected graphiteFindResults
		actual   graphiteFindResults
		failMsg  string
	}
	var (
		verifyFindQueries         func(node *graphiteNode, level int) (checkResult, *checkFailure, error)
		parallelVerifyFindQueries func(node *graphiteNode, level int)
		checkedSeriesAbort        = atomic.NewBool(false)
		numSeriesChecking         = uint64(len(generateSeries))
		checkedSeriesLogEvery     = numSeriesChecking / 10
		checkedSeries             = atomic.NewUint64(0)
		checkedSeriesLog          = atomic.NewUint64(0)
		// Use custom http client for higher number of max idle conns.
		httpClient = xhttp.NewHTTPClient(xhttp.DefaultHTTPClientOptions())
		wg         sync.WaitGroup
		workerPool = xsync.NewWorkerPool(testOpts.checkConcurrency)
	)
	workerPool.Init()
	parallelVerifyFindQueries = func(node *graphiteNode, level int) {
		// Verify this node at level.
		wg.Add(1)
		workerPool.Go(func() {
			defer wg.Done()

			if checkedSeriesAbort.Load() {
				// Do not execute if aborted.
				return
			}

			result, failure, err := verifyFindQueries(node, level)
			if failure == nil && err == nil {
				// Account for series checked (for progress report).
				checkedSeries.Add(uint64(result.leavesVerified))
				return
			}

			// Bail parallel execution (failed require/assert won't stop execution).
			if checkedSeriesAbort.CAS(false, true) {
				switch {
				case failure != nil:
					// Assert an error result and log once.
					assert.Equal(t, failure.expected, failure.actual, failure.failMsg)
					log.Error("aborting checks due to mismatch")
				case err != nil:
					assert.NoError(t, err)
					log.Error("aborting checks due to error")
				default:
					require.FailNow(t, "unknown error condition")
					log.Error("aborting checks due to unknown condition")
				}
			}
		})

		// Verify children of children.
		for _, child := range node.children {
			parallelVerifyFindQueries(child, level+1)
		}
	}
	graphiteFind := func(query string) (graphiteFindResults, error) {
		params := make(url.Values)
		params.Set("query", query)

		url := fmt.Sprintf("http://%s%s?%s", setup.QueryAddress(),
			graphitehandler.FindURL, params.Encode())

		req, err := http.NewRequestWithContext(context.Background(),
			http.MethodGet, url, nil)
		if err != nil {
			return graphiteFindResults{}, err
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return graphiteFindResults{}, err
		}
		if res.StatusCode != http.StatusOK {
			return graphiteFindResults{}, fmt.Errorf("bad response code: expected=%d, actual=%d",
				http.StatusOK, res.StatusCode)
		}

		defer res.Body.Close()

		// Compare results.
		var actual graphiteFindResults
		if err := json.NewDecoder(res.Body).Decode(&actual); err != nil {
			return graphiteFindResults{}, err
		}

		return actual, nil
	}
	verifyFindQueries = func(node *graphiteNode, level int) (checkResult, *checkFailure, error) {
		var r checkResult

		// Write progress report if progress made.
		checked := checkedSeries.Load()
		nextLog := checked - (checked % checkedSeriesLogEvery)
		if lastLog := checkedSeriesLog.Swap(nextLog); lastLog < nextLog {
			log.Info("checked series progressing", zap.Int("checked", int(checked)))
		}

		// Verify at depth.
		numPathParts := len(node.pathParts)
		queryPathParts := make([]string, 0, 1+numPathParts)
		if numPathParts > 0 {
			queryPathParts = append(queryPathParts, node.pathParts...)
		}
		queryPathParts = append(queryPathParts, "*")
		query := strings.Join(queryPathParts, ".")

		actual, err := graphiteFind(query)
		if err != nil {
			return r, nil, err
		}

		expected := make(graphiteFindResults, 0, len(node.children))
		for _, child := range node.children {
			leaf := 0
			if child.isLeaf {
				leaf = 1
				r.leavesVerified++
			}

			expected = append(expected, graphiteFindResult{
				ID:   child.ID(),
				Text: child.name,
				Leaf: leaf,
			})
		}

		sortGraphiteFindResults(actual)
		sortGraphiteFindResults(expected)

		if !reflect.DeepEqual(expected, actual) {
			failMsg := fmt.Sprintf("invalid results: level=%d, parts=%d, query=%s",
				level, len(node.pathParts), query)
			failMsg += fmt.Sprintf("\n\ndiff:\n%s\n\n",
				xtest.Diff(xtest.MustPrettyJSONObject(t, expected),
					xtest.MustPrettyJSONObject(t, actual)))
			return r, &checkFailure{
				expected: expected,
				actual:   actual,
				failMsg:  failMsg,
			}, nil
		}

		return r, nil, nil
	}

	// Issue sanity checks.
	// #1 exact match dir
	results, err := graphiteFind("lvl00_entry00_dir.lvl01_entry00_dir")
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "lvl00_entry00_dir.lvl01_entry00_dir", results[0].ID)
	require.Equal(t, "lvl01_entry00_dir", results[0].Text)
	require.Equal(t, 0, results[0].Leaf)

	// #2 exact match leaf
	results, err = graphiteFind("lvl00_entry00_dir.lvl01_entry00_dir.lvl02_entry00_leaf")
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "lvl00_entry00_dir.lvl01_entry00_dir.lvl02_entry00_leaf", results[0].ID)
	require.Equal(t, "lvl02_entry00_leaf", results[0].Text)
	require.Equal(t, 1, results[0].Leaf)

	// #3 wildcard match *leaf*
	results, err = graphiteFind("lvl00_entry00_dir.lvl01_entry00_dir.*lvl02_entry00_leaf*")
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "lvl00_entry00_dir.lvl01_entry00_dir.lvl02_entry00_leaf", results[0].ID)
	require.Equal(t, "lvl02_entry00_leaf", results[0].Text)
	require.Equal(t, 1, results[0].Leaf)

	// Check all top level entries and recurse.
	log.Info("checking series",
		zap.Int("checkConcurrency", testOpts.checkConcurrency),
		zap.Uint64("numSeriesChecking", numSeriesChecking))
	parallelVerifyFindQueries(rootNode, 0)

	// Wait for execution.
	wg.Wait()

	// Allow for debugging by issuing queries, etc.
	if DebugTest() {
		log.Info("debug test set, pausing for investigate")
		<-make(chan struct{})
	}
}

type graphiteFindResults []graphiteFindResult

type graphiteFindResult struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Leaf int    `json:"leaf"`
}

func sortGraphiteFindResults(r graphiteFindResults) {
	sort.Slice(r, func(i, j int) bool {
		if r[i].Leaf != r[j].Leaf {
			return r[i].Leaf < r[j].Leaf
		}
		return r[i].Text < r[j].Text
	})
}

type graphiteNode struct {
	name      string
	pathParts []string
	isLeaf    bool
	children  []*graphiteNode
}

func (n graphiteNode) ID() string {
	return strings.Join(n.pathParts, ".")
}

type graphiteNodeChildOptions struct {
	isLeaf bool
}

func (n *graphiteNode) child(
	name string,
	opts graphiteNodeChildOptions,
) *graphiteNode {
	pathParts := append(make([]string, 0, 1+len(n.pathParts)), n.pathParts...)
	pathParts = append(pathParts, name)

	child := &graphiteNode{
		name:      name,
		pathParts: pathParts,
		isLeaf:    opts.isLeaf,
	}

	n.children = append(n.children, child)

	return child
}
