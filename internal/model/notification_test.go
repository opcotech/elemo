package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotification(t *testing.T) {
	type args struct {
		title     string
		recipient ID
	}
	tests := []struct {
		name    string
		args    args
		want    *Notification
		wantErr error
	}{
		{
			name: "create new notification",
			args: args{
				title:     "Test Notification",
				recipient: ID{Inner: xid.NilID(), Type: ResourceTypeUser},
			},
			want: &Notification{
				ID:        ID{Inner: xid.NilID(), Type: ResourceTypeNotification},
				Title:     "Test Notification",
				Recipient: ID{Inner: xid.NilID(), Type: ResourceTypeUser},
			},
		},
		{
			name: "create new notification with invalid title",
			args: args{
				title:     "he",
				recipient: ID{Inner: xid.NilID(), Type: ResourceTypeUser},
			},
			wantErr: ErrInvalidNotificationDetails,
		},
		{
			name: "create new notification with invalid recipient",
			args: args{
				title:     "Test Notification",
				recipient: ID{Inner: xid.NilID(), Type: ResourceTypeOrganization},
			},
			wantErr: ErrInvalidNotificationRecipient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewNotification(tt.args.title, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNotification_Validate(t *testing.T) {
	type fields struct {
		ID          ID
		Title       string
		Description string
		Recipient   ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid notification",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNotification},
				Title:       "Test Notification",
				Description: "Test description",
				Recipient:   ID{Inner: xid.NilID(), Type: ResourceTypeUser},
			},
		},
		{
			name: "invalid notification title",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNotification},
				Title:       "he",
				Description: "Test description",
				Recipient:   ID{Inner: xid.NilID(), Type: ResourceTypeUser},
			},
			wantErr: ErrInvalidNotificationDetails,
		},
		{
			name: "invalid notification description",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNotification},
				Title:       "Test Notification",
				Description: "Test",
				Recipient:   ID{Inner: xid.NilID(), Type: ResourceTypeUser},
			},
			wantErr: ErrInvalidNotificationDetails,
		},
		{
			name: "invalid notification ID",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceType(0)},
				Title:       "Test Notification",
				Description: "Test description",
				Recipient:   ID{Inner: xid.NilID(), Type: ResourceTypeUser},
			},
			wantErr: ErrInvalidNotificationDetails,
		},
		{
			name: "invalid recipient ID",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNotification},
				Title:       "Test Notification",
				Description: "Test description",
				Recipient:   ID{Inner: xid.NilID(), Type: ResourceType(0)},
			},
			wantErr: ErrInvalidNotificationRecipient,
		},
		{
			name: "invalid recipient ID type",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNotification},
				Title:       "Test Notification",
				Description: "Test description",
				Recipient:   ID{Inner: xid.NilID(), Type: ResourceTypeOrganization},
			},
			wantErr: ErrInvalidNotificationRecipient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &Notification{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				Recipient:   tt.fields.Recipient,
			}
			require.ErrorIs(t, p.Validate(), tt.wantErr)
		})
	}
}
