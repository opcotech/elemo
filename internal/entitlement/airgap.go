package entitlement

import (
	"context"
	"time"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement/license"
)

// AirGapPolicy evaluates a verified (or failed) License against an
// injected clock and the logical installation identity.
type AirGapPolicy struct {
	eval    license.Evaluation
	clock   license.Clock
	counter ActiveHumanCounter
}

// NewAirGapPolicy binds a lifecycle evaluation to seat counting. Counter may
// be nil only in tests that never call Status.
func NewAirGapPolicy(eval license.Evaluation, clock license.Clock, counter ActiveHumanCounter) *AirGapPolicy {
	if clock == nil {
		clock = time.Now
	}
	return &AirGapPolicy{eval: eval, clock: clock, counter: counter}
}

func (p *AirGapPolicy) evaluate() license.Evaluation {
	if p.eval.State == license.StateMissing || p.eval.State == license.StateInvalid {
		eval := p.eval
		eval.Now = p.clock().UTC()
		return eval
	}
	return license.EvaluateLicense(p.eval.License, p.clock(), p.eval.InstallationID)
}

func entitledLifecycle(s license.State) bool {
	return s == license.StateValid || s == license.StateGrace
}

func (p *AirGapPolicy) ActivationLimit(_ context.Context) (Limit, error) {
	eval := p.evaluate()
	if !entitledLifecycle(eval.State) {
		return Limit{}, ErrActivationDenied
	}
	return Limit{Max: eval.License.Payload.Seats}, nil
}

func (p *AirGapPolicy) AllowsMutation(_ context.Context) error {
	eval := p.evaluate()
	if !entitledLifecycle(eval.State) {
		return ErrMutationDenied
	}
	return nil
}

func (p *AirGapPolicy) Status(ctx context.Context) (Status, error) {
	eval := p.evaluate()
	status := &AirGapStatus{
		State:          eval.State,
		Reason:         eval.Reason,
		InstallationID: eval.InstallationID,
	}

	if eval.License.Payload.ID != "" {
		payload := eval.License.Payload
		issued := payload.IssuedAt.UTC()
		notBefore := payload.NotBefore.UTC()
		expires := payload.ExpiresAt.UTC()
		graceEnds := eval.GraceEndsAt.UTC()
		status.LicenseID = payload.ID
		status.Customer = payload.Customer
		status.KeyID = eval.License.KeyID
		status.SeatsLicensed = int(payload.Seats)
		status.IssuedAt = &issued
		status.NotBefore = &notBefore
		status.ExpiresAt = &expires
		if eval.State == license.StateGrace || eval.State == license.StateExpired || eval.State == license.StateValid {
			status.GraceEndsAt = &graceEnds
		}
	}

	if p.counter == nil {
		return Status{}, ErrNoActiveHumanCounter
	}

	active, err := p.counter.ActiveHumanCount(ctx)
	if err != nil {
		return Status{}, err
	}
	status.SeatsActive = active

	return Status{
		DeploymentMode: deployment.ModeAirGap,
		AirGap:         status,
	}, nil
}

func (p *AirGapPolicy) Evaluation() license.Evaluation {
	return p.evaluate()
}
