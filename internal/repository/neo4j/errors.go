package neo4j

import "github.com/neo4j/neo4j-go-driver/v5/neo4j"

var (
	// ErrNoMoreRecords is returned by neo4j.Result.Next() when there are no
	// more records to be read, and the result has been fully consumed, but
	// we are still trying to read more.
	ErrNoMoreRecords = &neo4j.UsageError{
		Message: "Result contains no more records",
	}
)
