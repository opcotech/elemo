package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"flag"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/opcotech/elemo/internal/entitlement/license"
)

var (
	keyID           string
	privateKeyFile  string
	outputLicense   string
	customer        string
	installationID  string
	seats           uint
	validityDays    int
	notBeforeOffset time.Duration
)

func parseFlags() error {
	flag.StringVar(&keyID, "key-id", license.ProductionKeyID, "Signing key identifier")
	flag.StringVar(&privateKeyFile, "private-key", "", "Hex-encoded Ed25519 private key file")
	flag.StringVar(&outputLicense, "license", "license.json", "Output license file")
	flag.StringVar(&customer, "customer", "", "Customer name")
	flag.StringVar(&installationID, "installation-id", "", "Logical installation UUID")
	flag.UintVar(&seats, "seats", 0, "Active human seat count")
	flag.IntVar(&validityDays, "validity-period", 365, "License validity period in days from not-before")
	flag.DurationVar(&notBeforeOffset, "not-before-offset", 0, "Offset from now for not_before (may be negative)")
	flag.Parse()

	return validateOptions()
}

func validateOptions() error {
	if privateKeyFile == "" {
		return errors.New("private-key is required")
	}
	if customer == "" {
		return errors.New("customer is required")
	}
	parsedInstallationID, err := uuid.Parse(installationID)
	if err != nil {
		return errors.New("installation-id must be a UUID")
	}
	installationID = parsedInstallationID.String()
	if seats < 1 {
		return errors.New("seats must be at least 1")
	}
	if uint64(seats) > math.MaxUint32 {
		return errors.New("seats must not exceed 4294967295")
	}
	if validityDays <= 0 {
		return errors.New("validity-period must be greater than 0 days")
	}
	if outputLicense == "" {
		return errors.New("license output path is required")
	}
	if keyID == "" {
		return errors.New("key-id is required")
	}
	return nil
}

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	raw, err := os.ReadFile(privateKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	decoded, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		log.Fatal(err)
	}
	if len(decoded) != ed25519.PrivateKeySize {
		log.Fatalf("private key has invalid length %d", len(decoded))
	}

	now := time.Now().UTC().Truncate(time.Second)
	notBefore := now.Add(notBeforeOffset)
	payload := license.Payload{
		ID:             uuid.NewString(),
		Customer:       customer,
		InstallationID: installationID,
		Seats:          uint32(seats), //nolint:gosec // parseFlags bounds seats to uint32.
		IssuedAt:       now,
		NotBefore:      notBefore,
		ExpiresAt:      notBefore.AddDate(0, 0, validityDays),
	}

	signed, err := license.Sign(keyID, ed25519.PrivateKey(decoded), payload)
	if err != nil {
		log.Fatal(err)
	}
	if err := license.WriteFile(outputLicense, signed); err != nil {
		log.Fatal(err)
	}

	log.Printf("License generated: %s", outputLicense)
	log.Printf("license_id=%s installation_id=%s seats=%d expires_at=%s", payload.ID, payload.InstallationID, payload.Seats, payload.ExpiresAt.Format(time.RFC3339))
}
