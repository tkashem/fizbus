// Automatically generated by MockGen. DO NOT EDIT!
// Source: golang.org/x/net/context (interfaces: Context)

package fizbus

import (
	gomock "github.com/golang/mock/gomock"
	time "time"
)

// Mock of Context interface
type MockContext struct {
	ctrl     *gomock.Controller
	recorder *_MockContextRecorder
}

// Recorder for MockContext (not exported)
type _MockContextRecorder struct {
	mock *MockContext
}

func NewMockContext(ctrl *gomock.Controller) *MockContext {
	mock := &MockContext{ctrl: ctrl}
	mock.recorder = &_MockContextRecorder{mock}
	return mock
}

func (_m *MockContext) EXPECT() *_MockContextRecorder {
	return _m.recorder
}

func (_m *MockContext) Deadline() (time.Time, bool) {
	ret := _m.ctrl.Call(_m, "Deadline")
	ret0, _ := ret[0].(time.Time)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

func (_mr *_MockContextRecorder) Deadline() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Deadline")
}

func (_m *MockContext) Done() <-chan struct{} {
	ret := _m.ctrl.Call(_m, "Done")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

func (_mr *_MockContextRecorder) Done() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Done")
}

func (_m *MockContext) Err() error {
	ret := _m.ctrl.Call(_m, "Err")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockContextRecorder) Err() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Err")
}

func (_m *MockContext) Value(_param0 interface{}) interface{} {
	ret := _m.ctrl.Call(_m, "Value", _param0)
	ret0, _ := ret[0].(interface{})
	return ret0
}

func (_mr *_MockContextRecorder) Value(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Value", arg0)
}
