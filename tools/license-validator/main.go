package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-json"

	"github.com/hyperboloide/lk"

	elemoLicense "github.com/opcotech/elemo/internal/license"
)

var (
	publicKeyFile  string
	licenseKeyFile string
)

func parseFlags() error {
	flag.StringVar(&publicKeyFile, "public-key", "", "The public key to use")
	flag.StringVar(&licenseKeyFile, "license", "", "License to validate")
	flag.Parse()

	if publicKeyFile == "" {
		return errors.New("no public key provided")
	}

	if licenseKeyFile == "" {
		return errors.New("no license provided")
	}

	return nil
}

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	publicKey, err := os.ReadFile(publicKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	licenseKey, err := os.ReadFile(licenseKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	key, err := lk.PublicKeyFromB32String(string(publicKey))
	if err != nil {
		log.Fatal(err)
	}

	license, err := elemoLicense.NewLicense(string(licenseKey), key.ToB32String())
	if err != nil {
		log.Fatal(err)
	}

	licenseInfo, err := json.MarshalIndent(license, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(licenseInfo))
}
