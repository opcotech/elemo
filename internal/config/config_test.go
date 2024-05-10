package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisConfig_Address(t *testing.T) {
	t.Parallel()

	conf := RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	assert.Equal(t, "localhost:6379", conf.Address())
}

func TestCacheDatabaseConfig_ConnectionURL(t *testing.T) {
	tests := []struct {
		name string
		c    *CacheDatabaseConfig
		want string
	}{
		{
			name: "secure connection",
			c: &CacheDatabaseConfig{
				RedisConfig: RedisConfig{
					Host:     "localhost",
					Port:     6379,
					Username: "user",
					Password: "secret",
					IsSecure: true,
				},
			},
			want: "rediss://user:secret@localhost:6379/0?sslmode=require",
		},
		{
			name: "unsecure connection",
			c: &CacheDatabaseConfig{
				RedisConfig: RedisConfig{
					Host:     "localhost",
					Port:     6379,
					Database: 0,
					Username: "user",
					Password: "secret",
					IsSecure: false,
				},
			},
			want: "redis://user:secret@localhost:6379/0?sslmode=disable",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.c.ConnectionURL())
		})
	}
}

func TestGraphDatabaseConfig_ConnectionURL(t *testing.T) {
	tests := []struct {
		name string
		c    *GraphDatabaseConfig
		want string
	}{
		{
			name: "secure connection",
			c: &GraphDatabaseConfig{
				Host:     "localhost",
				Port:     7687,
				IsSecure: true,
			},
			want: "neo4j+s://localhost:7687",
		},
		{
			name: "unsecure connection",
			c: &GraphDatabaseConfig{
				Host:     "localhost",
				Port:     7687,
				IsSecure: false,
			},
			want: "neo4j://localhost:7687",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.c.ConnectionURL())
		})
	}
}

func TestRelationalDatabaseConfig_ConnectionURL(t *testing.T) {
	tests := []struct {
		name string
		c    *RelationalDatabaseConfig
		want string
	}{
		{
			name: "secure connection",
			c: &RelationalDatabaseConfig{
				Username: "user",
				Password: "secret",
				Host:     "localhost",
				Port:     7687,
				Database: "database",
				IsSecure: true,
			},
			want: "postgres://user:secret@localhost:7687/database?sslmode=require",
		},
		{
			name: "unsecure connection",
			c: &RelationalDatabaseConfig{
				Username: "user",
				Password: "secret",
				Host:     "localhost",
				Port:     7687,
				Database: "database",
				IsSecure: false,
			},
			want: "postgres://user:secret@localhost:7687/database?sslmode=disable",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.c.ConnectionURL())
		})
	}
}
