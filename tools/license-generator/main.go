package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hyperboloide/lk"
	"github.com/rs/xid"

	elemoLicense "github.com/opcotech/elemo/internal/license"
)

var (
	licenseEmail          string
	licenseOrganization   string
	licenseValidityPeriod int
	licenseFeatures       = elemoLicense.DefaultFeatures
	licenseQuotas         = elemoLicense.DefaultQuotas

	privateKeyFile    string
	outputLicenseFile string
)

func parseFlags() error {
	// Company information and validity period
	flag.StringVar(&licenseEmail, "email", "", "License email")
	flag.StringVar(&licenseOrganization, "organization", "", "License organization")
	flag.IntVar(&licenseValidityPeriod, "validity-period", elemoLicense.DefaultValidityPeriod, "License validity period in days")

	// Features
	features := flag.String("features", "", "Comma-separated list of features")

	// Quotas
	quotas := flag.String("quota", "", "Comma-separated key-value pairs of quotas")

	// License keys
	flag.StringVar(&privateKeyFile, "private-key", "", "The private key to use")
	flag.StringVar(&outputLicenseFile, "license", "license.key", "Output license file")
	flag.Parse()

	if licenseEmail == "" {
		return errors.New("email is required")
	}

	if licenseOrganization == "" {
		return errors.New("organization is required")
	}

	if licenseValidityPeriod <= 0 {
		return errors.New("validity period must be greater than 0 days")
	}

	if privateKeyFile == "" {
		return errors.New("no private key provided")
	}

	if outputLicenseFile == "" {
		return errors.New("no output license provided")
	}

	if *features != "" {
		licenseFeatures = make([]elemoLicense.Feature, 0)
		for _, feature := range strings.Split(*features, ",") {
			licenseFeatures = append(licenseFeatures, elemoLicense.Feature(feature))
		}
	}

	if *quotas != "" {
		for _, quota := range strings.Split(*quotas, ",") {
			quotaParts := strings.Split(quota, "=")
			if len(quotaParts) != 2 {
				return errors.New("invalid quota format")
			}

			quotaKey := elemoLicense.Quota(quotaParts[0])
			quotaValue, err := strconv.Atoi(quotaParts[1])
			if err != nil {
				return errors.New("invalid quota value")
			}

			licenseQuotas[quotaKey] = uint32(quotaValue)
		}
	}

	return nil
}

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	privateKey, err := os.ReadFile(privateKeyFile) //
	if err != nil {
		log.Fatal(err)
	}

	key, err := lk.PrivateKeyFromB32String(string(privateKey))
	if err != nil {
		log.Fatal(err)
	}

	licenseBytes, err := json.Marshal(&elemoLicense.License{
		ID:           xid.New(),
		Email:        licenseEmail,
		Organization: licenseOrganization,
		Features:     licenseFeatures,
		Quotas:       licenseQuotas,
		ExpiresAt:    time.Now().AddDate(0, 0, licenseValidityPeriod).UTC(),
	})
	if err != nil {
		log.Fatal(err)
	}

	license, err := lk.NewLicense(key, licenseBytes)
	if err != nil {
		log.Fatal(err)
	}

	licenseData, err := license.ToB32String()
	if err != nil {
		log.Fatal(err)
	}

	// #nosec
	if err := os.WriteFile(outputLicenseFile, []byte(licenseData), 0644); err != nil {
		log.Fatal(err)
	}

	log.Printf("License generated: %s", outputLicenseFile)
}
