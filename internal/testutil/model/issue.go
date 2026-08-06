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
func NewCreateIssueOpts(projectID, reportedBy model.ID) repository.CreateIssueOpts {
	numericID, err := strconv.Atoi(pkg.GenerateRandomStringNumeric(4))
	if err != nil {
		panic(err)
	}

	return repository.CreateIssueOpts{
		ProjectID:   projectID,
		NumericID:   uint(numericID), // nolint:gosec
		Kind:        model.IssueKindStory,
		Title:       pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		Status:      model.IssueStatusOpen,
		Priority:    model.IssuePriorityMedium,
		Resolution:  model.IssueResolutionNone,
		ReportedBy:  reportedBy,
		Links: []string{
			"https://example.com/" + pkg.GenerateRandomString(10),
			"https://example.com/" + pkg.GenerateRandomString(10),
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
		ID:          model.MustNewID(model.ResourceTypeIssue),
		NumericID:   uint(numericID), // nolint:gosec
		Kind:        model.IssueKindStory,
		Title:       pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		Status:      model.IssueStatusOpen,
		Priority:    model.IssuePriorityMedium,
		Resolution:  model.IssueResolutionNone,
		ReportedBy:  reportedBy,
		Assignees:   make([]model.ID, 0),
		Labels:      make([]model.ID, 0),
		Comments:    make([]model.ID, 0),
		Attachments: make([]model.ID, 0),
		Watchers:    make([]model.ID, 0),
		Relations:   make([]model.ID, 0),
		Links: []string{
			"https://example.com/" + pkg.GenerateRandomString(10),
			"https://example.com/" + pkg.GenerateRandomString(10),
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
	issue.Links = []string{
		"https://example.com/" + pkg.GenerateRandomString(10),
		"https://example.com/" + pkg.GenerateRandomString(10),
	}
	issue.DueDate = convert.ToPointer(time.Now().UTC().Add(24 * time.Hour))

	return issue
}
