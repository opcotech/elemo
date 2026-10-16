package repository_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type UserSeatRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
}

func (s *UserSeatRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, "UserSeatRepositoryIntegrationTestSuite")
}

func (s *UserSeatRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *UserSeatRepositoryIntegrationTestSuite) SetupTest() {
	s.BootstrapNeo4jDatabase(&s.ContainerIntegrationTestSuite)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestCreateAvailableAndReached() {
	ctx := context.Background()
	created, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
	s.Require().NoError(err)
	s.Require().Equal(model.UserStatusActive, created.Status)

	_, err = s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
	s.Require().ErrorIs(err, repository.ErrSeatLimitReached)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestActiveHumanCountIgnoresPendingAndInactive() {
	ctx := context.Background()
	activeOpts := testModel.NewCreateActiveUserOpts(repository.UnrestrictedActivation())
	_, err := s.UserRepo.Create(ctx, activeOpts)
	s.Require().NoError(err)

	pending := testModel.NewCreateUserOpts()
	pending.Status = model.UserStatusPending
	_, err = s.UserRepo.Create(ctx, pending)
	s.Require().NoError(err)

	inactive := testModel.NewCreateUserOpts()
	inactive.Status = model.UserStatusInactive
	_, err = s.UserRepo.Create(ctx, inactive)
	s.Require().NoError(err)

	count, err := s.UserRepo.ActiveHumanCount(ctx)
	s.Require().NoError(err)
	s.Assert().Equal(1, count)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestRawUpdateCannotActivateUsers() {
	ctx := context.Background()
	rawRepo, err := repository.NewNeo4jUserRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	s.Require().NoError(err)

	pending := testModel.NewCreateUserOpts()
	pending.Status = model.UserStatusPending
	user, err := rawRepo.Create(ctx, pending)
	s.Require().NoError(err)
	_, err = rawRepo.Update(ctx, user.ID, repository.UpdateUserOpts{
		Status: optional.Some(model.UserStatusActive),
	})
	s.Require().ErrorIs(err, repository.ErrUserActivationPolicy)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestCreateActiveRequiresExplicitAuthorization() {
	ctx := context.Background()
	opts := testModel.NewCreateUserOpts()
	opts.Activation = nil
	_, err := s.UserRepo.Create(ctx, opts)
	s.Require().ErrorIs(err, repository.ErrUserActivationPolicy)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestMissingInstallationIsNotSeatLimit() {
	ctx := context.Background()
	s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)

	_, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
	s.Require().ErrorIs(err, repository.ErrInstallationRead)
	s.Require().NotErrorIs(err, repository.ErrSeatLimitReached)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestActivateAlreadyActiveIsIdempotent() {
	ctx := context.Background()
	created, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
	s.Require().NoError(err)

	updated, err := s.UserRepo.Activate(ctx, created.ID, repository.UpdateUserOpts{
		Status: optional.Some(model.UserStatusActive),
		Title:  optional.Some("unchanged-active"),
	}, repository.FiniteActivation(1))
	s.Require().NoError(err)
	s.Assert().Equal(model.UserStatusActive, updated.Status)
	s.Assert().Equal("unchanged-active", updated.Title)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestSeatReleaseAllowsNewActivation() {
	ctx := context.Background()
	created, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
	s.Require().NoError(err)

	_, err = s.UserRepo.Update(ctx, created.ID, repository.UpdateUserOpts{
		Status: optional.Some(model.UserStatusInactive),
	})
	s.Require().NoError(err)

	_, err = s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
	s.Require().NoError(err)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestDeletedUserCannotReactivate() {
	ctx := context.Background()
	user, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.UnrestrictedActivation()))
	s.Require().NoError(err)
	_, err = s.UserRepo.Update(ctx, user.ID, repository.UpdateUserOpts{
		Status: optional.Some(model.UserStatusDeleted),
	})
	s.Require().NoError(err)

	_, err = s.UserRepo.Activate(ctx, user.ID, repository.UpdateUserOpts{
		Status: optional.Some(model.UserStatusActive),
	}, repository.UnrestrictedActivation())
	s.Require().ErrorIs(err, repository.ErrUserActivationStatus)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestInvitationAcceptanceIsAtomicAtSeatLimit() {
	ctx := context.Background()
	_, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
	s.Require().NoError(err)

	owner, err := s.UserRepo.Create(ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	org, err := s.OrganizationRepo.Create(ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	invitedOpts := testModel.NewCreateUserOpts()
	invitedOpts.Status = model.UserStatusPending
	invited, err := s.UserRepo.Create(ctx, invitedOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddInvitation(ctx, org.ID, invited.ID))

	_, err = s.UserRepo.AcceptInvitation(ctx, invited.ID, org.ID, "hashed-password", ptrActivation(repository.FiniteActivation(1)), nil)
	s.Require().ErrorIs(err, repository.ErrSeatLimitReached)

	got, err := s.UserRepo.Get(ctx, invited.ID, repository.UserDetailProjection())
	s.Require().NoError(err)
	s.Equal(model.UserStatusPending, got.Status)
	invitations, err := s.OrganizationRepo.GetInvitations(ctx, org.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)
	s.Equal(invited.ID, invitations[0].ID)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestInvitationAcceptanceCannotActivateWithoutPolicyDecision() {
	ctx := context.Background()
	owner, err := s.UserRepo.Create(ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	org, err := s.OrganizationRepo.Create(ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	invitedOpts := testModel.NewCreateUserOpts()
	invitedOpts.Status = model.UserStatusPending
	invited, err := s.UserRepo.Create(ctx, invitedOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddInvitation(ctx, org.ID, invited.ID))

	_, err = s.UserRepo.AcceptInvitation(ctx, invited.ID, org.ID, "", nil, nil)
	s.Require().ErrorIs(err, repository.ErrUserActivationStatus)

	got, err := s.UserRepo.Get(ctx, invited.ID, repository.UserDetailProjection())
	s.Require().NoError(err)
	s.Equal(model.UserStatusPending, got.Status)
	invitations, err := s.OrganizationRepo.GetInvitations(ctx, org.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)
	s.Equal(invited.ID, invitations[0].ID)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestInvitationAcceptanceActivatesAndJoins() {
	ctx := context.Background()
	owner, err := s.UserRepo.Create(ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	org, err := s.OrganizationRepo.Create(ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	invitedOpts := testModel.NewCreateUserOpts()
	invitedOpts.Status = model.UserStatusPending
	invited, err := s.UserRepo.Create(ctx, invitedOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddInvitation(ctx, org.ID, invited.ID))

	accepted, err := s.UserRepo.AcceptInvitation(ctx, invited.ID, org.ID, "hashed-password", ptrActivation(repository.FiniteActivation(2)), nil)
	s.Require().NoError(err)
	s.Equal(model.UserStatusActive, accepted.Status)

	invitations, err := s.OrganizationRepo.GetInvitations(ctx, org.ID)
	s.Require().NoError(err)
	s.Empty(invitations)
	members, err := s.OrganizationRepo.ListMembers(ctx, org.ID, repository.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Require().Len(members.Items, 2)
	var found bool
	for _, member := range members.Items {
		found = found || member.ID == invited.ID
	}
	s.True(found)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestLowerReplacementLimitDoesNotDeactivate() {
	ctx := context.Background()
	first, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(2)))
	s.Require().NoError(err)
	second, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(2)))
	s.Require().NoError(err)

	got, err := s.UserRepo.Get(ctx, first.ID, repository.UserDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(model.UserStatusActive, got.Status)
	got, err = s.UserRepo.Get(ctx, second.ID, repository.UserDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(model.UserStatusActive, got.Status)

	_, err = s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
	s.Require().ErrorIs(err, repository.ErrSeatLimitReached)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestConcurrentCreateHonorsLimit() {
	ctx := context.Background()
	const workers = 8
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			_, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.FiniteActivation(1)))
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	var ok, limited int
	for err := range errs {
		if err == nil {
			ok++
			continue
		}
		require.ErrorIs(s.T(), err, repository.ErrSeatLimitReached)
		limited++
	}
	s.Assert().Equal(1, ok)
	s.Assert().Equal(workers-1, limited)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestConcurrentMixedCreateActivateAndAccept() {
	ctx := context.Background()
	owner, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(repository.UnrestrictedActivation()))
	s.Require().NoError(err)
	org, err := s.OrganizationRepo.Create(ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	pendingOpts := testModel.NewCreateUserOpts()
	pendingOpts.Status = model.UserStatusPending
	pendingOpts.Activation = nil
	pending, err := s.UserRepo.Create(ctx, pendingOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddInvitation(ctx, org.ID, pending.ID))

	inactiveOpts := testModel.NewCreateUserOpts()
	inactiveOpts.Status = model.UserStatusInactive
	inactiveOpts.Activation = nil
	inactive, err := s.UserRepo.Create(ctx, inactiveOpts)
	s.Require().NoError(err)

	limit := repository.FiniteActivation(2)
	var wg sync.WaitGroup
	errs := make(chan error, 3)
	wg.Add(3)
	go func() {
		defer wg.Done()
		_, err := s.UserRepo.Create(ctx, testModel.NewCreateActiveUserOpts(limit))
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := s.UserRepo.Activate(ctx, inactive.ID, repository.UpdateUserOpts{
			Status: optional.Some(model.UserStatusActive),
		}, limit)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := s.UserRepo.AcceptInvitation(ctx, pending.ID, org.ID, "hashed-password", ptrActivation(limit), nil)
		errs <- err
	}()
	wg.Wait()
	close(errs)

	var ok, limited int
	for err := range errs {
		if err == nil {
			ok++
			continue
		}
		require.ErrorIs(s.T(), err, repository.ErrSeatLimitReached)
		limited++
	}
	s.Assert().Equal(1, ok)
	s.Assert().Equal(2, limited)
}

func (s *UserSeatRepositoryIntegrationTestSuite) TestInvitationAcceptanceGrantsRoleAtomically() {
	ctx := context.Background()
	owner, err := s.UserRepo.Create(ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	org, err := s.OrganizationRepo.Create(ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	role, err := s.RoleRepo.Create(ctx, testModel.NewCreateRoleOpts(owner.ID, org.ID))
	s.Require().NoError(err)

	invitedOpts := testModel.NewCreateUserOpts()
	invitedOpts.Status = model.UserStatusPending
	invitedOpts.Activation = nil
	invited, err := s.UserRepo.Create(ctx, invitedOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddInvitation(ctx, org.ID, invited.ID))

	missingRole := model.MustNewID(model.ResourceTypeRole)
	_, err = s.UserRepo.AcceptInvitation(ctx, invited.ID, org.ID, "hashed-password", ptrActivation(repository.UnrestrictedActivation()), &missingRole)
	s.Require().ErrorIs(err, repository.ErrNotFound)

	got, err := s.UserRepo.Get(ctx, invited.ID, repository.UserDetailProjection())
	s.Require().NoError(err)
	s.Equal(model.UserStatusPending, got.Status)
	invitations, err := s.OrganizationRepo.GetInvitations(ctx, org.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)

	accepted, err := s.UserRepo.AcceptInvitation(ctx, invited.ID, org.ID, "hashed-password", ptrActivation(repository.UnrestrictedActivation()), &role.ID)
	s.Require().NoError(err)
	s.Equal(model.UserStatusActive, accepted.Status)

	grants, err := s.PermissionRepo.ListByPrincipal(ctx, invited.ID)
	s.Require().NoError(err)
	var found bool
	for _, grant := range grants {
		if grant.RoleID != nil && *grant.RoleID == role.ID && grant.Scope == org.ID {
			found = true
		}
	}
	s.True(found)
}

func ptrActivation(auth repository.ActivationAuthorization) *repository.ActivationAuthorization {
	return &auth
}

func TestUserSeatRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserSeatRepositoryIntegrationTestSuite))
}
