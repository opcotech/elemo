package mock

import (
	"time"

	"github.com/stretchr/testify/mock"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	mock.Mock
}

func (m *RateLimiter) Limit() rate.Limit {
	args := m.Called()
	return args.Get(0).(rate.Limit)
}

func (m *RateLimiter) Burst() int {
	args := m.Called()
	return args.Int(0)
}

func (m *RateLimiter) TokensAt(t time.Time) float64 {
	args := m.Called(t)
	return args.Get(0).(float64)
}

func (m *RateLimiter) Tokens() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func (m *RateLimiter) Allow() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *RateLimiter) AllowN(t time.Time, n int) bool {
	args := m.Called(t, n)
	return args.Bool(0)
}
