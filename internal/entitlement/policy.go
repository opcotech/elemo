package entitlement

import (
	"context"
	"time"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement/license"
)

// Limit is the finite or unrestricted human-activation ceiling.
type Limit struct {
	Unlimited bool
	Max       uint32
}

// SeatPolicy is consulted when creating or activating a human user.
type SeatPolicy interface {
	ActivationLimit(ctx context.Context) (Limit, error)
}

// MutationPolicy is consulted before user-initiated domain writes.
type MutationPolicy interface {
	AllowsMutation(ctx context.Context) error
}

// Status is the administrator-facing entitlement view.
type Status struct {
	DeploymentMode deployment.Mode
	AirGap         *AirGapStatus
}

// AirGapStatus is present only on the AirGap artifact.
type AirGapStatus struct {
	State          license.State
	Reason         string
	InstallationID string
	LicenseID      string
	Customer       string
	KeyID          string
	SeatsLicensed  int
	SeatsActive    int
	IssuedAt       *time.Time
	NotBefore      *time.Time
	ExpiresAt      *time.Time
	GraceEndsAt    *time.Time
}

// Reporter returns the administrator entitlement status.
type Reporter interface {
	Status(ctx context.Context) (Status, error)
}

// Policy combines seat enforcement, mutation gating, and administrator status.
type Policy interface {
	SeatPolicy
	MutationPolicy
	Reporter
}

// ActiveHumanCounter counts User nodes with status active.
type ActiveHumanCounter interface {
	ActiveHumanCount(ctx context.Context) (int, error)
}
