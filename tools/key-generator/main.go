package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	privateKeyFile string
	publicKeyFile  string
)

func parseFlags() error {
	flag.StringVar(&privateKeyFile, "private", "private.key", "Output private key file (hex-encoded Ed25519 seed+public)")
	flag.StringVar(&publicKeyFile, "public", "public.key", "Output public key file (hex-encoded Ed25519 public key)")
	flag.Parse()

	if privateKeyFile == "" {
		return errors.New("no private key file provided")
	}
	if publicKeyFile == "" {
		return errors.New("no public key file provided")
	}
	return nil
}

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	if err := writeKeyPair(privateKeyFile, publicKeyFile, priv, pub); err != nil {
		log.Fatal(err)
	}

	log.Println("Private key:", privateKeyFile)
	log.Println("Public key:", publicKeyFile)
}

func writeKeyPair(privatePath, publicPath string, privateKey, publicKey []byte) error {
	if privatePath == publicPath {
		return errors.New("private and public key paths must differ")
	}
	if err := writeNewKeyFile(privatePath, privateKey); err != nil {
		return err
	}
	if err := writeNewKeyFile(publicPath, publicKey); err != nil {
		if removeErr := os.Remove(privatePath); removeErr != nil {
			return errors.Join(err, fmt.Errorf("remove partial private key: %w", removeErr))
		}
		return err
	}
	return nil
}

func writeNewKeyFile(path string, key []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create key file %s: %w", path, err)
	}
	if _, err := file.WriteString(hex.EncodeToString(key) + "\n"); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return fmt.Errorf("write key file %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("close key file %s: %w", path, err)
	}
	return nil
}
