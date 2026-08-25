package main

import "strings"

const (
	profileFull  = "full"
	profileSmoke = "smoke"

	defaultPassword = "AppleTree123"
	adminEmail      = "demo@elemo.example"
	adminUsername   = "demo"
	mainOrgName     = "Elemo"
	migratedNS      = "Migrated"
)

type orgSpec struct {
	name       string
	slug       string
	email      string
	domain     string
	website    string
	userCount  int
	adminFirst string
	adminLast  string
	adminEmail string
	adminTitle string
	teams      []teamSpec
	namespaces []namespaceSpec
	teamGrants []teamGrantSpec
	isMain     bool
}

type teamSpec struct {
	name        string
	description string
}

type namespaceSpec struct {
	name             string
	slug             string
	description      string
	projects         []projectSpec
	migratedProjects int
	migratedIssueMin int
	migratedIssueMax int
}

type projectSpec struct {
	key         string
	name        string
	description string
	issueCount  int
	live        bool
}

type teamGrantSpec struct {
	team      string
	namespace string
	roleKey   string
}

type collabKind int

const (
	collabOrgViewer collabKind = iota
	collabTeamMaintainer
	collabTeamViewer
	collabDualMember
)

type collabSpec struct {
	kind         collabKind
	fromOrg      string
	fromTeam     string
	toOrg        string
	toProjectKey string
	dualMemberN  int
}

type scenarioSpec struct {
	main           orgSpec
	partners       []orgSpec
	collaborations []collabSpec
	documentCount  int
}

func scenarioFor(profile string) scenarioSpec {
	if profile == profileSmoke {
		return smokeScenario()
	}
	return fullScenario()
}

func fullScenario() scenarioSpec {
	engineering := teamSpec{name: "Engineering", description: "Builds and maintains Elemo product and platform."}
	design := teamSpec{name: "Design", description: "Product design, research, and the design system."}
	management := teamSpec{name: "Management", description: "Leadership team spanning product and operations."}
	operations := teamSpec{name: "Operations", description: "Runs internal tooling, billing, and procurement."}
	hr := teamSpec{name: "Human Resources", description: "People operations and employee experience."}
	sales := teamSpec{name: "Sales", description: "Enterprise sales and account executives."}
	clientRel := teamSpec{name: "Client Relations", description: "Customer success and named-account care."}
	product := teamSpec{name: "Product", description: "Product management and roadmap ownership."}
	marketing := teamSpec{name: "Marketing", description: "Demand generation and product marketing."}
	finance := teamSpec{name: "Finance", description: "Controller, FP&A, and revenue operations."}
	security := teamSpec{name: "Security", description: "Security engineering and compliance."}
	support := teamSpec{name: "Support", description: "Technical support and incident communication."}

	return applySlugs(scenarioSpec{
		documentCount: 280,
		main: orgSpec{
			name:       mainOrgName,
			email:      "hello@elemo.example",
			domain:     "elemo.example",
			website:    "https://elemo.example",
			userCount:  300,
			adminFirst: "Demo",
			adminLast:  "User",
			adminEmail: adminEmail,
			adminTitle: "Organization Admin",
			isMain:     true,
			teams: []teamSpec{
				engineering, design, management, operations, hr, sales,
				clientRel, product, marketing, finance, security, support,
			},
			namespaces: []namespaceSpec{
				{
					name:        "Product",
					description: "Customer-facing product lines and the design system.",
					projects: []projectSpec{
						{key: "CORE", name: "Core Platform", description: "Shared product services and identity.", issueCount: 120, live: true},
						{key: "MOBL", name: "Mobile Apps", description: "iOS and Android client applications.", issueCount: 110, live: true},
						{key: "DESN", name: "Design System", description: "Component library and brand guidelines.", issueCount: 100, live: true},
						{key: "APIS", name: "Public API", description: "External developer API and SDK work.", issueCount: 100, live: true},
						{key: "WEB", name: "Web App", description: "Browser workspace and account settings.", issueCount: 100, live: true},
					},
				},
				{
					name:        "Platform",
					description: "Internal developer platform and data infrastructure.",
					projects: []projectSpec{
						{key: "PLAT", name: "Platform Services", description: "CI, delivery, and shared runtime services.", issueCount: 110, live: true},
						{key: "INFR", name: "Infrastructure", description: "Cloud accounts, networking, and capacity.", issueCount: 100, live: true},
						{key: "DATA", name: "Data Platform", description: "Warehouses, pipelines, and analytics.", issueCount: 100, live: true},
						{key: "AUTH", name: "Identity Platform", description: "SSO, directory sync, and session policy.", issueCount: 100, live: true},
						{key: "OBS", name: "Observability", description: "Metrics, traces, and alerting pipelines.", issueCount: 100, live: true},
					},
				},
				{
					name:        "Operations",
					description: "Finance-adjacent operations and internal process tooling.",
					projects: []projectSpec{
						{key: "OPS", name: "Internal Operations", description: "Runbooks and operational checklists.", issueCount: 100, live: true},
						{key: "BILL", name: "Billing Ops", description: "Invoicing, credits, and usage export.", issueCount: 100, live: true},
						{key: "PROC", name: "Procurement", description: "Vendor onboarding and purchase orders.", issueCount: 100, live: true},
						{key: "COMP", name: "Compliance", description: "Controls, audits, and evidence collection.", issueCount: 100, live: true},
						{key: "RISK", name: "Risk Register", description: "Vendor risk and operational incidents.", issueCount: 100, live: true},
					},
				},
				{
					name:        "Customer",
					description: "Implementation programs and the customer portal.",
					projects: []projectSpec{
						{key: "IMPL", name: "Implementations", description: "Paid onboarding and rollout programs.", issueCount: 100, live: true},
						{key: "PORT", name: "Customer Portal", description: "Self-service status and configuration.", issueCount: 100, live: true},
						{key: "CSAT", name: "Customer Success", description: "Health scores, renewals, and QBR prep.", issueCount: 100, live: true},
						{key: "NTBK", name: "Onboarding Kit", description: "Playbooks and kickoff templates.", issueCount: 100, live: true},
					},
				},
				{
					name:             migratedNS,
					description:      "Imported archives from retired systems and acquisitions.",
					migratedProjects: 72,
					migratedIssueMin: 75,
					migratedIssueMax: 125,
				},
			},
			teamGrants: []teamGrantSpec{
				{team: "Engineering", namespace: "Product", roleKey: "namespace-admin"},
				{team: "Engineering", namespace: "Platform", roleKey: "namespace-admin"},
				{team: "Engineering", namespace: migratedNS, roleKey: "project-viewer"},
				{team: "Design", namespace: "Product", roleKey: "namespace-admin"},
				{team: "Product", namespace: "Product", roleKey: "namespace-admin"},
				{team: "Product", namespace: "Platform", roleKey: "project-viewer"},
				{team: "Operations", namespace: "Operations", roleKey: "namespace-admin"},
				{team: "Operations", namespace: migratedNS, roleKey: "project-viewer"},
				{team: "Support", namespace: "Customer", roleKey: "namespace-admin"},
				{team: "Client Relations", namespace: "Customer", roleKey: "namespace-admin"},
				{team: "Management", namespace: "Product", roleKey: "project-viewer"},
				{team: "Management", namespace: "Platform", roleKey: "project-viewer"},
				{team: "Management", namespace: "Operations", roleKey: "project-viewer"},
				{team: "Management", namespace: "Customer", roleKey: "project-viewer"},
				{team: "Security", namespace: "Platform", roleKey: "project-viewer"},
				{team: "Sales", namespace: "Customer", roleKey: "project-viewer"},
				{team: "Marketing", namespace: "Customer", roleKey: "project-viewer"},
				{team: "Human Resources", namespace: "Operations", roleKey: "project-viewer"},
				{team: "Finance", namespace: "Operations", roleKey: "project-viewer"},
			},
		},
		partners: []orgSpec{
			partnerOrg(
				"Kite Analytics",
				"hello@kite.example",
				"kite.example",
				"https://kite.example",
				32,
				"Maya", "Chen", "maya.chen@kite.example", "Engineering Manager",
				[]teamSpec{
					{name: "Analytics Eng", description: "Data product engineering at Kite."},
					{name: "Solutions", description: "Customer-facing analytics solutions."},
				},
				[]namespaceSpec{{
					name:        "Delivery",
					description: "Analytics delivery programs for Elemo.",
					projects: []projectSpec{
						{key: "ANLT", name: "Elemo Analytics", description: "Embedded analytics for Elemo CORE.", issueCount: 100, live: true},
						{key: "PIPE", name: "Pipeline Kit", description: "Shared ingest jobs and models.", issueCount: 100, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Analytics Eng", namespace: "Delivery", roleKey: "namespace-admin"}},
			),
			partnerOrg(
				"Harbor Logistics",
				"ops@harbor.example",
				"harbor.example",
				"https://harbor.example",
				28,
				"Luis", "Navarro", "luis.navarro@harbor.example", "Operations Lead",
				[]teamSpec{
					{name: "Integrations", description: "EDI and API integrations with shippers."},
					{name: "Network Ops", description: "Yard and carrier operations."},
				},
				[]namespaceSpec{{
					name:        "Integrations",
					description: "Carrier and warehouse integrations.",
					projects: []projectSpec{
						{key: "INTG", name: "Elemo Link", description: "Shipment events into Elemo CORE.", issueCount: 100, live: true},
						{key: "YARD", name: "Yard Systems", description: "Dock scheduling and yard cameras.", issueCount: 100, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Integrations", namespace: "Integrations", roleKey: "namespace-admin"}},
			),
			partnerOrg(
				"Nimbus Cloud",
				"cloud@nimbus.example",
				"nimbus.example",
				"https://nimbus.example",
				35,
				"Priya", "Shah", "priya.shah@nimbus.example", "Cloud Architect",
				[]teamSpec{
					{name: "Cloud Operations", description: "Managed Kubernetes and account vending."},
					{name: "SRE", description: "Reliability engineering for Nimbus tenants."},
				},
				[]namespaceSpec{{
					name:        "Platform",
					description: "Managed cloud offerings used by Elemo.",
					projects: []projectSpec{
						{key: "HOST", name: "Managed Hosting", description: "Dedicated Elemo runtime cluster.", issueCount: 100, live: true},
						{key: "KUBE", name: "Kubernetes Control", description: "Cluster add-ons and policy.", issueCount: 100, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Cloud Operations", namespace: "Platform", roleKey: "namespace-admin"}},
			),
			partnerOrg(
				"Brightline Design",
				"studio@brightline.example",
				"brightline.example",
				"https://brightline.example",
				25,
				"Aisha", "Okoro", "aisha.okoro@brightline.example", "Design Director",
				[]teamSpec{
					{name: "Studio", description: "Brand and product design studio."},
					{name: "Research", description: "Customer research and usability."},
				},
				[]namespaceSpec{{
					name:        "Studio",
					description: "Retained design work for Elemo.",
					projects: []projectSpec{
						{key: "BRND", name: "Elemo Brand", description: "Brand evolution with Elemo Design.", issueCount: 100, live: true},
						{key: "UXRS", name: "UX Research", description: "Research sprints on the customer portal.", issueCount: 100, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Studio", namespace: "Studio", roleKey: "namespace-admin"}},
			),
			partnerOrg(
				"Fieldstone Consulting",
				"hello@fieldstone.example",
				"fieldstone.example",
				"https://fieldstone.example",
				40,
				"Jordan", "Lee", "jordan.lee@fieldstone.example", "Engagement Lead",
				[]teamSpec{
					{name: "Delivery", description: "Implementation consultants."},
					{name: "Enablement", description: "Training and playbooks for rollouts."},
				},
				[]namespaceSpec{{
					name:        "Engagements",
					description: "Paid implementation engagements.",
					projects: []projectSpec{
						{key: "ROLL", name: "Elemo Rollouts", description: "Joint delivery on Elemo IMPL.", issueCount: 100, live: true},
						{key: "PLAY", name: "Playbooks", description: "Reusable implementation playbooks.", issueCount: 100, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Delivery", namespace: "Engagements", roleKey: "namespace-admin"}},
			),
		},
		collaborations: []collabSpec{
			{kind: collabOrgViewer, fromOrg: "Kite Analytics", toOrg: mainOrgName, toProjectKey: "CORE"},
			{kind: collabOrgViewer, fromOrg: "Harbor Logistics", toOrg: mainOrgName, toProjectKey: "CORE"},
			{kind: collabTeamMaintainer, fromOrg: "Nimbus Cloud", fromTeam: "Cloud Operations", toOrg: mainOrgName, toProjectKey: "INFR"},
			{kind: collabOrgViewer, fromOrg: "Brightline Design", toOrg: mainOrgName, toProjectKey: "DESN"},
			{kind: collabOrgViewer, fromOrg: "Fieldstone Consulting", toOrg: mainOrgName, toProjectKey: "IMPL"},
			{kind: collabTeamViewer, fromOrg: mainOrgName, fromTeam: "Engineering", toOrg: "Kite Analytics", toProjectKey: "ANLT"},
			{kind: collabDualMember, fromOrg: mainOrgName, toOrg: "Kite Analytics", dualMemberN: 2},
			{kind: collabDualMember, fromOrg: mainOrgName, toOrg: "Harbor Logistics", dualMemberN: 1},
			{kind: collabDualMember, fromOrg: mainOrgName, toOrg: "Nimbus Cloud", dualMemberN: 1},
		},
	})
}

func smokeScenario() scenarioSpec {
	return applySlugs(scenarioSpec{
		documentCount: 20,
		main: orgSpec{
			name:       mainOrgName,
			email:      "hello@elemo.example",
			domain:     "elemo.example",
			website:    "https://elemo.example",
			userCount:  8,
			adminFirst: "Demo",
			adminLast:  "User",
			adminEmail: adminEmail,
			adminTitle: "Organization Admin",
			isMain:     true,
			teams: []teamSpec{
				{name: "Engineering", description: "Builds Elemo product."},
				{name: "Design", description: "Product design team."},
				{name: "Management", description: "Leadership team."},
			},
			namespaces: []namespaceSpec{
				{
					name:        "Product",
					description: "Customer-facing product work.",
					projects: []projectSpec{
						{key: "CORE", name: "Core Platform", description: "Shared product services and identity.", issueCount: 15, live: true},
						{key: "DESN", name: "Design System", description: "Component library and brand guidelines.", issueCount: 5, live: true},
					},
				},
				{
					name:        "Platform",
					description: "Internal developer platform.",
					projects: []projectSpec{
						{key: "INFR", name: "Infrastructure", description: "Cloud accounts, networking, and capacity.", issueCount: 5, live: true},
					},
				},
				{
					name:        "Customer",
					description: "Implementation programs.",
					projects: []projectSpec{
						{key: "IMPL", name: "Implementations", description: "Paid onboarding and rollout programs.", issueCount: 5, live: true},
					},
				},
				{
					name:             migratedNS,
					description:      "Imported archives from retired systems.",
					migratedProjects: 1,
					migratedIssueMin: 10,
					migratedIssueMax: 10,
				},
			},
			teamGrants: []teamGrantSpec{
				{team: "Engineering", namespace: "Product", roleKey: "namespace-admin"},
				{team: "Engineering", namespace: "Platform", roleKey: "namespace-admin"},
				{team: "Design", namespace: "Product", roleKey: "project-viewer"},
				{team: "Management", namespace: "Product", roleKey: "project-viewer"},
				{team: "Management", namespace: "Customer", roleKey: "project-viewer"},
			},
		},
		partners: []orgSpec{
			partnerOrg(
				"Kite Analytics",
				"hello@kite.example",
				"kite.example",
				"https://kite.example",
				3,
				"Maya", "Chen", "maya.chen@kite.example", "Engineering Manager",
				[]teamSpec{{name: "Analytics Eng", description: "Data product engineering at Kite."}},
				[]namespaceSpec{{
					name:        "Delivery",
					description: "Analytics delivery programs.",
					projects: []projectSpec{
						{key: "ANLT", name: "Elemo Analytics", description: "Embedded analytics for Elemo CORE.", issueCount: 5, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Analytics Eng", namespace: "Delivery", roleKey: "namespace-admin"}},
			),
			partnerOrg(
				"Harbor Logistics",
				"ops@harbor.example",
				"harbor.example",
				"https://harbor.example",
				3,
				"Luis", "Navarro", "luis.navarro@harbor.example", "Operations Lead",
				[]teamSpec{{name: "Integrations", description: "EDI and API integrations with shippers."}},
				[]namespaceSpec{{
					name:        "Integrations",
					description: "Carrier and warehouse integrations.",
					projects: []projectSpec{
						{key: "INTG", name: "Elemo Link", description: "Shipment events into Elemo CORE.", issueCount: 5, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Integrations", namespace: "Integrations", roleKey: "namespace-admin"}},
			),
			partnerOrg(
				"Nimbus Cloud",
				"cloud@nimbus.example",
				"nimbus.example",
				"https://nimbus.example",
				3,
				"Priya", "Shah", "priya.shah@nimbus.example", "Cloud Architect",
				[]teamSpec{{name: "Cloud Operations", description: "Managed Kubernetes and account vending."}},
				[]namespaceSpec{{
					name:        "Platform",
					description: "Managed cloud offerings used by Elemo.",
					projects: []projectSpec{
						{key: "HOST", name: "Managed Hosting", description: "Dedicated Elemo runtime cluster.", issueCount: 5, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Cloud Operations", namespace: "Platform", roleKey: "namespace-admin"}},
			),
			partnerOrg(
				"Brightline Design",
				"studio@brightline.example",
				"brightline.example",
				"https://brightline.example",
				3,
				"Aisha", "Okoro", "aisha.okoro@brightline.example", "Design Director",
				[]teamSpec{{name: "Studio", description: "Brand and product design studio."}},
				[]namespaceSpec{{
					name:        "Studio",
					description: "Retained design work for Elemo.",
					projects: []projectSpec{
						{key: "BRND", name: "Elemo Brand", description: "Brand evolution with Elemo Design.", issueCount: 5, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Studio", namespace: "Studio", roleKey: "namespace-admin"}},
			),
			partnerOrg(
				"Fieldstone Consulting",
				"hello@fieldstone.example",
				"fieldstone.example",
				"https://fieldstone.example",
				3,
				"Jordan", "Lee", "jordan.lee@fieldstone.example", "Engagement Lead",
				[]teamSpec{{name: "Delivery", description: "Implementation consultants."}},
				[]namespaceSpec{{
					name:        "Engagements",
					description: "Paid implementation engagements.",
					projects: []projectSpec{
						{key: "ROLL", name: "Elemo Rollouts", description: "Joint delivery on Elemo IMPL.", issueCount: 5, live: true},
					},
				}},
				[]teamGrantSpec{{team: "Delivery", namespace: "Engagements", roleKey: "namespace-admin"}},
			),
		},
		collaborations: []collabSpec{
			{kind: collabOrgViewer, fromOrg: "Kite Analytics", toOrg: mainOrgName, toProjectKey: "CORE"},
			{kind: collabOrgViewer, fromOrg: "Harbor Logistics", toOrg: mainOrgName, toProjectKey: "CORE"},
			{kind: collabTeamMaintainer, fromOrg: "Nimbus Cloud", fromTeam: "Cloud Operations", toOrg: mainOrgName, toProjectKey: "INFR"},
			{kind: collabOrgViewer, fromOrg: "Brightline Design", toOrg: mainOrgName, toProjectKey: "DESN"},
			{kind: collabOrgViewer, fromOrg: "Fieldstone Consulting", toOrg: mainOrgName, toProjectKey: "IMPL"},
			{kind: collabTeamViewer, fromOrg: mainOrgName, fromTeam: "Engineering", toOrg: "Kite Analytics", toProjectKey: "ANLT"},
			{kind: collabDualMember, fromOrg: mainOrgName, toOrg: "Kite Analytics", dualMemberN: 1},
			{kind: collabDualMember, fromOrg: mainOrgName, toOrg: "Harbor Logistics", dualMemberN: 1},
			{kind: collabDualMember, fromOrg: mainOrgName, toOrg: "Nimbus Cloud", dualMemberN: 1},
		},
	})
}

func partnerOrg(
	name, email, domain, website string,
	users int,
	adminFirst, adminLast, adminEmail, adminTitle string,
	teams []teamSpec,
	namespaces []namespaceSpec,
	grants []teamGrantSpec,
) orgSpec {
	return orgSpec{
		name:       name,
		email:      email,
		domain:     domain,
		website:    website,
		userCount:  users,
		adminFirst: adminFirst,
		adminLast:  adminLast,
		adminEmail: adminEmail,
		adminTitle: adminTitle,
		teams:      teams,
		namespaces: namespaces,
		teamGrants: grants,
	}
}

func (s scenarioSpec) liveProjectMinIssues() int {
	lowest := 0
	for _, org := range append([]orgSpec{s.main}, s.partners...) {
		for _, ns := range org.namespaces {
			for _, project := range ns.projects {
				if lowest == 0 || project.issueCount < lowest {
					lowest = project.issueCount
				}
			}
		}
	}
	return lowest
}

func (s scenarioSpec) liveProjectCount() int {
	n := 0
	for _, ns := range s.main.namespaces {
		if ns.name == migratedNS {
			continue
		}
		n += len(ns.projects)
	}
	return n
}

func (s scenarioSpec) migratedProjectCount() int {
	for _, ns := range s.main.namespaces {
		if ns.name == migratedNS {
			return ns.migratedProjects
		}
	}
	return 0
}

func applySlugs(spec scenarioSpec) scenarioSpec {
	applyOrg := func(org *orgSpec) {
		if org.slug == "" {
			if org.name == mainOrgName {
				org.slug = "elemo"
			} else {
				org.slug = strings.ToLower(strings.ReplaceAll(org.name, " ", "-"))
			}
		}
		for i := range org.namespaces {
			if org.namespaces[i].slug == "" {
				org.namespaces[i].slug = strings.ToLower(org.namespaces[i].name)
			}
		}
	}
	applyOrg(&spec.main)
	for i := range spec.partners {
		applyOrg(&spec.partners[i])
	}
	return spec
}
