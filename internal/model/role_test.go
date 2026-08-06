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

	_, err = SystemRole(0).MarshalText()
	require.ErrorIs(t, err, ErrInvalidSystemRole)
}

func TestSystemRole_UnmarshalText(t *testing.T) {
	t.Parallel()

	var role SystemRole
	require.NoError(t, role.UnmarshalText([]byte("Admin")))
	assert.Equal(t, SystemRoleAdmin, role)

	require.ErrorIs(t, role.UnmarshalText([]byte("Unknown")), ErrInvalidSystemRole)
}
