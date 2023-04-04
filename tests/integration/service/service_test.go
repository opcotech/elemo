//go:build integration

package service

import (
	"time"

	"github.com/opcotech/elemo/internal/config"
)

var (
	neo4jDBConf = &config.GraphDatabaseConfig{
		Host:                         "localhost",
		Port:                         7687,
		Username:                     "neo4j",
		Password:                     "neo4jsecret",
		Database:                     "neo4j",
		MaxTransactionRetryTime:      1,
		MaxConnectionPoolSize:        100,
		MaxConnectionLifetime:        1 * time.Hour,
		ConnectionAcquisitionTimeout: 1 * time.Minute,
		SocketConnectTimeout:         1 * time.Minute,
		SocketKeepalive:              true,
		FetchSize:                    0,
	}

	pgDBConf = &config.RelationalDatabaseConfig{
		Host:                  "localhost",
		Port:                  5432,
		Username:              "elemo",
		Password:              "pgsecret",
		Database:              "elemo",
		IsSecure:              false,
		MaxConnections:        10,
		MaxConnectionLifetime: 10 * time.Minute,
		MaxConnectionIdleTime: 10 * time.Minute,
		MinConnections:        1,
	}
)
