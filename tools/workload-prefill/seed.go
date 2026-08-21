package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"slices"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/service"
)

type seeder struct {
	deps   *deps
	opts   options
	spec   scenarioSpec
	rng    *rand.Rand
	hash   string
	issues atomic.Int64
}

type seededOrg struct {
	spec         orgSpec
	org          *service.Organization
	admin        *service.User
	users        []*service.User
	people       []personSpec
	teams        map[string]*service.Team
	roles        map[string]model.ID
	namespaces   map[string]*service.Namespace
	projects     map[string]*service.Project
	projectPlans map[string]projectSpec
}

type issueJob struct {
	index     int
	orgName   string
	admin     *service.User
	project   *service.Project
	spec      projectSpec
	assignees []*service.User
}

func seed(ctx context.Context, d *deps, opts options) (seedSummary, error) {
	s := &seeder{
		deps: d,
		opts: opts,
		spec: scenarioFor(opts.profile),
		rng:  rand.New(rand.NewPCG(uint64(opts.seed), 0)), //nolint:gosec // deterministic demo data, not a security source
		hash: password.HashPassword(opts.password),
	}
	return s.run(ctx)
}

func (s *seeder) run(ctx context.Context) (seedSummary, error) {
	s.deps.logger.Info(ctx, "creating installation admin")
	admin, err := s.createUser(ctx, personSpec{
		first:    s.spec.main.adminFirst,
		last:     s.spec.main.adminLast,
		username: adminUsername,
		email:    s.spec.main.adminEmail,
		title:    s.spec.main.adminTitle,
		bio:      "Demo administrator for the Meridian Systems mature-company workspace.",
		phone:    phoneNumber(0),
		address:  "100 Market Street, San Francisco, CA",
		picture:  pictureURL(0),
	})
	if err != nil {
		return seedSummary{}, err
	}

	if _, err := s.deps.permissions.Create(ctx, service.CreateGrantOpts{
		Principal: admin.ID,
		Scope:     model.InstallationID(),
		Actions:   []model.Action{model.ActionOrganizationCreate},
	}); err != nil {
		return seedSummary{}, fmt.Errorf("grant organization.create: %w", err)
	}

	demoCtx := withUser(ctx, admin.ID)
	orgs := make(map[string]*seededOrg, 1+len(s.spec.partners))

	s.deps.logger.Info(ctx, "seeding main organization")
	mainOrg, err := s.seedOrganization(demoCtx, s.spec.main, admin)
	if err != nil {
		return seedSummary{}, err
	}
	orgs[mainOrg.spec.name] = mainOrg

	for _, spec := range s.spec.partners {
		s.deps.logger.Info(ctx, "seeding partner organization", slog.String("organization", spec.name))
		partnerAdmin, err := s.createUser(ctx, personSpec{
			first:    spec.adminFirst,
			last:     spec.adminLast,
			username: adminUsernameFor(spec),
			email:    spec.adminEmail,
			title:    spec.adminTitle,
			bio:      fmt.Sprintf("%s %s leads %s.", spec.adminFirst, spec.adminLast, spec.name),
			phone:    phoneNumber(len(orgs) + 10),
			address:  "200 Partner Way, Austin, TX",
			picture:  pictureURL(len(orgs) + 10),
		})
		if err != nil {
			return seedSummary{}, err
		}
		partner, err := s.seedOrganization(demoCtx, spec, partnerAdmin)
		if err != nil {
			return seedSummary{}, err
		}
		orgs[partner.spec.name] = partner
	}

	if err := s.seedCollaborations(ctx, orgs); err != nil {
		return seedSummary{}, err
	}
	if err := s.seedIssues(ctx, orgs); err != nil {
		return seedSummary{}, err
	}
	docs, err := s.seedDocuments(ctx, orgs)
	if err != nil {
		return seedSummary{}, err
	}

	summary := seedSummary{adminEmail: admin.Email, documents: docs}
	for _, org := range orgs {
		summary.organizations++
		summary.users += len(org.users)
		summary.projects += len(org.projects)
	}
	summary.issues = int(s.issues.Load())
	return summary, nil
}

func (s *seeder) seedOrganization(
	ctx context.Context,
	spec orgSpec,
	admin *service.User,
) (*seededOrg, error) {
	org, err := s.deps.orgs.Create(ctx, admin.ID, service.CreateOrganizationOpts{
		Name:    spec.name,
		Email:   spec.email,
		Website: spec.website,
	})
	if err != nil {
		return nil, fmt.Errorf("create organization %s: %w", spec.name, err)
	}

	adminCtx := withUser(ctx, admin.ID)
	out := &seededOrg{
		spec:         spec,
		org:          org,
		admin:        admin,
		users:        []*service.User{admin},
		people:       generatePeople(s.rng, spec),
		teams:        make(map[string]*service.Team, len(spec.teams)),
		roles:        make(map[string]model.ID),
		namespaces:   make(map[string]*service.Namespace, len(spec.namespaces)),
		projects:     make(map[string]*service.Project),
		projectPlans: make(map[string]projectSpec),
	}

	if err := s.loadRoles(adminCtx, out); err != nil {
		return nil, err
	}

	for i, person := range out.people[1:] {
		user, err := s.createUser(ctx, person)
		if err != nil {
			return nil, fmt.Errorf("create user %d in %s: %w", i+1, spec.name, err)
		}
		if err := s.deps.orgs.AddMember(adminCtx, org.ID, user.ID); err != nil {
			return nil, fmt.Errorf("add member %s: %w", person.email, err)
		}
		out.users = append(out.users, user)
	}

	for _, teamSpec := range spec.teams {
		team, err := s.deps.teams.Create(adminCtx, org.ID, service.CreateTeamOpts{
			Name:        teamSpec.name,
			Description: teamSpec.description,
		})
		if err != nil {
			return nil, fmt.Errorf("create team %s: %w", teamSpec.name, err)
		}
		out.teams[teamSpec.name] = team
	}

	for i, person := range out.people {
		for _, teamName := range person.teams {
			team, ok := out.teams[teamName]
			if !ok {
				continue
			}
			if err := s.deps.teams.AddMember(adminCtx, team.ID, out.users[i].ID, org.ID); err != nil {
				return nil, fmt.Errorf("add %s to team %s: %w", person.email, teamName, err)
			}
		}
	}

	for _, nsSpec := range spec.namespaces {
		ns, err := s.deps.namespaces.Create(adminCtx, org.ID, service.CreateNamespaceOpts{
			Name:        nsSpec.name,
			Description: nsSpec.description,
		})
		if err != nil {
			return nil, fmt.Errorf("create namespace %s: %w", nsSpec.name, err)
		}
		out.namespaces[nsSpec.name] = ns

		projects := make([]projectSpec, 0, len(nsSpec.projects)+nsSpec.migratedProjects)
		projects = append(projects, nsSpec.projects...)
		for n := range nsSpec.migratedProjects {
			projects = append(projects, migratedProject(n, s.rng, nsSpec.migratedIssueMin, nsSpec.migratedIssueMax))
		}
		for _, projectSpec := range projects {
			project, err := s.deps.projects.Create(adminCtx, ns.ID, service.CreateProjectOpts{
				Key:         projectSpec.key,
				Name:        projectSpec.name,
				Description: projectSpec.description,
			})
			if err != nil {
				return nil, fmt.Errorf("create project %s: %w", projectSpec.key, err)
			}
			out.projects[projectSpec.key] = project
			out.projectPlans[projectSpec.key] = projectSpec
		}
	}

	if err := s.applyTeamGrants(adminCtx, out); err != nil {
		return nil, err
	}

	s.deps.logger.Info(ctx, "organization seeded",
		slog.String("organization", spec.name),
		slog.Int("users", len(out.users)),
		slog.Int("projects", len(out.projects)),
	)
	return out, nil
}

func (s *seeder) loadRoles(ctx context.Context, org *seededOrg) error {
	page, err := s.deps.roles.ListBelongsTo(ctx, org.org.ID, service.CursorPage{Size: 100})
	if err != nil {
		return fmt.Errorf("list roles for %s: %w", org.spec.name, err)
	}
	for _, role := range page.Items {
		org.roles[role.Key] = role.ID
	}
	return nil
}

func (s *seeder) applyTeamGrants(ctx context.Context, org *seededOrg) error {
	for _, grant := range org.spec.teamGrants {
		team, ok := org.teams[grant.team]
		if !ok {
			return fmt.Errorf("unknown team %s in %s", grant.team, org.spec.name)
		}
		ns, ok := org.namespaces[grant.namespace]
		if !ok {
			return fmt.Errorf("unknown namespace %s in %s", grant.namespace, org.spec.name)
		}
		roleID, ok := org.roles[grant.roleKey]
		if !ok {
			return fmt.Errorf("unknown role %s in %s", grant.roleKey, org.spec.name)
		}
		if err := s.deps.permissions.GrantRole(ctx, team.ID, ns.ID, roleID); err != nil {
			return fmt.Errorf("grant %s on %s to %s: %w", grant.roleKey, grant.namespace, grant.team, err)
		}
	}
	return nil
}

func (s *seeder) seedCollaborations(ctx context.Context, orgs map[string]*seededOrg) error {
	s.deps.logger.Info(ctx, "applying cross-organization grants")
	for _, collab := range s.spec.collaborations {
		from, ok := orgs[collab.fromOrg]
		if !ok {
			return fmt.Errorf("unknown from org %s", collab.fromOrg)
		}
		to, ok := orgs[collab.toOrg]
		if !ok {
			return fmt.Errorf("unknown to org %s", collab.toOrg)
		}

		switch collab.kind {
		case collabOrgViewer:
			project, ok := to.projects[collab.toProjectKey]
			if !ok {
				return fmt.Errorf("unknown project %s in %s", collab.toProjectKey, to.spec.name)
			}
			roleID, ok := to.roles[model.RoleKeyProjectViewer]
			if !ok {
				return fmt.Errorf("missing project-viewer role in %s", to.spec.name)
			}
			if err := s.deps.permissions.GrantRole(ctx, from.org.ID, project.ID, roleID); err != nil {
				return fmt.Errorf("org viewer grant %s -> %s: %w", from.spec.name, collab.toProjectKey, err)
			}
		case collabTeamMaintainer, collabTeamViewer:
			team, ok := from.teams[collab.fromTeam]
			if !ok {
				return fmt.Errorf("unknown team %s in %s", collab.fromTeam, from.spec.name)
			}
			project, ok := to.projects[collab.toProjectKey]
			if !ok {
				return fmt.Errorf("unknown project %s in %s", collab.toProjectKey, to.spec.name)
			}
			roleKey := model.RoleKeyProjectViewer
			if collab.kind == collabTeamMaintainer {
				roleKey = model.RoleKeyProjectMaintainer
			}
			roleID, ok := to.roles[roleKey]
			if !ok {
				return fmt.Errorf("missing role %s in %s", roleKey, to.spec.name)
			}
			if err := s.deps.permissions.GrantRole(ctx, team.ID, project.ID, roleID); err != nil {
				return fmt.Errorf("team grant %s -> %s: %w", collab.fromTeam, collab.toProjectKey, err)
			}
		case collabDualMember:
			adminCtx := withUser(ctx, to.admin.ID)
			added := 0
			for i, person := range from.people {
				if added >= collab.dualMemberN {
					break
				}
				if i == 0 {
					continue
				}
				if err := s.deps.orgs.AddMember(adminCtx, to.org.ID, from.users[i].ID); err != nil {
					return fmt.Errorf("dual member %s: %w", person.email, err)
				}
				added++
			}
		}
	}
	return nil
}

func (s *seeder) seedIssues(ctx context.Context, orgs map[string]*seededOrg) error {
	orgNames := make([]string, 0, len(orgs))
	for name := range orgs {
		orgNames = append(orgNames, name)
	}
	slices.Sort(orgNames)

	jobs := make([]issueJob, 0)
	index := 0
	for _, orgName := range orgNames {
		org := orgs[orgName]
		keys := make([]string, 0, len(org.projectPlans))
		for key := range org.projectPlans {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, key := range keys {
			spec := org.projectPlans[key]
			if spec.issueCount == 0 {
				continue
			}
			jobs = append(jobs, issueJob{
				index:     index,
				orgName:   org.spec.name,
				admin:     org.admin,
				project:   org.projects[key],
				spec:      spec,
				assignees: org.users,
			})
			index++
		}
	}

	s.deps.logger.Info(ctx, "seeding issues", slog.Int("projects", len(jobs)))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.opts.concurrency)
	for _, job := range jobs {
		g.Go(func() error {
			return s.seedProjectIssues(gctx, job)
		})
	}
	return g.Wait()
}

func (s *seeder) seedProjectIssues(ctx context.Context, job issueJob) error {
	rng := rand.New(rand.NewPCG(uint64(s.opts.seed), uint64(job.index+1))) //nolint:gosec // deterministic demo data, not a security source
	adminCtx := withUser(ctx, job.admin.ID)
	var lastEpic *model.ID

	for i := 0; i < job.spec.issueCount; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var parent *model.ID
		if job.spec.live && lastEpic != nil && i%7 == 0 {
			parent = lastEpic
		}
		opts := nextIssueOpts(rng, job.spec.live, parent)
		issue, err := s.deps.issues.Create(adminCtx, job.project.ID, opts)
		if err != nil {
			return fmt.Errorf("create issue %d in %s/%s: %w", i+1, job.orgName, job.spec.key, err)
		}
		if opts.Kind == model.IssueKindEpic {
			id := issue.ID
			lastEpic = &id
		}
		if job.spec.live && rng.IntN(100) < 40 && len(job.assignees) > 1 {
			assignee := job.assignees[1+rng.IntN(len(job.assignees)-1)]
			if _, err := s.deps.issues.Update(adminCtx, issue.ID, service.UpdateIssueOpts{
				Assignees: optional.Some([]model.ID{assignee.ID}),
			}); err != nil {
				return fmt.Errorf("assign issue %s: %w", issue.Key, err)
			}
		}
		n := s.issues.Add(1)
		if n%500 == 0 {
			s.deps.logger.Info(ctx, "created issues", slog.Int64("count", n))
		}
	}
	return nil
}

func (s *seeder) seedDocuments(ctx context.Context, orgs map[string]*seededOrg) (int, error) {
	s.deps.logger.Info(ctx, "seeding documents", slog.Int("count", s.spec.documentCount))
	main := orgs[s.spec.main.name]

	type target struct {
		admin model.ID
		id    model.ID
		label string
	}
	targets := make([]target, 0, 32)
	targets = append(targets, target{admin: main.admin.ID, id: main.org.ID, label: main.spec.name})
	for name, ns := range main.namespaces {
		targets = append(targets, target{admin: main.admin.ID, id: ns.ID, label: name})
	}
	for key, project := range main.projects {
		if projectSpecLive(main.spec, key) {
			targets = append(targets, target{admin: main.admin.ID, id: project.ID, label: key})
		}
	}
	for _, spec := range s.spec.partners {
		partner := orgs[spec.name]
		targets = append(targets, target{admin: partner.admin.ID, id: partner.org.ID, label: spec.name})
		for name, ns := range partner.namespaces {
			targets = append(targets, target{admin: partner.admin.ID, id: ns.ID, label: name})
		}
	}

	created := 0
	for i := 0; i < s.spec.documentCount; i++ {
		t := targets[i%len(targets)]
		title := documentTitle(i, t.label)
		if _, err := s.deps.documents.Create(withUser(ctx, t.admin), t.id, service.CreateDocumentOpts{
			Title:   title,
			Excerpt: documentExcerpt(t.label),
			Content: documentBody(title),
		}); err != nil {
			return created, fmt.Errorf("create document %s: %w", title, err)
		}
		created++
	}
	return created, nil
}

func projectSpecLive(spec orgSpec, key string) bool {
	for _, ns := range spec.namespaces {
		for _, project := range ns.projects {
			if project.key == key {
				return project.live
			}
		}
	}
	return false
}

func (s *seeder) createUser(ctx context.Context, person personSpec) (*service.User, error) {
	user, err := s.deps.users.Create(ctx, service.CreateUserOpts{
		Username:  person.username,
		Email:     person.email,
		Password:  s.hash,
		FirstName: person.first,
		LastName:  person.last,
		Title:     person.title,
		Bio:       person.bio,
		Phone:     person.phone,
		Address:   person.address,
		Picture:   person.picture,
		Languages: []model.Language{model.LanguageEN},
	})
	if err != nil {
		return nil, fmt.Errorf("create user %s: %w", person.email, err)
	}
	return user, nil
}

func withUser(ctx context.Context, id model.ID) context.Context {
	return context.WithValue(ctx, pkg.CtxKeyUserID, id)
}
