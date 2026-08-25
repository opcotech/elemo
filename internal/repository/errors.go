package repository

import (
	"errors"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

const (
	constraintOrganizationSlug      = "organization_slug_unique"
	constraintNamespaceOrgSlug      = "namespace_organization_slug_unique"
	constraintProjectNamespaceKey   = "project_namespace_key_unique"
	neo4jConstraintValidationFailed = "Neo.ClientError.Schema.ConstraintValidationFailed"
)

var (
	ErrInvalidConfig            = errors.New("invalid config")                 // the config is invalid
	ErrInvalidDatabase          = errors.New("invalid database")               // the database is invalid
	ErrInvalidDriver            = errors.New("invalid driver")                 // the driver is invalid
	ErrInvalidPool              = errors.New("invalid pool")                   // the pool is invalid
	ErrInvalidRepository        = errors.New("invalid repository")             // the repository is invalid
	ErrMalformedResult          = errors.New("malformed result")               // the result is malformed
	ErrNoBucket                 = errors.New("no bucket")                      // the bucket is missing
	ErrNoClient                 = errors.New("no client")                      // the client is missing
	ErrNoCacheBackend           = errors.New("no cache backend")               // the cache backend is missing
	ErrNoDriver                 = errors.New("no driver")                      // the driver is missing
	ErrNoLicenseRepository      = errors.New("no license repository provided") // no license repository provided
	ErrNoPool                   = errors.New("no pool")                        // the pool is nil
	ErrNotFound                 = errors.New("resource not found")             // the resource was not found
	ErrSlugConflict             = errors.New("slug already in use")            // organization or namespace slug is not unique
	ErrProjectKeyConflict       = errors.New("project key already in use")     // project key is not unique within its namespace
	ErrReadResourceCount        = errors.New("failed to read resource count")  // the resource count could not be retrieved
	ErrRelationRead             = errors.New("failed to read relation")        // relation cannot be read
	ErrSearchDelete             = errors.New("failed to delete search index")  // search documents cannot be deleted
	ErrSearchFilter             = errors.New("invalid search filter")          // the search filter is invalid
	ErrSearchIndex              = errors.New("failed to update search index")  // search documents cannot be written
	ErrSearchPing               = errors.New("failed to ping search engine")   // search engine is unreachable
	ErrSearchQuery              = errors.New("failed to query search index")   // search cannot be queried
	ErrSystemRoleRead           = errors.New("failed to read system role")     // the system role could not be retrieved
	ErrUnexpectedCachedResource = errors.New("unexpected cached resource")     // received cache resource was not expected
)

func mapUniquenessError(err error) error {
	if err == nil {
		return nil
	}

	var neoErr *neo4j.Neo4jError
	if !errors.As(err, &neoErr) {
		return err
	}
	if neoErr.Code != neo4jConstraintValidationFailed {
		return err
	}

	msg := strings.ToLower(neoErr.Msg)
	switch {
	case strings.Contains(msg, constraintOrganizationSlug),
		strings.Contains(msg, constraintNamespaceOrgSlug):
		return errors.Join(ErrSlugConflict, err)
	case strings.Contains(msg, constraintProjectNamespaceKey):
		return errors.Join(ErrProjectKeyConflict, err)
	case strings.Contains(msg, "`slug`") || strings.Contains(msg, "property slug"):
		return errors.Join(ErrSlugConflict, err)
	case strings.Contains(msg, "`key`") || strings.Contains(msg, "property key"):
		return errors.Join(ErrProjectKeyConflict, err)
	default:
		return err
	}
}
