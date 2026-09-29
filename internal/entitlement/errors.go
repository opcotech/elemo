package entitlement

import "errors"

var (
	ErrNoSeatPolicy         = errors.New("no seat policy provided")
	ErrNoMutationPolicy     = errors.New("no mutation policy provided")
	ErrNoActiveHumanCounter = errors.New("no active human counter provided")
	ErrSeatLimitReached     = errors.New("licensed seat limit reached")
	ErrActivationDenied     = errors.New("user activation is not entitled")
	ErrMutationDenied       = errors.New("installation is read-only")
	ErrNoEntitlementService = errors.New("no entitlement service provided")
)
