// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/m3db/m3/src/query/graphite/storage (interfaces: Storage)

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

// Package storage is a generated GoMock package.
package storage

import (
	"context"
	"reflect"

	context0 "github.com/m3db/m3/src/query/graphite/context"
	storage0 "github.com/m3db/m3/src/query/storage"
	"github.com/m3db/m3/src/query/storage/m3/consolidators"

	"github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// CompleteTags mocks base method.
func (m *MockStorage) CompleteTags(arg0 context.Context, arg1 *storage0.CompleteTagsQuery, arg2 *storage0.FetchOptions) (*consolidators.CompleteTagsResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompleteTags", arg0, arg1, arg2)
	ret0, _ := ret[0].(*consolidators.CompleteTagsResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CompleteTags indicates an expected call of CompleteTags.
func (mr *MockStorageMockRecorder) CompleteTags(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompleteTags", reflect.TypeOf((*MockStorage)(nil).CompleteTags), arg0, arg1, arg2)
}

// FetchByQuery mocks base method.
func (m *MockStorage) FetchByQuery(arg0 context0.Context, arg1 string, arg2 FetchOptions) (*FetchResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchByQuery", arg0, arg1, arg2)
	ret0, _ := ret[0].(*FetchResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchByQuery indicates an expected call of FetchByQuery.
func (mr *MockStorageMockRecorder) FetchByQuery(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchByQuery", reflect.TypeOf((*MockStorage)(nil).FetchByQuery), arg0, arg1, arg2)
}
