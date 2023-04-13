package license

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hyperboloide/lk"

	testConfig "github.com/opcotech/elemo/internal/testutil/config"
)

// GetKeyPair returns the public and private key pair for testing.
func GetKeyPair(t *testing.T) (public string, private string) {
	privateKey, err := os.ReadFile(testConfig.RootDir + "/tests/assets/keys/generator.key")
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := os.ReadFile(testConfig.RootDir + "/tests/assets/keys/generator.pub.key")
	if err != nil {
		t.Fatal(err)
	}

	return string(publicKey), string(privateKey)
}

// GenerateLicense generates a license for testing.
func GenerateLicense(t *testing.T, key string, license any) string {
	pk, err := lk.PrivateKeyFromB32String(key)
	if err != nil {
		t.Fatal(err)
	}

	licenseBytes, err := json.Marshal(license)
	if err != nil {
		t.Fatal(err)
	}

	licenseData, err := lk.NewLicense(pk, licenseBytes)
	if err != nil {
		t.Fatal(err)
	}

	licenseString, err := licenseData.ToB32String()
	if err != nil {
		t.Fatal(err)
	}

	return licenseString
}
