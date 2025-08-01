// Code generated by MockGen. DO NOT EDIT.
// Source: middleware.go
//
// Generated by this command:
//
//	mockgen -source=middleware.go -destination=../../testutil/mock/middleware_gen.go -package=mock -mock_names RateLimiter=RateLimiter
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
	rate "golang.org/x/time/rate"
)

// RateLimiter is a mock of RateLimiter interface.
type RateLimiter struct {
	ctrl     *gomock.Controller
	recorder *RateLimiterMockRecorder
	isgomock struct{}
}

// RateLimiterMockRecorder is the mock recorder for RateLimiter.
type RateLimiterMockRecorder struct {
	mock *RateLimiter
}

// NewRateLimiter creates a new mock instance.
func NewRateLimiter(ctrl *gomock.Controller) *RateLimiter {
	mock := &RateLimiter{ctrl: ctrl}
	mock.recorder = &RateLimiterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *RateLimiter) EXPECT() *RateLimiterMockRecorder {
	return m.recorder
}

// Allow mocks base method.
func (m *RateLimiter) Allow() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Allow")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Allow indicates an expected call of Allow.
func (mr *RateLimiterMockRecorder) Allow() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Allow", reflect.TypeOf((*RateLimiter)(nil).Allow))
}

// AllowN mocks base method.
func (m *RateLimiter) AllowN(t time.Time, n int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllowN", t, n)
	ret0, _ := ret[0].(bool)
	return ret0
}

// AllowN indicates an expected call of AllowN.
func (mr *RateLimiterMockRecorder) AllowN(t, n any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllowN", reflect.TypeOf((*RateLimiter)(nil).AllowN), t, n)
}

// Burst mocks base method.
func (m *RateLimiter) Burst() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Burst")
	ret0, _ := ret[0].(int)
	return ret0
}

// Burst indicates an expected call of Burst.
func (mr *RateLimiterMockRecorder) Burst() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Burst", reflect.TypeOf((*RateLimiter)(nil).Burst))
}

// Limit mocks base method.
func (m *RateLimiter) Limit() rate.Limit {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Limit")
	ret0, _ := ret[0].(rate.Limit)
	return ret0
}

// Limit indicates an expected call of Limit.
func (mr *RateLimiterMockRecorder) Limit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Limit", reflect.TypeOf((*RateLimiter)(nil).Limit))
}

// Tokens mocks base method.
func (m *RateLimiter) Tokens() float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tokens")
	ret0, _ := ret[0].(float64)
	return ret0
}

// Tokens indicates an expected call of Tokens.
func (mr *RateLimiterMockRecorder) Tokens() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tokens", reflect.TypeOf((*RateLimiter)(nil).Tokens))
}

// TokensAt mocks base method.
func (m *RateLimiter) TokensAt(t time.Time) float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TokensAt", t)
	ret0, _ := ret[0].(float64)
	return ret0
}

// TokensAt indicates an expected call of TokensAt.
func (mr *RateLimiterMockRecorder) TokensAt(t any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TokensAt", reflect.TypeOf((*RateLimiter)(nil).TokensAt), t)
}
