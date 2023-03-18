package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoPriority_String(t *testing.T) {
	tests := []struct {
		name string
		p    TodoPriority
		want string
	}{
		{"normal", TodoPriorityNormal, "normal"},
		{"important", TodoPriorityImportant, "important"},
		{"urgent", TodoPriorityUrgent, "urgent"},
		{"critical", TodoPriorityCritical, "critical"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.p.String())
		})
	}
}

func TestTodoPriority_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		p        TodoPriority
		wantText []byte
		wantErr  bool
	}{
		{"normal", TodoPriorityNormal, []byte("normal"), false},
		{"important", TodoPriorityImportant, []byte("important"), false},
		{"urgent", TodoPriorityUrgent, []byte("urgent"), false},
		{"critical", TodoPriorityCritical, []byte("critical"), false},
		{"status high", TodoPriority(100), []byte("100"), true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.p.MarshalText()
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.wantText, gotText)
			}
		})
	}
}

func TestTodoPriority_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		p       *TodoPriority
		text    []byte
		want    TodoPriority
		wantErr bool
	}{
		{"normal", new(TodoPriority), []byte("normal"), TodoPriorityNormal, false},
		{"important", new(TodoPriority), []byte("important"), TodoPriorityImportant, false},
		{"urgent", new(TodoPriority), []byte("urgent"), TodoPriorityUrgent, false},
		{"critical", new(TodoPriority), []byte("critical"), TodoPriorityCritical, false},
		{"status high", new(TodoPriority), []byte("100"), TodoPriorityNormal, true},
		{"status invalid", new(TodoPriority), []byte("invalid"), TodoPriorityNormal, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.p.UnmarshalText(tt.text); (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewTodo(t *testing.T) {
	type args struct {
		title     string
		ownedBy   ID
		createdBy ID
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "create todo",
			args: args{
				title:     "title",
				ownedBy:   ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: UserIDType},
				createdBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: UserIDType},
			},
			wantErr: nil,
		},
		{
			name: "create todo with empty title",
			args: args{
				title:     "",
				ownedBy:   ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: UserIDType},
				createdBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: UserIDType},
			},
			wantErr: ErrInvalidTodoDetails,
		},
		{
			name: "create todo with no owner",
			args: args{
				title:     "title",
				ownedBy:   ID{},
				createdBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: UserIDType},
			},
			wantErr: ErrInvalidTodoDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			todo, err := NewTodo(tt.args.title, tt.args.ownedBy, tt.args.createdBy)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.NotNil(t, todo.ID)
				assert.Equal(t, tt.args.title, todo.Title)
				assert.Equal(t, TodoPriorityNormal, todo.Priority)
				assert.Equal(t, tt.args.ownedBy, todo.OwnedBy)
				assert.Equal(t, tt.args.createdBy, todo.CreatedBy)
				assert.Nil(t, todo.DueDate)
				assert.Nil(t, todo.CreatedAt)
				assert.Nil(t, todo.UpdatedAt)
			}
		})
	}
}

func TestTodo_Validate(t1 *testing.T) {
	type fields struct {
		ID          ID
		Title       string
		Description string
		Priority    TodoPriority
		OwnedBy     ID
		CreatedBy   ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid todo",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: TodoIDType},
				Title:       "title",
				Description: "description",
				Priority:    TodoPriorityNormal,
				OwnedBy:     ID{inner: xid.NilID(), label: UserIDType},
				CreatedBy:   ID{inner: xid.NilID(), label: UserIDType},
			},
		},
		{
			name: "invalid todo id",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ""},
				Title:       "title",
				Description: "description",
				Priority:    TodoPriorityNormal,
				OwnedBy:     ID{inner: xid.NilID(), label: UserIDType},
				CreatedBy:   ID{inner: xid.NilID(), label: UserIDType},
			},
			wantErr: ErrInvalidTodoDetails,
		},
		{
			name: "invalid todo title",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: TodoIDType},
				Title:       "",
				Description: "description",
				Priority:    TodoPriorityNormal,
				OwnedBy:     ID{inner: xid.NilID(), label: UserIDType},
				CreatedBy:   ID{inner: xid.NilID(), label: UserIDType},
			},
			wantErr: ErrInvalidTodoDetails,
		},
		{
			name: "invalid todo description",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: TodoIDType},
				Title:       "title",
				Description: "desc",
				Priority:    TodoPriorityNormal,
				OwnedBy:     ID{inner: xid.NilID(), label: UserIDType},
				CreatedBy:   ID{inner: xid.NilID(), label: UserIDType},
			},
			wantErr: ErrInvalidTodoDetails,
		},
		{
			name: "invalid todo priority",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: TodoIDType},
				Title:       "title",
				Description: "description",
				Priority:    TodoPriority(0),
				OwnedBy:     ID{inner: xid.NilID(), label: UserIDType},
				CreatedBy:   ID{inner: xid.NilID(), label: UserIDType},
			},
			wantErr: ErrInvalidTodoDetails,
		},
		{
			name: "invalid todo owner",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: TodoIDType},
				Title:       "title",
				Description: "description",
				Priority:    TodoPriorityNormal,
				OwnedBy:     ID{inner: xid.NilID(), label: ""},
				CreatedBy:   ID{inner: xid.NilID(), label: UserIDType},
			},
			wantErr: ErrInvalidTodoDetails,
		},
		{
			name: "invalid todo creator",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: TodoIDType},
				Title:       "title",
				Description: "description",
				Priority:    TodoPriorityNormal,
				OwnedBy:     ID{inner: xid.NilID(), label: UserIDType},
				CreatedBy:   ID{inner: xid.NilID(), label: ""},
			},
			wantErr: ErrInvalidTodoDetails,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Todo{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				Priority:    tt.fields.Priority,
				OwnedBy:     tt.fields.OwnedBy,
				CreatedBy:   tt.fields.CreatedBy,
			}
			require.ErrorIs(t1, t.Validate(), tt.wantErr)
		})
	}
}
