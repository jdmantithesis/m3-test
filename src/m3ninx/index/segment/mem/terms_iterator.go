// Copyright (c) 2018 Uber Technologies, Inc.
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

package mem

import (
	"fmt"

	sgmt "github.com/m3db/m3/src/m3ninx/index/segment"
	"github.com/m3db/m3/src/m3ninx/postings"
)

var errNotReset = fmt.Errorf("terms iterator not reset")

type termsLookup interface {
	Get(key []byte) (postings.List, bool)
}

type termsIter struct {
	seg *memSegment

	err  error
	done bool

	currentIdx      int
	current         []byte
	currentPostings postings.List
	backingSlice    [][]byte
	backingPostings termsLookup
	opts            Options
}

var _ sgmt.ReuseableTermsIterator = &termsIter{}

func newTermsIter(seg *memSegment) *termsIter {
	termsIter := termsIterInitState(seg)
	return &termsIter
}

func termsIterInitState(seg *memSegment) termsIter {
	return termsIter{seg: seg, err: errNotReset}
}

func (b *termsIter) ResetField(field []byte) error {
	terms, err := b.seg.termsDict.terms(field)
	if err != nil {
		return err
	}

	keys := terms.Keys()
	b.reset(terms, keys)
	return nil
}

func (b *termsIter) ResetFieldWithNumTerms(field []byte, numTerms int) error {
	terms, err := b.seg.termsDict.terms(field)
	if err != nil {
		return err
	}
	keys := terms.Keys()
	if len(keys) != numTerms {
		return fmt.Errorf("expected %d terms, got %d", numTerms, len(keys))
	}
	b.reset(terms, keys)
	return nil
}

func (b *termsIter) reset(
	terms *concurrentPostingsMap,
	keys [][]byte,
) {
	sortSliceOfByteSlices(keys)
	b.err = nil
	b.done = false
	b.currentIdx = -1
	b.backingSlice = keys
	b.backingPostings = terms
	b.opts = b.seg.termsDict.opts
}

func (b *termsIter) AllTermsLength() int {
	return len(b.backingSlice)
}

func (b *termsIter) Next() bool {
	if b.done || b.err != nil {
		return false
	}
	b.currentIdx++
	if b.currentIdx >= len(b.backingSlice) {
		b.done = true
		return false
	}
	var ok bool
	b.current = b.backingSlice[b.currentIdx]
	b.currentPostings, ok = b.backingPostings.Get(b.current)
	if !ok {
		b.err = fmt.Errorf("term not found during iteration: %s", b.current)
		return false
	}
	return true
}

func (b *termsIter) Current() (term []byte, postings postings.List) {
	return b.current, b.currentPostings
}

func (b *termsIter) Err() error {
	return b.err
}

func (b *termsIter) Len() int {
	return len(b.backingSlice)
}

func (b *termsIter) Close() error {
	if cap(b.backingSlice) > 0 {
		b.opts.BytesSliceArrayPool().Put(b.backingSlice)
	}
	// Reset to zero state.
	*b = termsIterInitState(b.seg)
	return nil
}
