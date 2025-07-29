package model

import (
	"strconv"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

func NewIssue(reportedBy model.ID) *model.Issue {
	numericID, err := strconv.Atoi(pkg.GenerateRandomStringNumeric(4))
	if err != nil {
		panic(err)
	}

	issue, err := model.NewIssue(
		uint(numericID), // nolint:gosec
		pkg.GenerateRandomString(10),
		model.IssueKindStory,
		reportedBy,
	)
	if err != nil {
		panic(err)
	}

	issue.Description = pkg.GenerateRandomString(10)
	issue.Links = []string{
		"https://example.com/" + pkg.GenerateRandomString(10),
		"https://example.com/" + pkg.GenerateRandomString(10),
	}
	issue.DueDate = convert.ToPointer(time.Now().UTC().Add(24 * time.Hour))

	return issue
}
