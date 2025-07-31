package mock

import (
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

type PGRows struct {
	mock.Mock
}

func (r *PGRows) Close() {
	r.Called()
}

func (r *PGRows) Err() error {
	return r.Called().Error(0)
}

func (r *PGRows) CommandTag() pgconn.CommandTag {
	return r.Called().Get(0).(pgconn.CommandTag)
}

func (r *PGRows) FieldDescriptions() []pgconn.FieldDescription {
	return r.Called().Get(0).([]pgconn.FieldDescription)
}

func (r *PGRows) Next() bool {
	return r.Called().Bool(0)
}

func (r *PGRows) Scan(dest ...any) error {
	args := r.Called(dest)
	if args.Get(0) == nil {
		return args.Error(1)
	}
	for i, x := range args.Get(0).([]any) {
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(x))
	}

	return args.Error(1)
}

func (r *PGRows) Values() ([]any, error) {
	args := r.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]any), args.Error(1)
	}
	return nil, args.Error(1)
}

func (r *PGRows) RawValues() [][]byte {
	return r.Called().Get(0).([][]byte)
}

func (r *PGRows) Conn() *pgx.Conn {
	return r.Called().Get(0).(*pgx.Conn)
}
