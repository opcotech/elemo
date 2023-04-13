package license

import (
	"testing"
	"time"

	"github.com/hyperboloide/lk"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testLicense "github.com/opcotech/elemo/internal/testutil/license"
)

const (
	testFeature Feature = "test"
)

func TestNewLicense(t *testing.T) {
	publicKey, privateKey := testLicense.GetKeyPair(t)
	licenseID := xid.New()
	expiresAt := time.Now().AddDate(0, 0, 1).UTC()

	tests := []struct {
		name     string
		license  string
		expected *License
		wantErr  bool
	}{
		{
			name: "valid license",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas:       DefaultQuotas,
				Features:     DefaultFeatures,
				ExpiresAt:    expiresAt,
			}),
			expected: &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas:       DefaultQuotas,
				Features:     DefaultFeatures,
				ExpiresAt:    expiresAt,
			},
		},
		{
			name: "missing features field",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas:       DefaultQuotas,
				ExpiresAt:    expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "missing email field",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Organization: "Test",
				Quotas:       DefaultQuotas,
				Features:     DefaultFeatures,
				ExpiresAt:    expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "missing organization field",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:        licenseID,
				Email:     "test@example.com",
				Quotas:    DefaultQuotas,
				Features:  DefaultFeatures,
				ExpiresAt: expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "empty users quota",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas: func() map[Quota]uint32 {
					quotas := make(map[Quota]uint32)
					for k, v := range DefaultQuotas {
						quotas[k] = v
					}
					quotas[QuotaUsers] = 0
					return quotas
				}(),
				Features:  DefaultFeatures,
				ExpiresAt: expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "empty documents quota",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas: func() map[Quota]uint32 {
					quotas := make(map[Quota]uint32)
					for k, v := range DefaultQuotas {
						quotas[k] = v
					}
					quotas[QuotaDocuments] = 0
					return quotas
				}(),
				ExpiresAt: expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "empty namespaces quota",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas: func() map[Quota]uint32 {
					quotas := make(map[Quota]uint32)
					for k, v := range DefaultQuotas {
						quotas[k] = v
					}
					quotas[QuotaNamespaces] = 0
					return quotas
				}(),
				ExpiresAt: expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "empty organizations quota",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas: func() map[Quota]uint32 {
					quotas := make(map[Quota]uint32)
					for k, v := range DefaultQuotas {
						quotas[k] = v
					}
					quotas[QuotaOrganizations] = 0
					return quotas
				}(),
				ExpiresAt: expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "empty projects quota",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas: func() map[Quota]uint32 {
					quotas := make(map[Quota]uint32)
					for k, v := range DefaultQuotas {
						quotas[k] = v
					}
					quotas[QuotaProjects] = 0
					return quotas
				}(),
				ExpiresAt: expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "empty roles quota",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas: func() map[Quota]uint32 {
					quotas := make(map[Quota]uint32)
					for k, v := range DefaultQuotas {
						quotas[k] = v
					}
					quotas[QuotaRoles] = 0
					return quotas
				}(),
				ExpiresAt: expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "empty users quota",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas: func() map[Quota]uint32 {
					quotas := make(map[Quota]uint32)
					for k, v := range DefaultQuotas {
						quotas[k] = v
					}
					quotas[QuotaUsers] = 0
					return quotas
				}(),
				ExpiresAt: expiresAt,
			}),
			wantErr: true,
		},
		{
			name: "expired license",
			license: testLicense.GenerateLicense(t, privateKey, &License{
				ID:           licenseID,
				Email:        "test@example.com",
				Organization: "Test",
				Quotas:       DefaultQuotas,
				Features:     DefaultFeatures,
				ExpiresAt:    time.Now().AddDate(0, 0, -1),
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewLicense(tt.license, publicKey)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestNewLicense_Errors(t *testing.T) {
	publicKey, privateKey := testLicense.GetKeyPair(t)

	type args struct {
		license   string
		publicKey string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "invalid private key",
			args: args{
				license:   testLicense.GenerateLicense(t, privateKey, new(License)),
				publicKey: publicKey,
			},
		},
		{
			name: "invalid public key",
			args: args{
				license:   testLicense.GenerateLicense(t, privateKey, new(License)),
				publicKey: "invalid",
			},
		},
		{
			name: "invalid license public key",
			args: args{
				license: testLicense.GenerateLicense(t, privateKey, new(License)),
				publicKey: func() string {
					privateKey, _ := lk.NewPrivateKey()
					return privateKey.GetPublicKey().ToB32String()
				}(),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewLicense(tt.args.license, tt.args.publicKey)
			require.Error(t, err)
		})
	}
}

func TestHasFeature(t *testing.T) {
	tests := []struct {
		name    string
		license *License
		feature Feature
		want    bool
	}{
		{
			name: "license has feature",
			license: &License{Features: []Feature{
				testFeature,
			}},
			feature: testFeature,
			want:    true,
		},
		{
			name:    "license does not have feature",
			license: &License{},
			feature: testFeature,
			want:    false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.license.HasFeature(tt.feature))
		})
	}
}

func TestIsWithinThreshold(t *testing.T) {
	tests := []struct {
		name    string
		license *License
		quota   Quota
		value   int
		want    bool
	}{
		{
			name: "license is within threshold",
			license: &License{Quotas: map[Quota]uint32{
				QuotaUsers: 5,
			}},
			quota: QuotaUsers,
			value: 4,
			want:  true,
		},
		{
			name: "license is on the threshold",
			license: &License{Quotas: map[Quota]uint32{
				QuotaUsers: 5,
			}},
			quota: QuotaUsers,
			value: 5,
			want:  true,
		},
		{
			name: "license is not within threshold",
			license: &License{Quotas: map[Quota]uint32{
				QuotaUsers: 5,
			}},
			quota: QuotaUsers,
			value: 6,
			want:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.license.WithinThreshold(tt.quota, tt.value))
		})
	}
}
