package service

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

func TestMapPage(t *testing.T) {
	t.Parallel()

	token := "next-token"
	total := int64(42)

	tests := []struct {
		name     string
		page     repository.Page[int]
		want     []string
		wantInfo PageInfo
	}{
		{
			name: "maps items and copies page info",
			page: repository.Page[int]{
				Items: []int{1, 2, 3},
				PageInfo: PageInfo{
					HasMore:       true,
					NextPageToken: &token,
					TotalCount:    &total,
				},
			},
			want: []string{"1", "2", "3"},
			wantInfo: PageInfo{
				HasMore:       true,
				NextPageToken: &token,
				TotalCount:    &total,
			},
		},
		{
			name: "empty items",
			page: repository.Page[int]{
				Items: []int{},
				PageInfo: PageInfo{
					HasMore: false,
				},
			},
			want: []string{},
			wantInfo: PageInfo{
				HasMore: false,
			},
		},
		{
			name: "nil items",
			page: repository.Page[int]{
				Items: nil,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mapPage(tt.page, strconv.Itoa)
			assert.Equal(t, tt.want, got.Items)
			assert.Equal(t, tt.wantInfo, got.PageInfo)
		})
	}

	t.Run("does not alias the source items slice", func(t *testing.T) {
		t.Parallel()

		src := repository.Page[int]{Items: []int{1, 2}}
		got := mapPage(src, func(n int) int { return n * 10 })
		require.Equal(t, []int{10, 20}, got.Items)

		got.Items[0] = 99
		assert.Equal(t, []int{1, 2}, src.Items)
	})
}

func TestCursorPageAlias(t *testing.T) {
	t.Parallel()

	page, err := CursorPage{Size: 0}.Normalize()
	require.NoError(t, err)
	assert.Equal(t, repository.DefaultPageSize, page.Size)

	page, err = CursorPage{Size: 10, Token: convert.ToPointer("")}.Normalize()
	require.NoError(t, err)
	assert.Nil(t, page.Token)
}
