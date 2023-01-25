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

package native

import (
	"fmt"
	"math"
	"testing"

	"github.com/m3db/m3/src/query/graphite/common"
	"github.com/m3db/m3/src/query/graphite/lexer"
	"github.com/m3db/m3/src/query/graphite/storage"
	graphitetest "github.com/m3db/m3/src/query/graphite/testing"
	"github.com/m3db/m3/src/query/graphite/ts"
	"github.com/m3db/m3/src/x/instrument"
	xtest "github.com/m3db/m3/src/x/test"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
)

type testCompile struct {
	input   string
	result  interface{}
	series  func(*common.Context) []*ts.Series
	asserts func(require.TestingT, tally.TestScope, Expression)
}

func hello(ctx *common.Context) (string, error)         { return "hello", nil }
func noArgs(ctx *common.Context) (ts.SeriesList, error) { return ts.SeriesList{}, nil }
func defaultArgs(ctx *common.Context, b bool, f1, f2 float64, s string) (ts.SeriesList, error) {
	return ts.SeriesList{}, nil
}

func TestCompile1(t *testing.T) {
	var (
		sortByName        = findFunction("sortByName")
		noArgs            = findFunction("noArgs")
		aliasByNode       = findFunction("aliasByNode")
		summarize         = findFunction("summarize")
		defaultArgs       = findFunction("defaultArgs")
		sumSeries         = findFunction("sumSeries")
		asPercent         = findFunction("asPercent")
		scale             = findFunction("scale")
		logarithm         = findFunction("logarithm")
		removeEmptySeries = findFunction("removeEmptySeries")
		filterSeries      = findFunction("filterSeries")
	)

	tests := []testCompile{
		{input: "", result: noopExpression{}},
		{input: "foobar", result: newFetchExpression("foobar")},
		{input: "foo.bar.{a,b,c}.baz-*.stat[0-9]", result: newFetchExpression("foo.bar.{a,b,c}.baz-*.stat[0-9]")},
		{input: "noArgs()", result: &funcExpression{&functionCall{f: noArgs}}},
		{input: "sortByName(foo.bar.zed)", result: &funcExpression{
			&functionCall{
				f: sortByName,
				in: []funcArg{
					newFetchExpression("foo.bar.zed"),
				},
			},
		}},
		{input: "aliasByNode(foo.bar4.*.metrics.written, 2, 4)", result: &funcExpression{
			&functionCall{
				f: aliasByNode,
				in: []funcArg{
					newFetchExpression("foo.bar4.*.metrics.written"),
					newIntConst(2),
					newIntConst(4),
				},
			},
		}},
		{input: "summarize(foo.bar.baz.quux, \"1h\", \"max\", TRUE)", result: &funcExpression{
			&functionCall{
				f: summarize,
				in: []funcArg{
					newFetchExpression("foo.bar.baz.quux"),
					newStringConst("1h"),
					newStringConst("max"),
					newBoolConst(true),
				},
			},
		}},
		{input: "summarize(foo.bar.baz.quuz, \"1h\")", result: &funcExpression{
			&functionCall{
				f: summarize,
				in: []funcArg{
					newFetchExpression("foo.bar.baz.quuz"),
					newStringConst("1h"),
					newStringConst(""),
					newBoolConst(false),
				},
			},
		}},
		{input: "defaultArgs(true)", result: &funcExpression{
			&functionCall{
				f: defaultArgs,
				in: []funcArg{
					newBoolConst(true),          // non-default value
					newFloat64Const(math.NaN()), // default value
					newFloat64Const(100),        // default value
					newStringConst("foobar"),    // default value
				},
			},
		}},
		{input: "sortByName(aliasByNode(foo.bar72.*.metrics.written,2,4,6))", result: &funcExpression{
			&functionCall{
				f: sortByName,
				in: []funcArg{
					&functionCall{
						f: aliasByNode,
						in: []funcArg{
							newFetchExpression("foo.bar72.*.metrics.written"),
							newIntConst(2),
							newIntConst(4),
							newIntConst(6),
						},
					},
				},
			},
		}},
		{input: "sumSeries(foo.bar.baz.quux, foo.bar72.*.metrics.written)", result: &funcExpression{
			&functionCall{
				f: sumSeries,
				in: []funcArg{
					newFetchExpression("foo.bar.baz.quux"),
					newFetchExpression("foo.bar72.*.metrics.written"),
				},
			},
		}},
		{input: "asPercent(foo.bar72.*.metrics.written, foo.bar.baz.quux)", result: &funcExpression{
			&functionCall{
				f: asPercent,
				in: []funcArg{
					newFetchExpression("foo.bar72.*.metrics.written"),
					newFetchExpression("foo.bar.baz.quux"),
				},
			},
		}},
		{input: "asPercent(foo.bar72.*.metrics.written, sumSeries(foo.bar.baz.quux))", result: &funcExpression{
			&functionCall{
				f: asPercent,
				in: []funcArg{
					newFetchExpression("foo.bar72.*.metrics.written"),
					&functionCall{
						f: sumSeries,
						in: []funcArg{
							newFetchExpression("foo.bar.baz.quux"),
						},
					},
				},
			},
		}},
		{input: "asPercent(foo.bar72.*.metrics.written, 100)", result: &funcExpression{
			&functionCall{
				f: asPercent,
				in: []funcArg{
					newFetchExpression("foo.bar72.*.metrics.written"),
					newIntConst(100),
				},
			},
		}},
		{input: "asPercent(foo.bar72.*.metrics.written)", result: &funcExpression{
			&functionCall{
				f: asPercent,
				in: []funcArg{
					newFetchExpression("foo.bar72.*.metrics.written"),
					newConstArg([]*ts.Series(nil)),
				},
			},
		}},
		{input: "asPercent(foo.bar72.*.metrics.written, total=sumSeries(foo.bar.baz.quux))", result: &funcExpression{
			&functionCall{
				f: asPercent,
				in: []funcArg{
					newFetchExpression("foo.bar72.*.metrics.written"),
					&functionCall{
						f: sumSeries,
						in: []funcArg{
							newFetchExpression("foo.bar.baz.quux"),
						},
					},
				},
			},
		}},
		{input: "asPercent(foo.bar72.*.metrics.written, total=100)", result: &funcExpression{
			&functionCall{
				f: asPercent,
				in: []funcArg{
					newFetchExpression("foo.bar72.*.metrics.written"),
					newIntConst(100),
				},
			},
		}},
		{input: "scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1e+3)", result: &funcExpression{
			&functionCall{
				f: scale,
				in: []funcArg{
					newFetchExpression("servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*"),
					newFloat64Const(1000),
				},
			},
		}},
		{input: "scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1e-3)", result: &funcExpression{
			&functionCall{
				f: scale,
				in: []funcArg{
					newFetchExpression("servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*"),
					newFloat64Const(0.001),
				},
			},
		}},
		{input: "scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1e3)", result: &funcExpression{
			&functionCall{
				f: scale,
				in: []funcArg{
					newFetchExpression("servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*"),
					newFloat64Const(1000),
				},
			},
		}},
		{input: "scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.1e3)", result: &funcExpression{
			&functionCall{
				f: scale,
				in: []funcArg{
					newFetchExpression("servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*"),
					newFloat64Const(1100),
				},
			},
		}},
		{input: "scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.1e+3)", result: &funcExpression{
			&functionCall{
				f: scale,
				in: []funcArg{
					newFetchExpression("servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*"),
					newFloat64Const(1100),
				},
			},
		}},
		{input: "scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.2e-3)", result: &funcExpression{
			&functionCall{
				f: scale,
				in: []funcArg{
					newFetchExpression("servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*"),
					newFloat64Const(0.0012),
				},
			},
		}},
		{input: "scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, .1e+3)", result: &funcExpression{
			&functionCall{
				f: scale,
				in: []funcArg{
					newFetchExpression("servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*"),
					newFloat64Const(100),
				},
			},
		}},
		{input: "scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 2.e+3)", result: &funcExpression{
			&functionCall{
				f: scale,
				in: []funcArg{
					newFetchExpression("servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*"),
					newFloat64Const(2000),
				},
			},
		}},
		{input: "logarithm(a.b.c)", result: &funcExpression{
			&functionCall{
				f: logarithm,
				in: []funcArg{
					newFetchExpression("a.b.c"),
					newFloat64Const(10),
				},
			},
		}},
		{input: "removeEmptySeries(a.b.c)", result: &funcExpression{
			&functionCall{
				f: removeEmptySeries,
				in: []funcArg{
					newFetchExpression("a.b.c"),
					newFloat64Const(0),
				},
			},
		}},
		{input: "filterSeries(a.b.c, 'max', '>', 1000)", result: &funcExpression{
			&functionCall{
				f: filterSeries,
				in: []funcArg{
					newFetchExpression("a.b.c"),
					newStringConst("max"),
					newStringConst(">"),
					newFloat64Const(1000),
				},
			},
		}},
		{
			input: "movingAverage(foo.bar.baz, '10min', 0)",
			result: &funcExpression{
				&functionCall{
					f: findFunction("movingAverage"),
					in: []funcArg{
						newFetchExpression("foo.bar.baz"),
						newStringConst("10min"),
						newFloat64Const(0),
					},
				},
			},
			asserts: func(t require.TestingT, s tally.TestScope, e Expression) {
				c := s.Snapshot().Counters()
				v, ok := c["function-call.optimized-shift-result-optimized+function=movingAverage"]
				require.True(t, ok)
				assert.Equal(t, int64(1), v.Value())
			},
		},
		{
			// Test whether multi-fetch optimization is working.
			input: "sumSeries(foo.bar, foo.baz)",
			series: func(ctx *common.Context) []*ts.Series {
				return []*ts.Series{
					ts.NewSeries(ctx, "foo.bar", ctx.StartTime,
						ts.NewConstantValues(ctx, 42, 10, 60000)),
					ts.NewSeries(ctx, "foo.baz", ctx.StartTime,
						ts.NewConstantValues(ctx, 42, 10, 60000)),
				}
			},
			asserts: func(t require.TestingT, s tally.TestScope, e Expression) {
				// Following is for debugging, the fetches should be of form:
				// sumSeries(fetch({foo.bar,foo.baz}))
				// Not:
				// sumSeries(fetch(foo.bar),fetch(foo.baz))
				fmt.Println("parsed", e.String())

				c := s.Snapshot().Counters()

				v, ok := c["graphite-compiler.multi-fetch-optimized+"]
				require.True(t, ok)
				assert.Equal(t, int64(1), v.Value())

				v, ok = c["fetch-expression.execute-fetch+"]
				require.True(t, ok)
				assert.Equal(t, int64(1), v.Value())
			},
		},
		{
			// Test whether multi-fetch optimization with patterns is working.
			input: "sumSeries(foo.bar.*, foo.baz.{qux,qaz})",
			series: func(ctx *common.Context) []*ts.Series {
				return []*ts.Series{
					ts.NewSeries(ctx, "foo.bar", ctx.StartTime,
						ts.NewConstantValues(ctx, 42, 10, 60000)),
					ts.NewSeries(ctx, "foo.baz", ctx.StartTime,
						ts.NewConstantValues(ctx, 42, 10, 60000)),
				}
			},
			asserts: func(t require.TestingT, s tally.TestScope, e Expression) {
				// Following is for debugging, the fetches should be of form:
				// sumSeries(fetch({foo.bar,foo.baz}))
				// Not:
				// sumSeries(fetch(foo.bar),fetch(foo.baz))
				fmt.Println("parsed", e.String())

				c := s.Snapshot().Counters()

				v, ok := c["graphite-compiler.multi-fetch-optimized+"]
				require.True(t, ok)
				assert.Equal(t, int64(1), v.Value())

				v, ok = c["fetch-expression.execute-fetch+"]
				require.True(t, ok)
				assert.Equal(t, int64(1), v.Value())
			},
		},
		{
			// Test that multi-fetch optimization is disabled for diffSeries
			// which relies on order of metrics returned.
			input: "diffSeries(foo.bar, foo.baz)",
			series: func(ctx *common.Context) []*ts.Series {
				return []*ts.Series{
					ts.NewSeries(ctx, "foo.bar", ctx.StartTime,
						ts.NewConstantValues(ctx, 42, 10, 60000)),
					ts.NewSeries(ctx, "foo.baz", ctx.StartTime,
						ts.NewConstantValues(ctx, 42, 10, 60000)),
				}
			},
			asserts: func(t require.TestingT, s tally.TestScope, e Expression) {
				// Following is for debugging, the fetches should be of form:
				// diffSeries(fetch(foo.bar),fetch(foo.baz))
				// Not:
				// diffSeries(fetch({foo.bar,foo.baz}))
				fmt.Println("parsed", e.String())

				c := s.Snapshot().Counters()

				v, ok := c["fetch-expression.execute-fetch+"]
				require.True(t, ok)
				assert.Equal(t, int64(2), v.Value())
			},
		},
		{
			// Test whether complex deeply nested multi-fetch optimization is working.
			input: "maxSeries(divideSeries(movingAverage(sumSeries(foo.count.QAZ,foo.count.FOO,foo.count.BAR,foo.count.BAZ,foo.count.QUX,bar.count.QAZ,bar.count.FOO,bar.count.BAR,bar.count.BAZ,bar.count.QUX), 60), movingAverage(sumSeries(baz.count.QAZ,baz.count.FOO,baz.count.BAR,baz.count.BAZ,baz.count.QUX,qux.count.QAZ,qux.count.FOO,qux.count.BAR,qux.count.BAZ,qux.count.QUX), 60)),divideSeries(movingAverage(sumSeries(foo.count.QAZ,foo.count.FOO,foo.count.BAR,foo.count.BAZ,foo.count.QUX,bar.count.QAZ,bar.count.FOO,bar.count.BAR,bar.count.BAZ,bar.count.QUX), 15), movingAverage(sumSeries(baz.count.QAZ,baz.count.FOO,baz.count.BAR,baz.count.BAZ,baz.count.QUX,qux.count.QAZ,qux.count.FOO,qux.count.BAR,qux.count.BAZ,qux.count.QUX), 15)))",
			series: func(ctx *common.Context) []*ts.Series {
				return []*ts.Series{
					ts.NewSeries(ctx, "foo.count:QAZ", ctx.StartTime,
						ts.NewConstantValues(ctx, 42, 10, 60000)),
				}
			},
			asserts: func(t require.TestingT, s tally.TestScope, e Expression) {
				// Following is for debugging, the fetches should be of form:
				// sumSeries(fetch({foo.bar,foo.baz}))
				// Not:
				// sumSeries(fetch(foo.bar),fetch(foo.baz))
				fmt.Println("parsed", e.String())

				c := s.Snapshot().Counters()

				v, ok := c["graphite-compiler.multi-fetch-optimized+"]
				require.True(t, ok)
				assert.Equal(t, int64(4), v.Value())

				v, ok = c["fetch-expression.execute-fetch+"]
				require.True(t, ok)
				assert.Equal(t, int64(8), v.Value())
			},
		},
	}

	ctrl := xtest.NewController(t)
	defer ctrl.Finish()

	for _, test := range tests {
		ctx := common.NewTestContext()
		store := storage.NewMockStorage(ctrl)

		result := &storage.FetchResult{}
		if test.series != nil {
			result.SeriesList = test.series(ctx)
		}

		store.EXPECT().FetchByQuery(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(result, nil).
			AnyTimes()

		ctx.Engine = NewEngine(store, CompileOptions{})

		scope := tally.NewTestScope("", nil)

		expr, err := Compile(test.input, CompileOptions{
			InstrumentOptions: instrument.NewOptions().SetMetricsScope(scope),
		})
		require.NoError(t, err, "error compiling: expression='%s', error='%v'", test.input, err)
		require.NotNil(t, expr)

		// Assert the expression tree if checking result.
		if test.result != nil {
			assertExprTree(t, test.result, expr, fmt.Sprintf("invalid result for %s: %v vs %v",
				test.input, test.result, expr))
		}

		// Ensure that the function can execute.
		_, err = expr.Execute(ctx)
		require.NoError(t, err)

		// Run assertions.
		if test.asserts != nil {
			test.asserts(t, scope, expr)
		}
	}
}

type testCompilerError struct {
	input string
	err   string
}

func TestCompileErrors(t *testing.T) {
	tests := []testCompilerError{
		{"hello()", "top-level functions must return timeseries data"},
		{"foobar(", "invalid expression 'foobar(': could not find function named foobar"},
		{"foobar()", "invalid expression 'foobar()': could not find function named foobar"},
		{"sortByName(foo.*.zed)junk", "invalid expression 'sortByName(foo.*.zed)junk': " +
			"extra data junk"},
		{
			"aliasByNode(",
			"invalid expression 'aliasByNode(': unexpected eof while parsing aliasByNode",
		},
		{
			"unknownFunc()",
			"invalid expression 'unknownFunc()': could not find function named unknownFunc",
		},
		{
			"aliasByNode(10)",
			"invalid expression 'aliasByNode(10)': invalid function call aliasByNode," +
				" arg 0: expected a singlePathSpec, received a float64 '10'",
		},
		{
			"sortByName(hello())",
			"invalid expression 'sortByName(hello())': invalid function call " +
				"sortByName, arg 0: expected a singlePathSpec, received a functionCall 'hello()'",
		},
		{
			"aliasByNode()",
			"invalid expression 'aliasByNode()': invalid number of arguments for aliasByNode;" +
				" expected at least 2, received 0",
		},
		{
			"aliasByNode(foo.*.zed, 2, false)",
			"invalid expression 'aliasByNode(foo.*.zed, 2, false)': invalid function call " +
				"aliasByNode, arg 2: expected a int, received a bool 'false'",
		},
		{
			"aliasByNode(foo.*.bar,",
			"invalid expression 'aliasByNode(foo.*.bar,': unexpected eof while" +
				" parsing aliasByNode",
		},
		{
			"aliasByNode(foo.*.bar,)",
			"invalid expression 'aliasByNode(foo.*.bar,)': invalid function call" +
				" aliasByNode, arg 1: invalid expression 'aliasByNode(foo.*.bar,)': ) not valid",
		},
		// TODO(jayp): Not providing all required parameters in a function with default parameters
		// leads to an error message that states that a greater than required number of expected
		// arguments. We could do better, but punting for now.
		{
			"summarize(foo.bar.baz.quux)",
			"invalid expression 'summarize(foo.bar.baz.quux)':" +
				" invalid number of arguments for summarize; expected 4, received 1",
		},
		{
			"sumSeries(foo.bar.baz.quux,)",
			"invalid expression 'sumSeries(foo.bar.baz.quux,)': invalid function call sumSeries, " +
				"arg 1: invalid expression 'sumSeries(foo.bar.baz.quux,)': ) not valid",
		},
		{
			"asPercent(foo.bar72.*.metrics.written, total",
			"invalid expression 'asPercent(foo.bar72.*.metrics.written, total': " +
				"invalid function call asPercent, " +
				"arg 1: invalid expression 'asPercent(foo.bar72.*.metrics.written, total': " +
				"unexpected eof, total should be followed by = or (",
		},
		{
			"asPercent(foo.bar72.*.metrics.written, total=",
			"invalid expression 'asPercent(foo.bar72.*.metrics.written, total=': " +
				"invalid function call asPercent, " +
				"arg 1: invalid expression 'asPercent(foo.bar72.*.metrics.written, total=': " +
				"unexpected eof, named argument total should be followed by its value",
		},
		{
			"asPercent(foo.bar72.*.metrics.written, total=randomStuff",
			"invalid expression 'asPercent(foo.bar72.*.metrics.written, total=randomStuff': " +
				"invalid function call asPercent, " +
				"arg 1: invalid expression 'asPercent(foo.bar72.*.metrics.written, total=randomStuff': " +
				"unexpected eof, randomStuff should be followed by = or (",
		},
		{
			"asPercent(foo.bar72.*.metrics.written, total=sumSeries(",
			"invalid expression 'asPercent(foo.bar72.*.metrics.written, total=sumSeries(': " +
				"invalid function call asPercent, " +
				"arg 1: invalid expression 'asPercent(foo.bar72.*.metrics.written, total=sumSeries(': " +
				"unexpected eof while parsing sumSeries",
		},
		{
			"scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.e)",
			"invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.e)': " +
				"invalid function call scale, " +
				"arg 1: invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.e)': " +
				"expected one of 0123456789, found ) not valid",
		},
		{
			"scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, .1e)",
			"invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, .1e)': " +
				"invalid function call scale, " +
				"arg 1: invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, .1e)': " +
				"expected one of 0123456789, found ) not valid",
		},
		{
			"scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, .e)",
			"invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, .e)': " +
				"invalid function call scale, " +
				"arg 1: invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, .e)': " +
				"expected one of 0123456789, found e not valid",
		},
		{
			"scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, e)",
			"invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, e)': " +
				"invalid function call scale, " +
				"arg 1: expected a float64, received a fetchExpression 'fetch(e)'",
		},
		{
			"scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.2ee)",
			"invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.2ee)': " +
				"invalid function call scale, " +
				"arg 1: invalid expression 'scale(servers.foobar*-qaz.quail.qux-qaz-qab.cpu.*, 1.2ee)': " +
				"expected one of 0123456789, found e not valid",
		},
	}

	for _, test := range tests {
		expr, err := Compile(test.input, CompileOptions{})
		require.NotNil(t, err, "no error for %s", test.input)
		assert.Equal(t, test.err, err.Error(), "wrong error for %s", test.input)
		assert.Nil(t, expr, "non-nil expression for %s", test.input)
	}
}

func assertExprTree(t *testing.T, expected interface{}, actual interface{}, msg string) {
	switch e := expected.(type) {
	case *functionCall:
		a, ok := actual.(*functionCall)
		require.True(t, ok, msg)
		require.Equal(t, e.f.name, a.f.name, msg)
		require.Equal(t, len(e.f.in), len(a.f.in), msg)
		for i := range e.in {
			assertExprTree(t, e.in[i], a.in[i], msg)
		}
	case noopExpression:
		_, ok := actual.(noopExpression)
		require.True(t, ok, msg)
	case *funcExpression:
		a, ok := actual.(*funcExpression)
		require.True(t, ok, msg)
		assertExprTree(t, e.call, a.call, msg)
	case *fetchExpression:
		a, ok := actual.(*fetchExpression)
		require.True(t, ok, msg)
		assert.Equal(t, e.pathArg.path, a.pathArg.path, msg)
	case constFuncArg:
		a, ok := actual.(constFuncArg)
		require.True(t, ok, msg)
		if !a.value.IsValid() {
			// Explicit nil.
			require.True(t, e.value.IsZero())
		} else {
			graphitetest.Equalish(t, e.value.Interface(), a.value.Interface(), msg)
		}
	default:
		assert.Equal(t, expected, actual, msg)
	}
}

func TestExtractFetchExpressions(t *testing.T) {
	tests := []struct {
		expr    string
		targets []string
	}{
		{"summarize(groupByNode(nonNegativeDerivative(foo.qaz.gauges.bar.baz.qux.foobar.*.quz.quail.count), 8, 'sum'), '10min', 'avg', true)", []string{
			"foo.qaz.gauges.bar.baz.qux.foobar.*.quz.quail.count",
		}},
		{"asPercent(foo.bar72.*.metrics.written, total=sumSeries(foo.bar.baz.quux))", []string{
			"foo.bar72.*.metrics.written", "foo.bar.baz.quux",
		}},
		{"foo.bar.{a,b,c}.baz-*.stat[0-9]", []string{
			"foo.bar.{a,b,c}.baz-*.stat[0-9]",
		}},
	}

	for _, test := range tests {
		targets, err := ExtractFetchExpressions(test.expr)
		require.NoError(t, err)
		assert.Equal(t, test.targets, targets, test.expr)
	}
}

func TestTokenLookforward(t *testing.T) {
	tokenVals := []string{"a", "b", "c"}
	tokens := make(chan *lexer.Token)
	go func() {
		for _, v := range tokenVals {
			tokens <- lexer.MustMakeToken(v)
		}

		close(tokens)
	}()

	lookforward := newTokenLookforward(tokens)
	token := lookforward.get()
	assert.Equal(t, "a", token.Value())

	// assert that peek does not iterate token.
	token, found := lookforward.peek()
	assert.True(t, found)
	assert.Equal(t, "b", token.Value())
	token, found = lookforward.peek()
	assert.True(t, found)
	assert.Equal(t, "b", token.Value())

	// assert that next get after peek will iterate forward.
	token = lookforward.get()
	assert.Equal(t, "b", token.Value())
	token = lookforward.get()
	assert.Equal(t, "c", token.Value())

	// assert peek is empty once channel is closed.
	_, found = lookforward.peek()
	assert.False(t, found)
}

func init() {
	MustRegisterFunction(noArgs)
	MustRegisterFunction(hello)
	MustRegisterFunction(defaultArgs).WithDefaultParams(map[uint8]interface{}{
		1: false,
		2: math.NaN(),
		3: 100.0,
		4: "foobar",
	})
}
