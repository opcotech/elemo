package license

import (
	"time"

	"github.com/google/uuid"
)

const (
	StateValid       State = iota + 1 // valid
	StateGrace                        // grace
	StateExpired                      // expired
	StateNotYetValid                  // not_yet_valid
	StateInvalid                      // invalid
	StateMissing                      // missing
)

// State is the evaluated lifecycle of an License.
//
//go:generate go tool enumer -type=State -text -transform=noop -linecomment -output=lifecycle_state_gen.go
type State uint8

// Evaluation is the clocked lifecycle of a verified or failed license.
type Evaluation struct {
	State          State
	License        License
	Now            time.Time
	GraceEndsAt    time.Time
	InstallationID string
	Reason         string
}

// EvaluateLicense evaluates a verified license against now and the logical
// installation identity.
func EvaluateLicense(lic License, now time.Time, installationID string) Evaluation {
	now = now.UTC()
	eval := Evaluation{
		License:        lic,
		Now:            now,
		InstallationID: installationID,
		GraceEndsAt:    lic.Payload.ExpiresAt.UTC().Add(GracePeriod),
	}

	licensedInstallationID, licensedErr := uuid.Parse(lic.Payload.InstallationID)
	currentInstallationID, currentErr := uuid.Parse(installationID)
	if licensedErr != nil || currentErr != nil || licensedInstallationID != currentInstallationID {
		eval.State = StateInvalid
		eval.Reason = "installation mismatch"
		return eval
	}

	notBefore := lic.Payload.NotBefore.UTC()
	if now.Add(NotBeforeSkew).Before(notBefore) {
		eval.State = StateNotYetValid
		eval.Reason = "not yet valid"
		return eval
	}

	expires := lic.Payload.ExpiresAt.UTC()
	if !now.After(expires) {
		eval.State = StateValid
		return eval
	}
	if !now.After(eval.GraceEndsAt) {
		eval.State = StateGrace
		eval.Reason = "in grace period"
		return eval
	}

	eval.State = StateExpired
	eval.Reason = "license expired"
	return eval
}

// MissingEvaluation is used when no license file is configured or readable.
func MissingEvaluation(now time.Time, installationID, reason string) Evaluation {
	return Evaluation{
		State:          StateMissing,
		Now:            now.UTC(),
		InstallationID: installationID,
		Reason:         reason,
	}
}

// InvalidEvaluation is used when verification fails.
func InvalidEvaluation(now time.Time, installationID, reason string) Evaluation {
	return Evaluation{
		State:          StateInvalid,
		Now:            now.UTC(),
		InstallationID: installationID,
		Reason:         reason,
	}
}

// Clock returns the current time. Tests inject a frozen clock.
type Clock func() time.Time
