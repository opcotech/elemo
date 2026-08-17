package model

import (
	"strconv"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateIssueOpts creates repository.CreateIssueOpts for tests.
// NumericID is allocated by the repository from the project's counter.
func NewCreateIssueOpts(projectID, reportedBy model.ID) repository.CreateIssueOpts {
	return repository.CreateIssueOpts{
		ProjectID:   projectID,
		Kind:        model.IssueKindStory,
		Title:       pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		Status:      model.IssueStatusOpen,
		Priority:    model.IssuePriorityNormal,
		Resolution:  model.IssueResolutionNone,
		ReportedBy:  reportedBy,
		Links: []model.IssueLink{
			{URL: "https://example.com/" + pkg.GenerateRandomString(10), Label: "Example 1"},
			{URL: "https://example.com/" + pkg.GenerateRandomString(10), Label: "Example 2"},
		},
		DueDate: convert.ToPointer(time.Now().UTC().Add(24 * time.Hour)),
	}
}

// NewRepositoryIssue creates a repository.Issue for mock returns.
func NewRepositoryIssue(reportedBy model.ID) *repository.Issue {
	numericID, err := strconv.Atoi(pkg.GenerateRandomStringNumeric(4))
	if err != nil {
		panic(err)
	}

	return &repository.Issue{
		ID:              model.MustNewID(model.ResourceTypeIssue),
		Key:             model.FormatIssueKey("TEST", uint(numericID)), // nolint:gosec
		NumericID:       uint(numericID),                               // nolint:gosec
		Kind:            model.IssueKindStory,
		Title:           pkg.GenerateRandomString(10),
		Description:     pkg.GenerateRandomString(10),
		Status:          model.IssueStatusOpen,
		Priority:        model.IssuePriorityNormal,
		Resolution:      model.IssueResolutionNone,
		ReportedBy:      &repository.PartialUser{ID: reportedBy},
		Assignments:     make([]repository.PartialAssignee, 0),
		Labels:          make([]repository.PartialLabel, 0),
		CommentCount:    convert.ToPointer(int64(0)),
		DocumentCount:   convert.ToPointer(int64(0)),
		AttachmentCount: convert.ToPointer(int64(0)),
		WatcherCount:    convert.ToPointer(int64(0)),
		RelationCount:   convert.ToPointer(int64(0)),
		Links: []model.IssueLink{
			{URL: "https://example.com/" + pkg.GenerateRandomString(10), Label: "Example 1"},
			{URL: "https://example.com/" + pkg.GenerateRandomString(10), Label: "Example 2"},
		},
		DueDate:   convert.ToPointer(time.Now().UTC().Add(24 * time.Hour)),
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

// NewIssue creates a new issue for tests. It does not create the db record.
//
// Deprecated: prefer NewCreateIssueOpts / NewRepositoryIssue.
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
	issue.Links = []model.IssueLink{
		{URL: "https://example.com/" + pkg.GenerateRandomString(10), Label: "Example 1"},
		{URL: "https://example.com/" + pkg.GenerateRandomString(10), Label: "Example 2"},
	}
	issue.DueDate = convert.ToPointer(time.Now().UTC().Add(24 * time.Hour))

	return issue
}
