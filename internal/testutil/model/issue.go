package model

import (
	"strconv"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/testutil"
)

func NewIssue(reportedBy model.ID) *model.Issue {
	numericID, err := strconv.Atoi(testutil.GenerateRandomStringNumeric(4))
	if err != nil {
		panic(err)
	}

	issue, err := model.NewIssue(uint(numericID), testutil.GenerateRandomString(10), model.IssueKindStory, reportedBy)
	if err != nil {
		panic(err)
	}

	issue.Description = testutil.GenerateRandomString(10)
	issue.Links = []string{
		"https://example.com/" + testutil.GenerateRandomString(10),
		"https://example.com/" + testutil.GenerateRandomString(10),
	}
	issue.DueDate = convert.ToPointer(time.Now().UTC().Add(24 * time.Hour))

	return issue
}
