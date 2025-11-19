package repository

import (
	"errors"
)

var (
	ErrInvalidConfig       = errors.New("invalid config")                 // the config is invalid
	ErrInvalidDatabase     = errors.New("invalid database")               // the database is invalid
	ErrInvalidDriver       = errors.New("invalid driver")                 // the driver is invalid
	ErrInvalidPool         = errors.New("invalid pool")                   // the pool is invalid
	ErrInvalidRepository   = errors.New("invalid repository")             // the repository is invalid
	ErrMalformedResult     = errors.New("malformed result")               // the result is malformed
	ErrNoBucket            = errors.New("no bucket")                      // the bucket is missing
	ErrNoClient            = errors.New("no client")                      // the client is missing
	ErrNoDriver            = errors.New("no driver")                      // the driver is missing
	ErrNoLicenseRepository = errors.New("no license repository provided") // no license repository provided
	ErrNoPool              = errors.New("no pool")                        // the pool is nil
	ErrNotFound            = errors.New("resource not found")             // the resource was not found
	ErrReadResourceCount   = errors.New("failed to read resource count")  // the resource count could not be retrieved
	ErrRelationRead        = errors.New("failed to read relation")        // relation cannot be read
	ErrSystemRoleRead      = errors.New("failed to read system role")     // the system role could not be retrieved
)

var ErrUnexpectedCachedResource = errors.New("unexpected cached resource") // received cache resource was not expected
