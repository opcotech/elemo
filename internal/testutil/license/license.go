package license

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hyperboloide/lk"
	"github.com/stretchr/testify/require"

	testConfig "github.com/opcotech/elemo/internal/testutil/config"
)

// GetKeyPair returns the public and private key pair for testing.
func GetKeyPair(t *testing.T) (public string, private string) {
	privateKey, err := os.ReadFile(testConfig.RootDir + "/configs/test/generator.key")
	require.NoError(t, err)

	publicKey, err := os.ReadFile(testConfig.RootDir + "/configs/test/generator.pub.key")
	require.NoError(t, err)

	return string(publicKey), string(privateKey)
}

// GenerateLicense generates a license for testing.
func GenerateLicense(t *testing.T, key string, license any) string {
	pk, err := lk.PrivateKeyFromB32String(key)
	require.NoError(t, err)

	licenseBytes, err := json.Marshal(license)
	require.NoError(t, err)

	licenseData, err := lk.NewLicense(pk, licenseBytes)
	require.NoError(t, err)

	licenseString, err := licenseData.ToB32String()
	require.NoError(t, err)

	return licenseString
}
