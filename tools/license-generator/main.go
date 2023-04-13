package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hyperboloide/lk"
	"github.com/rs/xid"

	elemoLicense "github.com/opcotech/elemo/internal/license"
)

var (
	licenseEmail          string
	licenseOrganization   string
	quotaOrganizations    int
	quotaSeats            int
	licenseValidityPeriod int
	licenseFeatures       []elemoLicense.Feature

	privateKeyFile    string
	outputLicenseFile string

	defaultLicenseFeatures = func() string {
		features := make([]string, 0, len(elemoLicense.DefaultFeatures))
		for _, feature := range elemoLicense.DefaultFeatures {
			features = append(features, string(feature))
		}

		return strings.Join(features, ",")
	}()
)

func parseFlags() error {
	// Company information and validity period
	flag.StringVar(&licenseEmail, "email", "", "License email")
	flag.StringVar(&licenseOrganization, "organization", "", "License organization")
	flag.IntVar(&licenseValidityPeriod, "validity-period", elemoLicense.DefaultValidityPeriod, "License validity period in days")

	// Features
	features := flag.String("features", defaultLicenseFeatures, "comma-separated features")

	// Quotas
	flag.IntVar(&quotaOrganizations, "quota-organizations", elemoLicense.DefaultQuotas[elemoLicense.QuotaOrganizations], "License custom status quota")
	flag.IntVar(&quotaSeats, "quota-seats", elemoLicense.DefaultQuotas[elemoLicense.QuotaSeats], "License seat quota")

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

	if quotaOrganizations <= 0 {
		return errors.New("organizations must be greater than 0")
	}

	if quotaSeats <= 0 {
		return errors.New("seats must be greater than 0")
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

	if *features == "" {
		return errors.New("no features provided")
	}

	for _, feature := range strings.Split(*features, ",") {
		licenseFeatures = append(licenseFeatures, elemoLicense.Feature(feature))
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
		Quotas: map[elemoLicense.Quota]int{
			elemoLicense.QuotaOrganizations: quotaOrganizations,
			elemoLicense.QuotaSeats:         quotaSeats,
		},
		ExpiresAt: time.Now().AddDate(0, 0, licenseValidityPeriod).UTC(),
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
