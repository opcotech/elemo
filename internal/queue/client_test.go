package queue

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestWithClientConfig(t *testing.T) {
	type args struct {
		config *config.WorkerConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr error
	}{
		{
			name: "create new option with config",
			args: args{
				config: new(config.WorkerConfig),
			},
			want: &Client{
				conf: new(config.WorkerConfig),
			},
		},
		{
			name: "create new option with nil config",
			args: args{
				config: nil,
			},
			wantErr: config.ErrNoConfig,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := new(Client)
			err := WithClientConfig(tt.args.config)(client)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, client)
			}
		})
	}
}

func TestWithClientLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: new(mock.Logger),
			},
			want: new(mock.Logger),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := new(Client)
			err := WithClientLogger(tt.args.logger)(client)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, client.logger)
		})
	}
}

func TestWithClientTracer(t *testing.T) {
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    tracing.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: new(mock.Tracer),
			},
			want: new(mock.Tracer),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := new(Client)
			err := WithClientTracer(tt.args.tracer)(client)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, client.tracer)
		})
	}
}
