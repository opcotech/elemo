package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/hyperboloide/lk"
)

var (
	privateKeyFile string
	publicKeyFile  string
)

func parseFlags() error {
	flag.StringVar(&privateKeyFile, "private", "private.key", "Output private key file")
	flag.StringVar(&publicKeyFile, "public", "public.key", "Output public key file")
	flag.Parse()

	if privateKeyFile == "" {
		return errors.New("no private key file provided")
	}

	if publicKeyFile == "" {
		return errors.New("no public key file provided")
	}

	return nil
}

func writeKeyToFile(key, path string) error {
	return os.WriteFile(path, []byte(key), 0644) // #nosec
}

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	privateKey, err := lk.NewPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	privateKeyString, err := privateKey.ToB32String()
	if err != nil {
		log.Fatal(err)
	}

	publicKeyString := privateKey.GetPublicKey().ToB32String()

	if err := writeKeyToFile(privateKeyString, privateKeyFile); err != nil {
		log.Fatal(err)
	}

	if err := writeKeyToFile(publicKeyString, publicKeyFile); err != nil {
		log.Fatal(err)
	}

	log.Println("Private key:", privateKeyFile)
	log.Println("Public key:", publicKeyFile)
}
