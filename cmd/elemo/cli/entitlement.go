package cli

import (
	"context"
	"log/slog"
	"time"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/entitlement/license"
	"github.com/opcotech/elemo/internal/repository"
)

func loadEntitlementPolicy(ctx context.Context, graphDB *repository.Neo4jDatabase, counter entitlement.ActiveHumanCounter) entitlement.Policy {
	if !deployment.IsAirGap() {
		logger.Info(ctx, "deployment mode is self_hosted")
		return entitlement.Unrestricted()
	}

	if cfg.AirGap.BillingEmail == "" {
		logger.Warn(ctx, "airgap billing email is not configured; license expiration reminders are disabled")
	}

	instRepo, err := repository.NewNeo4jInstallationRepository(
		repository.WithNeo4jDatabase(graphDB),
		repository.WithNeo4jRepositoryLogger(logger.Named("installation_repository")),
		repository.WithNeo4jRepositoryTracer(tracer),
	)
	if err != nil {
		logger.Error(ctx, "failed to initialize installation repository", slog.Any("error", err))
		return entitlement.NewAirGapPolicy(
			license.InvalidEvaluation(time.Now().UTC(), "", "installation identity unavailable"),
			time.Now,
			counter,
		)
	}

	logicalID, err := instRepo.EnsureLogicalID(ctx)
	if err != nil {
		logger.Error(ctx, "failed to resolve installation identity", slog.Any("error", err))
		return entitlement.NewAirGapPolicy(
			license.InvalidEvaluation(time.Now().UTC(), "", "installation identity unavailable"),
			time.Now,
			counter,
		)
	}

	logger.Info(ctx, "airgap installation identity", slog.String("installation_id", logicalID))

	now := time.Now().UTC()
	path := cfg.AirGap.LicenseFile
	if path == "" {
		logger.Warn(ctx, "airgap license file is not configured; the installation is read-only and new human activations are disabled")
		return entitlement.NewAirGapPolicy(
			license.MissingEvaluation(now, logicalID, "license file not configured"),
			time.Now,
			counter,
		)
	}

	raw, err := license.ReadFile(path)
	if err != nil {
		logger.Warn(ctx, "failed to read airgap license file; the installation is read-only and new human activations are disabled",
			slog.String("path", path),
			slog.Any("error", err),
		)
		return entitlement.NewAirGapPolicy(
			license.MissingEvaluation(now, logicalID, "license file unreadable"),
			time.Now,
			counter,
		)
	}

	trust, err := license.VendorTrust()
	if err != nil {
		logger.Error(ctx, "failed to load airgap vendor trust store", slog.Any("error", err))
		return entitlement.NewAirGapPolicy(
			license.InvalidEvaluation(now, logicalID, "vendor trust store unavailable"),
			time.Now,
			counter,
		)
	}

	eval, err := evaluateAirGapLicense(raw, trust, now, logicalID)
	if err != nil {
		logger.Warn(ctx, "airgap license verification failed; the installation is read-only and new human activations are disabled",
			slog.Any("error", err),
		)
		return entitlement.NewAirGapPolicy(eval, time.Now, counter)
	}

	verifiedLicense := eval.License
	logger.Info(ctx, "airgap license evaluated",
		slog.String("state", eval.State.String()),
		slog.String("license_id", verifiedLicense.Payload.ID),
		slog.String("customer", verifiedLicense.Payload.Customer),
		slog.Int("seats", int(verifiedLicense.Payload.Seats)),
		slog.Time("expires_at", verifiedLicense.Payload.ExpiresAt),
		slog.String("reason", eval.Reason),
	)

	return entitlement.NewAirGapPolicy(eval, time.Now, counter)
}

func evaluateAirGapLicense(raw []byte, trust license.Trust, now time.Time, logicalID string) (license.Evaluation, error) {
	verifiedLicense, err := license.Verify(raw, trust)
	if err != nil {
		return license.InvalidEvaluation(now, logicalID, err.Error()), err
	}
	return license.EvaluateLicense(verifiedLicense, now, logicalID), nil
}
