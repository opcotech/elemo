package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemRole_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Owner", SystemRoleOwner.String())
	assert.Equal(t, "Admin", SystemRoleAdmin.String())
	assert.Equal(t, "Support", SystemRoleSupport.String())
}

func TestSystemRole_MarshalText(t *testing.T) {
	t.Parallel()

	text, err := SystemRoleOwner.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("Owner"), text)

	text, err = SystemRole(0).MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("SystemRole(0)"), text)
}

func TestSystemRole_UnmarshalText(t *testing.T) {
	t.Parallel()

	var role SystemRole
	require.NoError(t, role.UnmarshalText([]byte("Admin")))
	assert.Equal(t, SystemRoleAdmin, role)

	assert.Error(t, role.UnmarshalText([]byte("Unknown")))
}
