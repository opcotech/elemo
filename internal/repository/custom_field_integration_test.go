package repository_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type CustomFieldRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	scope model.ID
	owner model.ID
}

func (s *CustomFieldRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) SetupTest() {
	s.scope = model.MustNewID(model.ResourceTypeProject)
	s.owner = model.MustNewID(model.ResourceTypeUser)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TearDownTest() {
	s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TearDownSuite() {
	s.CleanupContainers()
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestCreateAndGet() {
	ctx := context.Background()
	created, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)
	s.NotEqual(model.MustNewNilID(model.ResourceTypeCustomFieldDefinition), created.ID)
	s.NotNil(created.CreatedAt)
	s.NotNil(created.Schema.Text)

	got, err := s.CustomFieldRepo.GetDefinition(ctx, created.ID)
	s.Require().NoError(err)
	s.Equal(created.Key, got.Key)
	s.Equal(created.Kind, got.Kind)
	s.Equal(s.scope, got.Scope)
	s.False(got.Archived)
	s.Equal(0, got.SortOrder)

	next, err := s.CustomFieldRepo.NextSortOrder(ctx, s.scope, model.ResourceTypeIssue)
	s.Require().NoError(err)
	s.Equal(1, next)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestKeyConflict() {
	ctx := context.Background()
	first, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)

	dup := testModel.NewCustomFieldDefinition(s.scope, s.owner)
	dup.Key = first.Key
	_, err = s.CustomFieldRepo.CreateDefinition(ctx, dup)
	s.Require().Error(err)
	s.ErrorIs(err, repository.ErrCustomFieldKeyConflict)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestListExcludesArchived() {
	ctx := context.Background()
	active, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)
	archived, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)
	archived.Archived = true
	_, err = s.CustomFieldRepo.UpdateDefinition(ctx, archived)
	s.Require().NoError(err)

	listed, err := s.CustomFieldRepo.ListDefinitions(ctx, []model.ID{s.scope}, model.ResourceTypeIssue, false)
	s.Require().NoError(err)
	s.Len(listed, 1)
	s.Equal(active.ID, listed[0].ID)

	listed, err = s.CustomFieldRepo.ListDefinitions(ctx, []model.ID{s.scope}, model.ResourceTypeIssue, true)
	s.Require().NoError(err)
	s.Len(listed, 2)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestListOrdersAncestorThenSortOrder() {
	ctx := context.Background()
	ancestorScope := model.MustNewID(model.ResourceTypeOrganization)

	ancestor := testModel.NewCustomFieldDefinition(ancestorScope, s.owner)
	ancestor.SortOrder = 9
	createdAncestor, err := s.CustomFieldRepo.CreateDefinition(ctx, ancestor)
	s.Require().NoError(err)

	localLate := testModel.NewCustomFieldDefinition(s.scope, s.owner)
	localLate.SortOrder = 5
	createdLate, err := s.CustomFieldRepo.CreateDefinition(ctx, localLate)
	s.Require().NoError(err)

	localEarly := testModel.NewCustomFieldDefinition(s.scope, s.owner)
	localEarly.SortOrder = 1
	createdEarly, err := s.CustomFieldRepo.CreateDefinition(ctx, localEarly)
	s.Require().NoError(err)

	listed, err := s.CustomFieldRepo.ListDefinitions(
		ctx,
		[]model.ID{s.scope, ancestorScope},
		model.ResourceTypeIssue,
		true,
	)
	s.Require().NoError(err)
	s.Require().Len(listed, 3)
	s.Equal(createdAncestor.ID, listed[0].ID)
	s.Equal(createdEarly.ID, listed[1].ID)
	s.Equal(createdLate.ID, listed[2].ID)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestSelectOptionsAndValues() {
	ctx := context.Background()
	def, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewSelectCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)
	s.Len(def.Schema.Select.Options, 2)

	resource := model.MustNewID(model.ResourceTypeIssue)
	key := "alpha"
	err = s.CustomFieldRepo.ReplaceValues(ctx, def, resource, []model.CustomFieldAtomicValue{{OptionKey: &key}}, true)
	s.Require().NoError(err)

	values, err := s.CustomFieldRepo.ListValues(ctx, resource, false)
	s.Require().NoError(err)
	s.Len(values, 1)
	s.Equal(key, *values[0].Value.OptionKey)

	key = "beta"
	err = s.CustomFieldRepo.ReplaceValues(ctx, def, resource, []model.CustomFieldAtomicValue{{OptionKey: &key}}, true)
	s.Require().NoError(err)
	values, err = s.CustomFieldRepo.ListValues(ctx, resource, false)
	s.Require().NoError(err)
	s.Len(values, 1)
	s.Equal("beta", *values[0].Value.OptionKey)

	count, err := s.CustomFieldRepo.CountValues(ctx, def.ID)
	s.Require().NoError(err)
	s.Equal(int64(1), count)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestUserAndResourceReferenceValues() {
	ctx := context.Background()
	userDef, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewUserReferenceCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)
	resourceDef, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewResourceReferenceCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)

	issue := model.MustNewID(model.ResourceTypeIssue)
	userID := model.MustNewID(model.ResourceTypeUser)
	refID := model.MustNewID(model.ResourceTypeIssue)

	err = s.CustomFieldRepo.ReplaceValues(ctx, userDef, issue, []model.CustomFieldAtomicValue{{UserID: &userID}}, true)
	s.Require().NoError(err)
	err = s.CustomFieldRepo.ReplaceValues(ctx, resourceDef, issue, []model.CustomFieldAtomicValue{{ResourceID: &refID}}, true)
	s.Require().NoError(err)

	values, err := s.CustomFieldRepo.ListValues(ctx, issue, false)
	s.Require().NoError(err)
	s.Require().Len(values, 2)

	byDefinition := make(map[string]model.CustomFieldAtomicValue, len(values))
	for _, stored := range values {
		byDefinition[stored.DefinitionID.String()] = stored.Value
	}
	s.Require().NotNil(byDefinition[userDef.ID.String()].UserID)
	s.Equal(userID, *byDefinition[userDef.ID.String()].UserID)
	s.Require().NotNil(byDefinition[resourceDef.ID.String()].ResourceID)
	s.Equal(refID, *byDefinition[resourceDef.ID.String()].ResourceID)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestDecimalExactnessAndSearch() {
	ctx := context.Background()
	def, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewDecimalCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)

	resource := model.MustNewID(model.ResourceTypeIssue)
	dec := "12.50"
	err = s.CustomFieldRepo.ReplaceValues(ctx, def, resource, []model.CustomFieldAtomicValue{{Decimal: &dec}}, true)
	s.Require().NoError(err)

	values, err := s.CustomFieldRepo.ListValues(ctx, resource, false)
	s.Require().NoError(err)
	s.Require().Len(values, 1)
	s.Equal("12.50", *values[0].Value.Decimal)

	ids, err := s.CustomFieldRepo.Search(ctx, def.ID, repository.CustomFieldPredicate{
		Op:      repository.CustomFieldPredEq,
		Decimal: &dec,
	}, 10)
	s.Require().NoError(err)
	s.Equal([]model.ID{resource}, ids)

	minDecimal := "10"
	ids, err = s.CustomFieldRepo.Search(ctx, def.ID, repository.CustomFieldPredicate{
		Op:      repository.CustomFieldPredGt,
		Decimal: &minDecimal,
	}, 10)
	s.Require().NoError(err)
	s.Equal([]model.ID{resource}, ids)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestIntegerRangeSearch() {
	ctx := context.Background()
	def, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewIntegerCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)

	resource := model.MustNewID(model.ResourceTypeIssue)
	n := int64(8)
	err = s.CustomFieldRepo.ReplaceValues(ctx, def, resource, []model.CustomFieldAtomicValue{{Integer: &n}}, true)
	s.Require().NoError(err)

	bound := int64(5)
	ids, err := s.CustomFieldRepo.Search(ctx, def.ID, repository.CustomFieldPredicate{
		Op:      repository.CustomFieldPredGte,
		Integer: &bound,
	}, 10)
	s.Require().NoError(err)
	s.Equal([]model.ID{resource}, ids)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestStageCommitAbort() {
	ctx := context.Background()
	def, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)
	resource := model.MustNewID(model.ResourceTypeIssue)
	text := "staged"

	op, err := s.CustomFieldRepo.CreateOperation(ctx, repository.CustomFieldOperation{
		Kind:       repository.CustomFieldOpStageValues,
		ResourceID: resource,
	})
	s.Require().NoError(err)
	s.Equal(repository.CustomFieldOpPending, op.Status)

	err = s.CustomFieldRepo.ReplaceValues(ctx, def, resource, []model.CustomFieldAtomicValue{{Text: &text}}, false)
	s.Require().NoError(err)

	hidden, err := s.CustomFieldRepo.ListValues(ctx, resource, false)
	s.Require().NoError(err)
	s.Empty(hidden)

	staged, err := s.CustomFieldRepo.ListValues(ctx, resource, true)
	s.Require().NoError(err)
	s.Len(staged, 1)
	s.False(staged[0].Committed)

	s.Require().NoError(s.CustomFieldRepo.CommitValues(ctx, resource))
	s.Require().NoError(s.CustomFieldRepo.UpdateOperationStatus(ctx, op.ID, repository.CustomFieldOpCommitted))

	committed, err := s.CustomFieldRepo.ListValues(ctx, resource, false)
	s.Require().NoError(err)
	s.Len(committed, 1)
	s.True(committed[0].Committed)

	other := model.MustNewID(model.ResourceTypeIssue)
	err = s.CustomFieldRepo.ReplaceValues(ctx, def, other, []model.CustomFieldAtomicValue{{Text: &text}}, false)
	s.Require().NoError(err)
	s.Require().NoError(s.CustomFieldRepo.AbortValues(ctx, other))
	aborted, err := s.CustomFieldRepo.ListValues(ctx, other, true)
	s.Require().NoError(err)
	s.Empty(aborted)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestDeleteForResource() {
	ctx := context.Background()
	def, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)
	resource := model.MustNewID(model.ResourceTypeIssue)
	text := "gone"
	s.Require().NoError(s.CustomFieldRepo.ReplaceValues(ctx, def, resource, []model.CustomFieldAtomicValue{{Text: &text}}, true))
	s.Require().NoError(s.CustomFieldRepo.DeleteForResource(ctx, resource))
	values, err := s.CustomFieldRepo.ListValues(ctx, resource, true)
	s.Require().NoError(err)
	s.Empty(values)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestDeleteDefinitionWithValuesFails() {
	ctx := context.Background()
	def, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)
	resource := model.MustNewID(model.ResourceTypeIssue)
	text := "held"
	s.Require().NoError(s.CustomFieldRepo.ReplaceValues(ctx, def, resource, []model.CustomFieldAtomicValue{{Text: &text}}, true))

	err = s.CustomFieldRepo.DeleteDefinition(ctx, def.ID)
	s.Require().Error(err)
	s.False(errors.Is(err, repository.ErrNotFound))
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestTransactionRollback() {
	ctx := context.Background()
	first, err := s.CustomFieldRepo.CreateDefinition(ctx, testModel.NewCustomFieldDefinition(s.scope, s.owner))
	s.Require().NoError(err)

	dup := testModel.NewCustomFieldDefinition(s.scope, s.owner)
	dup.Key = first.Key
	_, err = s.CustomFieldRepo.CreateDefinition(ctx, dup)
	s.Require().Error(err)

	listed, err := s.CustomFieldRepo.ListDefinitions(ctx, []model.ID{s.scope}, model.ResourceTypeIssue, true)
	s.Require().NoError(err)
	s.Len(listed, 1)
}

func (s *CustomFieldRepositoryIntegrationTestSuite) TestPendingOperations() {
	ctx := context.Background()
	resource := model.MustNewID(model.ResourceTypeIssue)
	_, err := s.CustomFieldRepo.CreateOperation(ctx, repository.CustomFieldOperation{
		Kind:       repository.CustomFieldOpStageValues,
		ResourceID: resource,
	})
	s.Require().NoError(err)

	ops, err := s.CustomFieldRepo.ListPendingOperations(ctx, time.Now().UTC().Add(time.Second), 10)
	s.Require().NoError(err)
	s.NotEmpty(ops)
}

func TestCustomFieldRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CustomFieldRepositoryIntegrationTestSuite))
}
