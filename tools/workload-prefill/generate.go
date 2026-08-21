package main

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
)

var firstNames = []string{
	"Ada", "Amir", "Ana", "Ben", "Cam", "Dana", "Elena", "Felix", "Greta", "Hugo",
	"Ines", "Jonas", "Keiko", "Leila", "Marco", "Nina", "Omar", "Pia", "Quinn", "Rosa",
	"Samir", "Tessa", "Uma", "Victor", "Wendy", "Xavier", "Yara", "Zane", "Blair", "Cody",
	"Drew", "Eden", "Farah", "Gabe", "Hana", "Ivan",
}

var lastNames = []string{
	"Adler", "Baker", "Cho", "Diaz", "Ellis", "Frost", "Garcia", "Hayes", "Ito", "Jung",
	"Khan", "Lopez", "Meyer", "Nair", "Ortiz", "Patel", "Quinn", "Reed", "Santos", "Tran",
	"Ueda", "Vega", "Walsh", "Xu", "Young", "Zimmerman", "Brooks", "Clark", "Dunn", "Engel",
	"Ford", "Grant", "Hill", "Ivers", "Jones", "Klein",
}

var migratedNameTemplates = []string{
	"Legacy Billing %d",
	"CRM v%d Archive",
	"Helpdesk %d Import",
	"ERP Module %d",
	"Intranet %d Snapshot",
	"Asset Register %d",
	"Payroll %d",
	"Facilities %d",
	"Vendor Portal %d",
	"Quality Tracker %d",
}

var issueVerbs = []string{
	"Investigate", "Fix", "Document", "Refactor", "Review", "Ship", "Audit", "Plan",
}

var issueNouns = []string{
	"timeout", "dashboard", "export", "onboarding", "permissions", "search",
	"notifications", "billing", "sync", "migration", "cache", "schema",
}

type personSpec struct {
	first    string
	last     string
	username string
	email    string
	title    string
	bio      string
	phone    string
	address  string
	picture  string
	teams    []string
}

func alphaKey(n int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	a := n / (len(alphabet) * len(alphabet))
	b := (n / len(alphabet)) % len(alphabet)
	c := n % len(alphabet)
	return string([]byte{alphabet[a], alphabet[b], alphabet[c]})
}

func generatePeople(rng *rand.Rand, org orgSpec) []personSpec {
	people := make([]personSpec, 0, org.userCount)
	people = append(people, personSpec{
		first:    org.adminFirst,
		last:     org.adminLast,
		username: adminUsernameFor(org),
		email:    org.adminEmail,
		title:    org.adminTitle,
		bio:      fmt.Sprintf("%s %s leads %s and keeps the demo workspace honest.", org.adminFirst, org.adminLast, org.name),
		phone:    phoneNumber(0),
		address:  "100 Market Street, San Francisco, CA",
		picture:  pictureURL(0),
		teams:    adminTeams(org),
	})

	used := map[string]struct{}{org.adminEmail: {}}
	suffix := 1
	for i := 1; i < org.userCount; i++ {
		first := firstNames[rng.IntN(len(firstNames))]
		last := lastNames[rng.IntN(len(lastNames))]
		var username, email string
		for {
			username = fmt.Sprintf("%s-%s-%03d", slug(first), slug(last), suffix)
			email = fmt.Sprintf("%s.%s.%03d@%s", slug(first), slug(last), suffix, org.domain)
			suffix++
			if _, exists := used[email]; !exists {
				break
			}
		}
		used[email] = struct{}{}

		title := titleForTeam(org.teams[i%len(org.teams)].name)
		people = append(people, personSpec{
			first:    first,
			last:     last,
			username: username,
			email:    email,
			title:    title,
			bio:      fmt.Sprintf("%s %s works at %s on %s.", first, last, org.name, strings.ToLower(title)),
			phone:    phoneNumber(i),
			address:  fmt.Sprintf("%d Market Street, San Francisco, CA", 100+i),
			picture:  pictureURL(i),
			teams:    teamsForIndex(org, i),
		})
	}
	return people
}

func adminUsernameFor(org orgSpec) string {
	if org.isMain {
		return adminUsername
	}
	return slug(org.adminFirst) + "-" + slug(org.adminLast)
}

func adminTeams(org orgSpec) []string {
	if len(org.teams) == 0 {
		return []string{}
	}
	if org.isMain && len(org.teams) > 2 {
		return []string{org.teams[0].name, org.teams[2].name}
	}
	return []string{org.teams[0].name}
}

func teamsForIndex(org orgSpec, i int) []string {
	if len(org.teams) == 0 {
		return []string{}
	}
	primary := org.teams[i%len(org.teams)].name
	if org.isMain && primary != "Management" && i%11 == 0 {
		return []string{primary, "Management"}
	}
	return []string{primary}
}

func titleForTeam(team string) string {
	switch team {
	case "Engineering", "Analytics Eng", "SRE":
		return "Software Engineer"
	case "Design", "Studio", "Research":
		return "Product Designer"
	case "Management":
		return "Engineering Manager"
	case "Operations", "Network Ops", "Cloud Operations":
		return "Operations Specialist"
	case "Human Resources":
		return "People Partner"
	case "Sales":
		return "Account Executive"
	case "Client Relations", "Solutions", "Enablement":
		return "Customer Lead"
	case "Product":
		return "Product Manager"
	case "Marketing":
		return "Marketing Manager"
	case "Finance":
		return "Finance Analyst"
	case "Security":
		return "Security Engineer"
	case "Support":
		return "Support Engineer"
	case "Integrations", "Delivery":
		return "Delivery Consultant"
	default:
		return "Team Member"
	}
}

func slug(v string) string {
	return strings.ToLower(strings.ReplaceAll(v, " ", "-"))
}

func phoneNumber(i int) string {
	return fmt.Sprintf("+1555%07d", i%10_000_000)
}

func pictureURL(i int) string {
	return fmt.Sprintf("https://picsum.photos/id/%d/200/200", (i%100)+1)
}

func migratedProject(n int, rng *rand.Rand, issueMin, issueMax int) projectSpec {
	year := 1995 + (n % 25)
	tmpl := migratedNameTemplates[n%len(migratedNameTemplates)]
	name := fmt.Sprintf(tmpl, year)
	issues := issueMin
	if issueMax > issueMin {
		issues = issueMin + rng.IntN(issueMax-issueMin+1)
	}
	return projectSpec{
		key:         alphaKey(n),
		name:        name,
		description: fmt.Sprintf("Archived workspace imported from the %d systems refresh.", year),
		issueCount:  issues,
		live:        false,
	}
}

func nextIssueOpts(rng *rand.Rand, live bool, parent *model.ID) service.CreateIssueOpts {
	kind := issueKind(rng, live)
	status, resolution := issueStatus(rng, live)
	title := issueTitle(rng, kind)
	opts := service.CreateIssueOpts{
		Kind:        kind,
		Title:       title,
		Description: fmt.Sprintf("Generated workload item: %s. Includes acceptance notes for the demo tenant.", strings.ToLower(title)),
		Status:      status,
		Priority:    issuePriority(rng),
		Resolution:  resolution,
	}
	if live && parent != nil && kind != model.IssueKindEpic {
		opts.Parent = parent
	}
	return opts
}

func issueKind(rng *rand.Rand, live bool) model.IssueKind {
	roll := rng.IntN(100)
	if live && roll < 5 {
		return model.IssueKindEpic
	}
	if roll < 25 {
		return model.IssueKindStory
	}
	if roll < 45 {
		return model.IssueKindBug
	}
	return model.IssueKindTask
}

func issueStatus(rng *rand.Rand, live bool) (model.IssueStatus, model.IssueResolution) {
	if !live {
		if rng.IntN(100) < 85 {
			return model.IssueStatusDone, model.IssueResolutionFixed
		}
		return model.IssueStatusClosed, model.IssueResolutionWontFix
	}
	switch rng.IntN(6) {
	case 0:
		return model.IssueStatusOpen, model.IssueResolutionNone
	case 1:
		return model.IssueStatusInProgress, model.IssueResolutionNone
	case 2:
		return model.IssueStatusBlocked, model.IssueResolutionNone
	case 3:
		return model.IssueStatusReview, model.IssueResolutionNone
	case 4:
		return model.IssueStatusDone, model.IssueResolutionFixed
	default:
		return model.IssueStatusClosed, model.IssueResolutionDuplicate
	}
}

func issuePriority(rng *rand.Rand) model.IssuePriority {
	switch rng.IntN(10) {
	case 0:
		return model.IssuePriorityHighest
	case 1:
		return model.IssuePriorityHigh
	case 8, 9:
		return model.IssuePriorityLow
	default:
		return model.IssuePriorityNormal
	}
}

func issueTitle(rng *rand.Rand, kind model.IssueKind) string {
	verb := issueVerbs[rng.IntN(len(issueVerbs))]
	noun := issueNouns[rng.IntN(len(issueNouns))]
	return fmt.Sprintf("%s %s %s", verb, kind.String(), noun)
}

func documentTitle(i int, scope string) string {
	return fmt.Sprintf("%s runbook %03d", scope, i+1)
}

func documentExcerpt(scope string) string {
	return fmt.Sprintf("Internal notes for the %s workspace used in mature-company demos.", scope)
}

func documentBody(title string) []byte {
	return fmt.Appendf(nil, "# %s\n\nThis document was generated for the workload prefill demo.\n\n## Context\n\nUse it to browse libraries, search, and related project links.\n", title)
}
