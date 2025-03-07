// Copyright © 2025 Christian Fritz <mail@chr-fritz.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by MockGen. DO NOT EDIT.
// Source: exporter.go

// Package fake is a generated GoMock package.
package fake

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	healthcheck "github.com/heptiolabs/healthcheck"
	prometheus "github.com/prometheus/client_golang/prometheus"
)

// MockExporter is a mock of Exporter interface.
type MockExporter struct {
	ctrl     *gomock.Controller
	recorder *MockExporterMockRecorder
}

// MockExporterMockRecorder is the mock recorder for MockExporter.
type MockExporterMockRecorder struct {
	mock *MockExporter
}

// NewMockExporter creates a new mock instance.
func NewMockExporter(ctrl *gomock.Controller) *MockExporter {
	mock := &MockExporter{ctrl: ctrl}
	mock.recorder = &MockExporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExporter) EXPECT() *MockExporterMockRecorder {
	return m.recorder
}

// AddLivenessCheck mocks base method.
func (m *MockExporter) AddLivenessCheck(name string, check healthcheck.Check) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddLivenessCheck", name, check)
}

// AddLivenessCheck indicates an expected call of AddLivenessCheck.
func (mr *MockExporterMockRecorder) AddLivenessCheck(name, check interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddLivenessCheck", reflect.TypeOf((*MockExporter)(nil).AddLivenessCheck), name, check)
}

// AddReadinessCheck mocks base method.
func (m *MockExporter) AddReadinessCheck(name string, check healthcheck.Check) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddReadinessCheck", name, check)
}

// AddReadinessCheck indicates an expected call of AddReadinessCheck.
func (mr *MockExporterMockRecorder) AddReadinessCheck(name, check interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddReadinessCheck", reflect.TypeOf((*MockExporter)(nil).AddReadinessCheck), name, check)
}

// MustRegister mocks base method.
func (m *MockExporter) MustRegister(collectors ...prometheus.Collector) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range collectors {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "MustRegister", varargs...)
}

// MustRegister indicates an expected call of MustRegister.
func (mr *MockExporterMockRecorder) MustRegister(collectors ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MustRegister", reflect.TypeOf((*MockExporter)(nil).MustRegister), collectors...)
}

// Register mocks base method.
func (m *MockExporter) Register(collector prometheus.Collector) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", collector)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockExporterMockRecorder) Register(collector interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockExporter)(nil).Register), collector)
}

// Run mocks base method.
func (m *MockExporter) Run(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockExporterMockRecorder) Run(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockExporter)(nil).Run), ctx)
}

// Unregister mocks base method.
func (m *MockExporter) Unregister(collector prometheus.Collector) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unregister", collector)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Unregister indicates an expected call of Unregister.
func (mr *MockExporterMockRecorder) Unregister(collector interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unregister", reflect.TypeOf((*MockExporter)(nil).Unregister), collector)
}
