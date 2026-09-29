package entitlement

import (
	"context"

	"github.com/opcotech/elemo/internal/deployment"
)

type unrestrictedPolicy struct{}

// Unrestricted is the normal self-hosted policy: every ordinary feature is
// available and human activations are not seat-limited.
func Unrestricted() Policy {
	return unrestrictedPolicy{}
}

func (unrestrictedPolicy) ActivationLimit(_ context.Context) (Limit, error) {
	return Limit{Unlimited: true}, nil
}

func (unrestrictedPolicy) AllowsMutation(_ context.Context) error {
	return nil
}

func (unrestrictedPolicy) Status(_ context.Context) (Status, error) {
	return Status{DeploymentMode: deployment.ModeSelfHosted}, nil
}
