package keys

import _ "embed"

// PublicKey is the public key used to validate licenses.
//
//go:embed public.key
var PublicKey string
