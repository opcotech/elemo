package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/opcotech/elemo/internal/entitlement/license"
)

var (
	publicKeyFile  string
	keyID          string
	licenseKeyFile string
	installationID string
)

func parseFlags() error {
	flag.StringVar(&publicKeyFile, "public-key", "", "Optional hex-encoded public key file. Defaults to the embedded vendor trust.")
	flag.StringVar(&keyID, "key-id", license.ProductionKeyID, "Key identifier for -public-key")
	flag.StringVar(&licenseKeyFile, "license", "", "License file to validate")
	flag.StringVar(&installationID, "installation-id", "", "Optional logical installation UUID used for lifecycle evaluation")
	flag.Parse()

	if licenseKeyFile == "" {
		return errors.New("license is required")
	}
	return nil
}

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	raw, err := license.ReadFile(licenseKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	trust, err := loadTrust()
	if err != nil {
		log.Fatal(err)
	}

	verifiedLicense, err := license.Verify(raw, trust)
	if err != nil {
		log.Fatal(err)
	}

	evalInstallation := installationID
	if evalInstallation == "" {
		evalInstallation = verifiedLicense.Payload.InstallationID
	} else {
		parsedInstallationID, err := uuid.Parse(evalInstallation)
		if err != nil {
			log.Fatal("installation-id must be a UUID")
		}
		evalInstallation = parsedInstallationID.String()
	}

	eval := license.EvaluateLicense(verifiedLicense, time.Now().UTC(), evalInstallation)
	out := map[string]any{
		"key_id":          verifiedLicense.KeyID,
		"state":           eval.State,
		"reason":          eval.Reason,
		"installation_id": eval.InstallationID,
		"payload":         verifiedLicense.Payload,
		"grace_ends_at":   eval.GraceEndsAt,
	}
	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(encoded))
}

func loadTrust() (license.Trust, error) {
	if publicKeyFile == "" {
		return license.VendorTrust()
	}
	raw, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return nil, err
	}
	decoded, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, err
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key has invalid length %d", len(decoded))
	}
	if keyID == "" {
		return nil, errors.New("key-id is required with -public-key")
	}
	return license.Trust{keyID: ed25519.PublicKey(decoded)}, nil
}
