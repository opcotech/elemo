package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/license"
	testConfig "github.com/opcotech/elemo/internal/testutil/config"
)

// ParseLicense parses the test license for testing.
func ParseLicense(t *testing.T) *license.License {
	data, err := os.ReadFile(testConfig.RootDir + "/" + testConfig.Conf.License.File)
	require.NoError(t, err)

	publicKey, err := os.ReadFile(testConfig.RootDir + "/tests/assets/keys/generator.pub.key")
	require.NoError(t, err)

	l, err := license.NewLicense(string(data), string(publicKey))
	require.NoError(t, err)

	if !l.Valid() {
		t.Fatal("invalid license")
	}

	return l
}
