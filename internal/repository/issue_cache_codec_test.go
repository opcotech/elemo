package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/opcotech/elemo/internal/model"
)

func TestPartialIssueMsgpackRoundTripZeroReportedBy(t *testing.T) {
	t.Parallel()

	page := Page[*PartialIssue]{
		Items: []*PartialIssue{
			{
				ID:         model.MustNewID(model.ResourceTypeIssue),
				Key:        "LMO-1",
				Title:      "Cached issue",
				Kind:       model.IssueKindTask,
				Status:     model.IssueStatusOpen,
				Priority:   model.IssuePriorityNormal,
				ReportedBy: nil,
			},
		},
		PageInfo: PageInfo{HasMore: false},
	}

	encoded, err := msgpack.Marshal(page)
	require.NoError(t, err)

	var decoded Page[*PartialIssue]
	require.NoError(t, msgpack.Unmarshal(encoded, &decoded))
	require.Len(t, decoded.Items, 1)
	assert.Equal(t, page.Items[0].ID, decoded.Items[0].ID)
	assert.Nil(t, decoded.Items[0].ReportedBy)
}
