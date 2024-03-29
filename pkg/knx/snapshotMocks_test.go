// Code generated by MockGen. DO NOT EDIT.
// Source: snapshot.go

// Package knx is a generated GoMock package.
package knx

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMetricSnapshotHandler is a mock of MetricSnapshotHandler interface.
type MockMetricSnapshotHandler struct {
	ctrl     *gomock.Controller
	recorder *MockMetricSnapshotHandlerMockRecorder
}

// MockMetricSnapshotHandlerMockRecorder is the mock recorder for MockMetricSnapshotHandler.
type MockMetricSnapshotHandlerMockRecorder struct {
	mock *MockMetricSnapshotHandler
}

// NewMockMetricSnapshotHandler creates a new mock instance.
func NewMockMetricSnapshotHandler(ctrl *gomock.Controller) *MockMetricSnapshotHandler {
	mock := &MockMetricSnapshotHandler{ctrl: ctrl}
	mock.recorder = &MockMetricSnapshotHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricSnapshotHandler) EXPECT() *MockMetricSnapshotHandlerMockRecorder {
	return m.recorder
}

// AddSnapshot mocks base method.
func (m *MockMetricSnapshotHandler) AddSnapshot(snapshot *Snapshot) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddSnapshot", snapshot)
}

// AddSnapshot indicates an expected call of AddSnapshot.
func (mr *MockMetricSnapshotHandlerMockRecorder) AddSnapshot(snapshot interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSnapshot", reflect.TypeOf((*MockMetricSnapshotHandler)(nil).AddSnapshot), snapshot)
}

// Close mocks base method.
func (m *MockMetricSnapshotHandler) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockMetricSnapshotHandlerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockMetricSnapshotHandler)(nil).Close))
}

// FindSnapshot mocks base method.
func (m *MockMetricSnapshotHandler) FindSnapshot(key SnapshotKey) (*Snapshot, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindSnapshot", key)
	ret0, _ := ret[0].(*Snapshot)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindSnapshot indicates an expected call of FindSnapshot.
func (mr *MockMetricSnapshotHandlerMockRecorder) FindSnapshot(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindSnapshot", reflect.TypeOf((*MockMetricSnapshotHandler)(nil).FindSnapshot), key)
}

// FindYoungestSnapshot mocks base method.
func (m *MockMetricSnapshotHandler) FindYoungestSnapshot(name string) *Snapshot {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindYoungestSnapshot", name)
	ret0, _ := ret[0].(*Snapshot)
	return ret0
}

// FindYoungestSnapshot indicates an expected call of FindYoungestSnapshot.
func (mr *MockMetricSnapshotHandlerMockRecorder) FindYoungestSnapshot(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindYoungestSnapshot", reflect.TypeOf((*MockMetricSnapshotHandler)(nil).FindYoungestSnapshot), name)
}

// GetMetricsChannel mocks base method.
func (m *MockMetricSnapshotHandler) GetMetricsChannel() chan *Snapshot {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricsChannel")
	ret0, _ := ret[0].(chan *Snapshot)
	return ret0
}

// GetMetricsChannel indicates an expected call of GetMetricsChannel.
func (mr *MockMetricSnapshotHandlerMockRecorder) GetMetricsChannel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricsChannel", reflect.TypeOf((*MockMetricSnapshotHandler)(nil).GetMetricsChannel))
}

// GetValueFunc mocks base method.
func (m *MockMetricSnapshotHandler) GetValueFunc(key SnapshotKey) func() float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValueFunc", key)
	ret0, _ := ret[0].(func() float64)
	return ret0
}

// GetValueFunc indicates an expected call of GetValueFunc.
func (mr *MockMetricSnapshotHandlerMockRecorder) GetValueFunc(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValueFunc", reflect.TypeOf((*MockMetricSnapshotHandler)(nil).GetValueFunc), key)
}

// IsActive mocks base method.
func (m *MockMetricSnapshotHandler) IsActive() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsActive")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsActive indicates an expected call of IsActive.
func (mr *MockMetricSnapshotHandlerMockRecorder) IsActive() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsActive", reflect.TypeOf((*MockMetricSnapshotHandler)(nil).IsActive))
}

// Run mocks base method.
func (m *MockMetricSnapshotHandler) Run() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Run")
}

// Run indicates an expected call of Run.
func (mr *MockMetricSnapshotHandlerMockRecorder) Run() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockMetricSnapshotHandler)(nil).Run))
}
