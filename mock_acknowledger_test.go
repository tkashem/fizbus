package fizbus

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of Acknowledger interface
type MockAcknowledger struct {
	ctrl     *gomock.Controller
	recorder *_MockAcknowledgerRecorder
}

// Recorder for MockAcknowledger (not exported)
type _MockAcknowledgerRecorder struct {
	mock *MockAcknowledger
}

func NewMockAcknowledger(ctrl *gomock.Controller) *MockAcknowledger {
	mock := &MockAcknowledger{ctrl: ctrl}
	mock.recorder = &_MockAcknowledgerRecorder{mock}
	return mock
}

func (_m *MockAcknowledger) EXPECT() *_MockAcknowledgerRecorder {
	return _m.recorder
}

func (_m *MockAcknowledger) Ack(_param0 uint64, _param1 bool) error {
	ret := _m.ctrl.Call(_m, "Ack", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockAcknowledgerRecorder) Ack(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Ack", arg0, arg1)
}

func (_m *MockAcknowledger) Nack(_param0 uint64, _param1 bool, _param2 bool) error {
	ret := _m.ctrl.Call(_m, "Nack", _param0, _param1, _param2)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockAcknowledgerRecorder) Nack(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Nack", arg0, arg1, arg2)
}

func (_m *MockAcknowledger) Reject(_param0 uint64, _param1 bool) error {
	ret := _m.ctrl.Call(_m, "Reject", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockAcknowledgerRecorder) Reject(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Reject", arg0, arg1)
}
