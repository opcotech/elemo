package mock

import (
	"context"
	"net/url"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/mock"
)

// Neo4jDriver is the mock implementation of the Neo4j driver.
type Neo4jDriver struct {
	mock.Mock
}

func (n *Neo4jDriver) DefaultExecuteQueryBookmarkManager() neo4j.BookmarkManager {
	args := n.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(neo4j.BookmarkManager)
}

func (n *Neo4jDriver) Target() url.URL {
	args := n.Called()
	return args.Get(0).(url.URL)
}

func (n *Neo4jDriver) NewSession(ctx context.Context, config neo4j.SessionConfig) neo4j.SessionWithContext {
	args := n.Called(ctx, config)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(neo4j.SessionWithContext)
}

func (n *Neo4jDriver) VerifyConnectivity(ctx context.Context) error {
	args := n.Called(ctx)
	return args.Error(0)
}

func (n *Neo4jDriver) Close(ctx context.Context) error {
	args := n.Called(ctx)
	return args.Error(0)
}

func (n *Neo4jDriver) IsEncrypted() bool {
	args := n.Called()
	return args.Bool(0)
}

func (n *Neo4jDriver) GetServerInfo(ctx context.Context) (neo4j.ServerInfo, error) {
	args := n.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(neo4j.ServerInfo), args.Error(1)
}

// Neo4jSession is the mock implementation of the Neo4j session.
type Neo4jSession struct {
	mock.Mock
}

func (n *Neo4jSession) LastBookmarks() neo4j.Bookmarks {
	args := n.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(neo4j.Bookmarks)
}

func (n *Neo4jSession) BeginTransaction(ctx context.Context, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ExplicitTransaction, error) {
	args := n.Called(ctx, configurers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(neo4j.ExplicitTransaction), args.Error(1)
}

func (n *Neo4jSession) ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error) {
	args := n.Called(ctx, work, configurers)
	return args.Get(0), args.Error(1)
}

func (n *Neo4jSession) ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error) {
	args := n.Called(ctx, work, configurers)
	return args.Get(0), args.Error(1)
}

func (n *Neo4jSession) Run(ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) (neo4j.ResultWithContext, error) {
	args := n.Called(ctx, cypher, params, configurers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(neo4j.ResultWithContext), args.Error(1)
}

func (n *Neo4jSession) Close(ctx context.Context) error {
	args := n.Called(ctx)
	return args.Error(0)
}

// Neo4jStore is the mock implementation of the Neo4j Store.
type Neo4jStore struct {
	mock.Mock
}

func (n *Neo4jStore) ReadSession(ctx context.Context) any {
	args := n.Called(ctx)
	return args.Get(0)
}

func (n *Neo4jStore) WriteSession(ctx context.Context) any {
	args := n.Called(ctx)
	return args.Get(0)
}

func (n *Neo4jStore) Ping(ctx context.Context) error {
	args := n.Called(ctx)
	return args.Error(0)
}
