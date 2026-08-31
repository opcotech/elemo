// ============================================================================
// Overview
//
// Demo seed for local development and UI exploration.
//
// Organizations
//   ACME Inc.   — Product (PLAT, MOB) and Operations (INF)
//   Nova Labs   — Delivery (INTEG); partner org collaborating on ACME PLAT
//
// Logins (password AppleTree123 for every account)
//   demo@elemo.example    ACME org-admin; can create orgs and install plugins
//   hector@elemo.example  ACME engineering; also Nova Labs member
//   priya@elemo.example   ACME operations
//   luis@elemo.example    ACME mobile
//   aisha@elemo.example   ACME design
//   maya@novalabs.dev     Nova Labs org-admin
//   jordan@novalabs.dev   Nova engineer; not an ACME member (PLAT via Nova)
//   sam@elemo.example     pending ACME invite (not a member yet)
//
// Authorization
//   Access is scoped ReBAC: Principal -[:GRANTED]-> scope, plus IN_SCOPE_OF.
//   There are no system roles and no wildcard actions. organization.create and
//   plugin.install are a direct grant on the Installation node for
//   demo@elemo.example (not a role, not org membership). Nova Labs holds
//   project-viewer on ACME's PLAT project; Nova members are not members of ACME.
//
// Covers: all issue kinds/statuses/priorities, common resolutions, parent
// issues, relation kinds (except reserved "depends on"), comments,
// attachments, labels, documents, teams, invitations, and todos.
//
// Requires bootstrap.cypher (Installation + constraints) to have run first.
// ============================================================================

// ============================================================================
// 1. Users
// ============================================================================

// Demo — ACME org-admin
MERGE (u:User:Principal {id: '9bsv0s46s6s002p9ltq0'})
  ON CREATE SET u += {
    username:   'demo',
    email:      'demo@elemo.example',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Demo',
    last_name:  'User',
    picture:    'https://picsum.photos/id/177/200/200.jpg',
    title:      'Senior Software Developer',
    bio:        "Hello. It's me!",
    phone:      '+12345678900',
    address:    '2900 S Congress Ave, Austin, TX',
    links:      ['https://example.com'],
    languages:  ['en'],
    created_at: datetime()
  };

// Hector — ACME engineer; also joins Nova Labs later
MERGE (u:User:Principal {id: '9bsv0s314mtg02goaimg'})
  ON CREATE SET u += {
    username:   'hector-henrik',
    email:      'hector@elemo.example',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Hector',
    last_name:  'Henrik',
    picture:    'https://picsum.photos/id/22/200/200',
    title:      'Senior Software Developer',
    bio:        'Platform engineer on the ACME product team.',
    phone:      '+12345678901',
    address:    '2900 S Congress Ave, Austin, TX',
    links:      ['https://example.com'],
    languages:  ['en', 'es'],
    created_at: datetime()
  };

// Maya — Nova Labs owner
MERGE (u:User:Principal {id: 'd9tcjmf92rs8isainivg'})
  ON CREATE SET u += {
    username:   'maya-nova',
    email:      'maya@novalabs.dev',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Maya',
    last_name:  'Chen',
    picture:    'https://picsum.photos/id/64/200/200',
    title:      'Engineering Manager',
    bio:        'Building delivery tooling at Nova Labs.',
    phone:      '+12345678902',
    address:    '500 Howard St, San Francisco, CA',
    links:      ['https://novalabs.dev'],
    languages:  ['en', 'zh'],
    created_at: datetime()
  };

// Jordan — Nova Labs engineer; guest collaborator on ACME PLAT
MERGE (u:User:Principal {id: 'd9tcjmf92rs8isainj00'})
  ON CREATE SET u += {
    username:   'jordan-lee',
    email:      'jordan@novalabs.dev',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Jordan',
    last_name:  'Lee',
    picture:    'https://picsum.photos/id/91/200/200',
    title:      'Integration Engineer',
    bio:        'Partner engineer working across ACME and Nova Labs.',
    phone:      '+12345678903',
    address:    '500 Howard St, San Francisco, CA',
    links:      ['https://novalabs.dev'],
    languages:  ['en'],
    created_at: datetime()
  };

// Sam — pending ACME invite only (no MEMBER_OF)
MERGE (u:User:Principal {id: 'd9tcjmf92rs8isainj0g'})
  ON CREATE SET u += {
    username:   'sam-rivera',
    email:      'sam@elemo.example',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Sam',
    last_name:  'Rivera',
    picture:    'https://picsum.photos/id/1011/200/200',
    title:      'Product Designer',
    bio:        'Pending invite to ACME.',
    phone:      '+12345678904',
    address:    '1 Market St, San Francisco, CA',
    links:      ['https://example.com'],
    languages:  ['en'],
    created_at: datetime()
  };

// Priya — ACME SRE
MERGE (u:User:Principal {id: 'd9tcjmf92rs8isainm00'})
  ON CREATE SET u += {
    username:   'priya-shah',
    email:      'priya@elemo.example',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Priya',
    last_name:  'Shah',
    picture:    'https://picsum.photos/id/65/200/200',
    title:      'Site Reliability Engineer',
    bio:        'Owns ACME infrastructure and incident follow-up.',
    phone:      '+12345678905',
    address:    '2900 S Congress Ave, Austin, TX',
    links:      ['https://example.com'],
    languages:  ['en', 'hi'],
    created_at: datetime()
  };

// Luis — ACME mobile
MERGE (u:User:Principal {id: 'd9tcjmf92rs8isainm0g'})
  ON CREATE SET u += {
    username:   'luis-ortega',
    email:      'luis@elemo.example',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Luis',
    last_name:  'Ortega',
    picture:    'https://picsum.photos/id/1005/200/200',
    title:      'Mobile Engineer',
    bio:        'Native clients for ACME customers.',
    phone:      '+12345678906',
    address:    '2900 S Congress Ave, Austin, TX',
    links:      ['https://example.com'],
    languages:  ['en', 'es'],
    created_at: datetime()
  };

// Aisha — ACME design (read on Product)
MERGE (u:User:Principal {id: 'd9tcjmf92rs8isainm10'})
  ON CREATE SET u += {
    username:   'aisha-khan',
    email:      'aisha@elemo.example',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Aisha',
    last_name:  'Khan',
    picture:    'https://picsum.photos/id/1027/200/200',
    title:      'Product Designer',
    bio:        'Design systems and mobile UX.',
    phone:      '+12345678907',
    address:    '1 Market St, San Francisco, CA',
    links:      ['https://example.com'],
    languages:  ['en', 'ar'],
    created_at: datetime()
  };

// Riley — inactive ACME member
MERGE (u:User:Principal {id: 'd9tcjmf92rs8isainm1g'})
  ON CREATE SET u += {
    username:   'riley-nash',
    email:      'riley@elemo.example',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'inactive',
    first_name: 'Riley',
    last_name:  'Nash',
    picture:    'https://picsum.photos/id/338/200/200',
    title:      'Contractor',
    bio:        'Previously helped with platform migrations.',
    phone:      '+12345678908',
    address:    '2900 S Congress Ave, Austin, TX',
    links:      [],
    languages:  ['en'],
    created_at: datetime() - duration('P90D')
  };

// ============================================================================
// 2. Organizations, memberships, and invitations
// ============================================================================

// ACME Inc. — demo is org-admin (granted in section 14)
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (o:Organization:Principal {id: '9bsv0s4vl6gg02sv7jrg'})
  ON CREATE SET o += {
    slug:       'acme',
    name:       'ACME Inc.',
    email:      'info@example.com',
    logo:       'https://picsum.photos/id/211/200/200.jpg',
    website:    'https://example.com',
    status:     'active',
    created_at: datetime()
  }
CREATE
  (u)-[:MEMBER_OF {id: '9m4e2mr0ui3e8a215n4g', created_at: datetime()}]->(o);

// Hector → ACME member
MATCH (u:User {id: '9bsv0s314mtg02goaimg'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: '9bsv0s314mtg02goain0', created_at: datetime()}]->(o);

// Priya → ACME member
MATCH (u:User {id: 'd9tcjmf92rs8isainm00'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm20', created_at: datetime()}]->(o);

// Luis → ACME member
MATCH (u:User {id: 'd9tcjmf92rs8isainm0g'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm30', created_at: datetime()}]->(o);

// Aisha → ACME member
MATCH (u:User {id: 'd9tcjmf92rs8isainm10'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm40', created_at: datetime()}]->(o);

// Riley → ACME member (inactive account)
MATCH (u:User {id: 'd9tcjmf92rs8isainm1g'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm50', created_at: datetime() - duration('P90D')}]->(o);

// Sam → ACME pending invite
MATCH (u:User {id: 'd9tcjmf92rs8isainj0g'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE (u)-[:INVITED_TO {id: 'd9tcjmf92rs8isainj4g', created_at: datetime()}]->(o);

// Nova Labs — Maya is org-admin (granted in section 14)
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MERGE (o:Organization:Principal {id: 'd9tcjmf92rs8isainj10'})
  ON CREATE SET o += {
    slug:       'nova-labs',
    name:       'Nova Labs',
    email:      'hello@novalabs.dev',
    logo:       'https://picsum.photos/id/180/200/200.jpg',
    website:    'https://novalabs.dev',
    status:     'active',
    created_at: datetime()
  }
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainj1g', created_at: datetime()}]->(o);

// Jordan → Nova Labs member
MATCH (u:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (o:Organization {id: 'd9tcjmf92rs8isainj10'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainj2g', created_at: datetime()}]->(o);

// Hector → Nova Labs member — multi-org membership
MATCH (u:User {id: '9bsv0s314mtg02goaimg'})
MATCH (o:Organization {id: 'd9tcjmf92rs8isainj10'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainj3g', created_at: datetime()}]->(o);

// ============================================================================
// 3. Namespaces and projects
// ============================================================================

// ACME → Product namespace
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MERGE (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
  ON CREATE SET ns += {
    slug:        'product',
    organization_id: '9bsv0s4vl6gg02sv7jrg',
    name:        'Product',
    description: 'Product engineering workspaces for ACME platforms and apps.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_NAMESPACE {id: 'd9tcjmf92rs8isainj5g', created_at: datetime()}]->(ns);

// ACME → Operations namespace
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MERGE (ns:Namespace {id: 'd9tcjmf92rs8isainj6g'})
  ON CREATE SET ns += {
    slug:        'operations',
    organization_id: '9bsv0s4vl6gg02sv7jrg',
    name:        'Operations',
    description: 'Infrastructure and internal operations for ACME.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_NAMESPACE {id: 'd9tcjmf92rs8isainj70', created_at: datetime()}]->(ns);

// Nova Labs → Delivery namespace
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MATCH (o:Organization {id: 'd9tcjmf92rs8isainj10'})
MERGE (ns:Namespace {id: 'd9tcjmf92rs8isainj80'})
  ON CREATE SET ns += {
    slug:        'delivery',
    organization_id: 'd9tcjmf92rs8isainj10',
    name:        'Delivery',
    description: 'Client delivery and partner integration projects.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_NAMESPACE {id: 'd9tcjmf92rs8isainj8g', created_at: datetime()}]->(ns);

// ACME Product → PLAT project
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainj9g'})
  ON CREATE SET p += {
    key:           'PLAT',
    namespace_id:  'd9tcjmf92rs8isainj50',
    name:          'Elemo Platform',
    description:   'Core product platform shared with partner teams.',
    logo:          'https://picsum.photos/id/201/200/200.jpg',
    status:        'active',
    next_issue_id: 10,
    created_at:    datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p);

// ACME Product → MOB project
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjag'})
  ON CREATE SET p += {
    key:           'MOB',
    namespace_id:  'd9tcjmf92rs8isainj50',
    name:          'Mobile App',
    description:   'Native mobile clients for ACME customers.',
    logo:          '',
    status:        'active',
    next_issue_id: 4,
    created_at:    datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p);

// ACME Operations → INF project
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj6g'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjbg'})
  ON CREATE SET p += {
    key:           'INF',
    namespace_id:  'd9tcjmf92rs8isainj6g',
    name:          'Infrastructure',
    description:   'Cloud infrastructure and reliability workstreams.',
    logo:          '',
    status:        'active',
    next_issue_id: 6,
    created_at:    datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p);

// Nova Delivery → INTEG project
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj80'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjcg'})
  ON CREATE SET p += {
    key:           'INTEG',
    namespace_id:  'd9tcjmf92rs8isainj80',
    name:          'ACME Integration',
    description:   'Partner delivery project for ACME platform integrations.',
    logo:          'https://picsum.photos/id/250/200/200.jpg',
    status:        'active',
    next_issue_id: 4,
    created_at:    datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p);

// ============================================================================
// 4. Teams
// ============================================================================

// ACME org team: Engineering (demo, hector, luis); namespace-admin on Product
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
MATCH (luis:User {id: 'd9tcjmf92rs8isainm0g'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MERGE (t:Team:Principal {id: 'd9tcjmf92rs8isainjdg'})
  ON CREATE SET t += {
    name:        'Engineering',
    description: 'ACME product engineering team.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_TEAM {id: 'd9tcjmf92rs8isainje0', created_at: datetime()}]->(t),
  (t)-[:IN_SCOPE_OF {id: 'd9tcjmf92rs8isait10', created_at: datetime()}]->(o),
  (owner)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainjeg', created_at: datetime()}]->(t),
  (hector)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainjfg', created_at: datetime()}]->(t),
  (luis)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm60', created_at: datetime()}]->(t);

// ACME org team: SRE (demo, priya); namespace-admin on Operations
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MERGE (t:Team:Principal {id: 'd9tcjmf92rs8isainm70'})
  ON CREATE SET t += {
    name:        'SRE',
    description: 'ACME site reliability and infrastructure.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_TEAM {id: 'd9tcjmf92rs8isainm7g', created_at: datetime()}]->(t),
  (t)-[:IN_SCOPE_OF {id: 'd9tcjmf92rs8isait20', created_at: datetime()}]->(o),
  (owner)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm80', created_at: datetime()}]->(t),
  (priya)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm90', created_at: datetime()}]->(t);

// ACME org team: Design (demo, aisha); project-viewer on Product
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (aisha:User {id: 'd9tcjmf92rs8isainm10'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MERGE (t:Team:Principal {id: 'd9tcjmf92rs8isainmag'})
  ON CREATE SET t += {
    name:        'Design',
    description: 'ACME product design.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_TEAM {id: 'd9tcjmf92rs8isainmb0', created_at: datetime()}]->(t),
  (t)-[:IN_SCOPE_OF {id: 'd9tcjmf92rs8isait30', created_at: datetime()}]->(o),
  (owner)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainmbg', created_at: datetime()}]->(t),
  (aisha)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainmcg', created_at: datetime()}]->(t);

// ACME PLAT project team: Contractors (demo + jordan); project-maintainer on PLAT
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MERGE (t:Team:Principal {id: 'd9tcjmf92rs8isainjgg'})
  ON CREATE SET t += {
    name:        'Contractors',
    description: 'External partners collaborating on the platform project.',
    created_at:  datetime()
  }
CREATE
  (p)-[:HAS_TEAM {id: 'd9tcjmf92rs8isainjh0', created_at: datetime()}]->(t),
  (t)-[:IN_SCOPE_OF {id: 'd9tcjmf92rs8isait40', created_at: datetime()}]->(p),
  (owner)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainjhg', created_at: datetime()}]->(t),
  (jordan)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainjig', created_at: datetime()}]->(t);

// ============================================================================
// 6. Labels
// ============================================================================

MERGE (l:Label {id: 'd9tcjmf92rs8isainjkg'})
  ON CREATE SET l += { name: 'bug', description: 'Defects and regressions.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainjl0'})
  ON CREATE SET l += { name: 'feature', description: 'New product capabilities.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainjlg'})
  ON CREATE SET l += { name: 'partner', description: 'Work involving partner organizations.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainmgg'})
  ON CREATE SET l += { name: 'frontend', description: 'Web client and work-surface UI.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainmh0'})
  ON CREATE SET l += { name: 'accessibility', description: 'Keyboard, contrast, and assistive tech.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainmhg'})
  ON CREATE SET l += { name: 'documentation', description: 'Guides, runbooks, and API notes.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainmi0'})
  ON CREATE SET l += { name: 'reliability', description: 'Availability, latency, and incident work.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainmig'})
  ON CREATE SET l += { name: 'api', description: 'HTTP API and integration contracts.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainmj0'})
  ON CREATE SET l += { name: 'design', description: 'UX and visual design.', created_at: datetime() };

MERGE (l:Label {id: 'd9tcjmf92rs8isainmjg'})
  ON CREATE SET l += { name: 'security', description: 'Auth, secrets, and access control.', created_at: datetime() };

// ============================================================================
// 7. Documents and folders
// ============================================================================

// Product library folders: Guides / Architecture
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (guides:Folder {id: 'd9tcjmf92rs8isainq00'})
  ON CREATE SET guides += {
    name:       'Guides',
    created_by: '9bsv0s46s6s002p9ltq0',
    created_at: datetime() - duration('P21D')
  }
MERGE (arch:Folder {id: 'd9tcjmf92rs8isainq10'})
  ON CREATE SET arch += {
    name:       'Architecture',
    created_by: '9bsv0s46s6s002p9ltq0',
    created_at: datetime() - duration('P21D')
  }
CREATE
  (guides)-[:SCOPED_TO {id: 'd9tcjmf92rs8isainq20', created_at: datetime() - duration('P21D')}]->(ns),
  (arch)-[:SCOPED_TO {id: 'd9tcjmf92rs8isainq30', created_at: datetime() - duration('P21D')}]->(ns),
  (arch)-[:LOCATED_IN {id: 'd9tcjmf92rs8isainq40', created_at: datetime() - duration('P21D')}]->(guides),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainq50', created_at: datetime() - duration('P21D')}]->(guides),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainq60', created_at: datetime() - duration('P21D')}]->(arch);

// Product handbook (Product namespace root)
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (d:Document {id: 'd9tcjmf92rs8isainjm0'})
  ON CREATE SET d += {
    title:      'Product Handbook',
    excerpt:    'Orientation guide for ACME product engineering and partners.',
    file_id:    'demo/product-handbook.md',
    created_by: '9bsv0s46s6s002p9ltq0',
    created_at: datetime() - duration('P20D')
  }
CREATE
  (d)-[:SCOPED_TO {id: 'd9tcjmf92rs8isainjmg', created_at: datetime() - duration('P20D')}]->(ns),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjn0', created_at: datetime() - duration('P20D')}]->(d);

MATCH (d:Document {id: 'd9tcjmf92rs8isainjm0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmhg'})
CREATE (d)-[:HAS_LABEL]->(l);

// Platform overview (Product library / Architecture, related to PLAT)
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (arch:Folder {id: 'd9tcjmf92rs8isainq10'})
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (d:Document {id: 'd9tcjmf92rs8isainjng'})
  ON CREATE SET d += {
    title:      'Platform Overview',
    excerpt:    'Architecture notes and onboarding for the Elemo platform project.',
    file_id:    'demo/platform-overview.md',
    created_by: '9bsv0s46s6s002p9ltq0',
    created_at: datetime() - duration('P14D')
  }
CREATE
  (d)-[:SCOPED_TO {id: 'd9tcjmf92rs8isainjo0', created_at: datetime() - duration('P14D')}]->(ns),
  (d)-[:LOCATED_IN {id: 'd9tcjmf92rs8isainq80', created_at: datetime() - duration('P14D')}]->(arch),
  (d)-[:RELATED_TO {id: 'd9tcjmf92rs8isainq90', created_at: datetime() - duration('P14D')}]->(p),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjog', created_at: datetime() - duration('P14D')}]->(d);

MATCH (d:Document {id: 'd9tcjmf92rs8isainjng'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmig'})
CREATE (d)-[:HAS_LABEL]->(l);

// Integration guide (Delivery library root, related to INTEG)
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj80'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MERGE (d:Document {id: 'd9tcjmf92rs8isainjp0'})
  ON CREATE SET d += {
    title:      'Integration Guide',
    excerpt:    'How Nova Labs integrates with ACME platform APIs and workflows.',
    file_id:    'demo/integration-guide.md',
    created_by: 'd9tcjmf92rs8isainivg',
    created_at: datetime() - duration('P10D')
  }
CREATE
  (d)-[:SCOPED_TO {id: 'd9tcjmf92rs8isainjpg', created_at: datetime() - duration('P10D')}]->(ns),
  (d)-[:RELATED_TO {id: 'd9tcjmf92rs8isainqb0', created_at: datetime() - duration('P10D')}]->(p),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjq0', created_at: datetime() - duration('P10D')}]->(d);

MATCH (d:Document {id: 'd9tcjmf92rs8isainjp0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjlg'})
CREATE (d)-[:HAS_LABEL]->(l);

// Operations runbook (Operations namespace root)
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj6g'})
MATCH (u:User {id: 'd9tcjmf92rs8isainm00'})
MERGE (d:Document {id: 'd9tcjmf92rs8isainmk0'})
  ON CREATE SET d += {
    title:      'Operations Runbook',
    excerpt:    'Paging, deploy freezes, and guest-collaborator secret rotation.',
    file_id:    'demo/operations-runbook.md',
    created_by: 'd9tcjmf92rs8isainm00',
    created_at: datetime() - duration('P8D')
  }
CREATE
  (d)-[:SCOPED_TO {id: 'd9tcjmf92rs8isainmkg', created_at: datetime() - duration('P8D')}]->(ns),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainml0', created_at: datetime() - duration('P8D')}]->(d);

MATCH (d:Document {id: 'd9tcjmf92rs8isainmk0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmi0'})
CREATE (d)-[:HAS_LABEL]->(l);

// Mobile design notes (Product library / Guides, related to MOB)
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainjag'})
MATCH (guides:Folder {id: 'd9tcjmf92rs8isainq00'})
MATCH (u:User {id: 'd9tcjmf92rs8isainm10'})
MERGE (d:Document {id: 'd9tcjmf92rs8isainmlg'})
  ON CREATE SET d += {
    title:      'Mobile Design Notes',
    excerpt:    'Navigation, offline states, and biometric unlock patterns.',
    file_id:    'demo/mobile-design-notes.md',
    created_by: 'd9tcjmf92rs8isainm10',
    created_at: datetime() - duration('P6D')
  }
CREATE
  (d)-[:SCOPED_TO {id: 'd9tcjmf92rs8isainmm0', created_at: datetime() - duration('P6D')}]->(ns),
  (d)-[:LOCATED_IN {id: 'd9tcjmf92rs8isainqe0', created_at: datetime() - duration('P6D')}]->(guides),
  (d)-[:RELATED_TO {id: 'd9tcjmf92rs8isainqf0', created_at: datetime() - duration('P6D')}]->(p),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainmmg', created_at: datetime() - duration('P6D')}]->(d);

MATCH (d:Document {id: 'd9tcjmf92rs8isainmlg'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmj0'})
CREATE (d)-[:HAS_LABEL]->(l);

// ============================================================================
// 8. Issues — PLAT (Elemo Platform)
// ============================================================================

// PLAT-1 epic — Partner collaboration rollout
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainjqg'})
  ON CREATE SET i += {
    numeric_id:  1,
    kind:        'epic',
    title:       'Partner collaboration rollout',
    description: 'Enable Nova Labs engineers to collaborate on ACME platform work.',
    status:      'in progress',
    priority:    'high',
    resolution:  'none',
    links:       ['https://example.com/epics/partner-rollout\tPartner rollout epic'],
    due_date:    datetime() + duration('P14D'),
    start_date:  datetime() - duration('P10D'),
    created_at:  datetime() - duration('P18D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainjr0', created_at: datetime() - duration('P18D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainjrg', created_at: datetime() - duration('P18D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainjs0', created_at: datetime() - duration('P18D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainjqg'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjlg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainjqg'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
CREATE
  (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainjsg', kind: 'assignee', created_at: datetime() - duration('P16D')}]->(i),
  (hector)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainjt0', kind: 'reviewer', created_at: datetime() - duration('P16D')}]->(i),
  (jordan)-[:WATCHES {id: 'd9tcjmf92rs8isainjtg', created_at: datetime() - duration('P16D')}]->(i);

// PLAT-2 story — Shared issue board for partners (subtask of epic)
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (parent:Issue {id: 'd9tcjmf92rs8isainjqg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainju0'})
  ON CREATE SET i += {
    numeric_id:  2,
    kind:        'story',
    title:       'Shared issue board for partners',
    description: 'Expose partner-visible issues and assignment workflows on PLAT.',
    status:      'open',
    priority:    'normal',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P10D'),
    start_date:  null,
    created_at:  datetime() - duration('P12D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainjug', created_at: datetime() - duration('P12D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainjv0', created_at: datetime() - duration('P12D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainjvg', created_at: datetime() - duration('P12D')}]->(p),
  (i)-[:RELATED_TO {id: 'd9tcjmf92rs8isaink00', kind: 'subtask of', created_at: datetime() - duration('P12D')}]->(parent);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainju0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjl0'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainju0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmgg'})
CREATE (i)-[:HAS_LABEL]->(l);

// PLAT-3 bug — Auth redirect for guest collaborators
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: '9bsv0s314mtg02goaimg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
  ON CREATE SET i += {
    numeric_id:  3,
    kind:        'bug',
    title:       'Auth redirect for guest collaborators',
    description: 'Guest users with project write permission hit an unexpected org redirect.',
    status:      'blocked',
    priority:    'highest',
    resolution:  'none',
    links:       ['https://example.com/incidents/guest-auth\tGuest auth incident'],
    due_date:    datetime() - duration('P2D'),
    start_date:  datetime() - duration('P6D'),
    created_at:  datetime() - duration('P8D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isaink10', created_at: datetime() - duration('P8D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isaink1g', created_at: datetime() - duration('P8D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaink20', created_at: datetime() - duration('P8D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjkg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmjg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (bug:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (story:Issue {id: 'd9tcjmf92rs8isainju0'})
CREATE (bug)-[:RELATED_TO {id: 'd9tcjmf92rs8isaink2g', kind: 'blocks', created_at: datetime() - duration('P7D')}]->(story);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
CREATE
  (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaink30', kind: 'assignee', created_at: datetime() - duration('P7D')}]->(i);

// PLAT-4 task — Keyboard-first quick create (subtask of epic)
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (parent:Issue {id: 'd9tcjmf92rs8isainjqg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainmq0'})
  ON CREATE SET i += {
    numeric_id:  4,
    kind:        'task',
    title:       'Keyboard-first quick create',
    description: 'Create work without leaving the active namespace and project.',
    status:      'review',
    priority:    'high',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P1D'),
    start_date:  datetime() - duration('P4D'),
    created_at:  datetime() - duration('P9D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainmqg', created_at: datetime() - duration('P9D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainmr0', created_at: datetime() - duration('P9D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainms0', created_at: datetime() - duration('P9D')}]->(p),
  (i)-[:RELATED_TO {id: 'd9tcjmf92rs8isainmsg', kind: 'subtask of', created_at: datetime() - duration('P9D')}]->(parent);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainmq0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmgg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainmq0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmh0'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainmq0'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
MATCH (aisha:User {id: 'd9tcjmf92rs8isainm10'})
CREATE
  (hector)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainmt0', kind: 'assignee', created_at: datetime() - duration('P5D')}]->(i),
  (aisha)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainmtg', kind: 'reviewer', created_at: datetime() - duration('P2D')}]->(i);

// PLAT-5 story — Preserve work context across projections
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainmv0'})
  ON CREATE SET i += {
    numeric_id:  5,
    kind:        'story',
    title:       'Preserve work context across projections',
    description: 'Keep filters and the selected inspector item stable while switching views.',
    status:      'in progress',
    priority:    'highest',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P3D'),
    start_date:  datetime() - duration('P3D'),
    created_at:  datetime() - duration('P11D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainmvg', created_at: datetime() - duration('P11D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainn00', created_at: datetime() - duration('P11D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainn10', created_at: datetime() - duration('P11D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainmv0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmgg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainmv0'})
MATCH (demo:User {id: '9bsv0s46s6s002p9ltq0'})
CREATE (demo)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainn1g', kind: 'assignee', created_at: datetime() - duration('P3D')}]->(i);

// PLAT-6 task — Document saved-view query semantics (unassigned backlog)
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm10'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainn20'})
  ON CREATE SET i += {
    numeric_id:  6,
    kind:        'task',
    title:       'Document saved-view query semantics',
    description: 'Describe portable filters, grouping, sorting, and display choices.',
    status:      'open',
    priority:    'low',
    resolution:  'none',
    links:       [],
    due_date:    null,
    start_date:  null,
    created_at:  datetime() - duration('P5D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainn2g', created_at: datetime() - duration('P5D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainn30', created_at: datetime() - duration('P5D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainn40', created_at: datetime() - duration('P5D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainn20'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmhg'})
CREATE (i)-[:HAS_LABEL]->(l);

// PLAT-7 bug — Closed, fixed
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainj00'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainn4g'})
  ON CREATE SET i += {
    numeric_id:  7,
    kind:        'bug',
    title:       'Webhook secret echoed in debug logs',
    description: 'Partner callback secrets appeared in verbose request logs.',
    status:      'closed',
    priority:    'high',
    resolution:  'fixed',
    links:       ['https://example.com/incidents/webhook-secret\tWebhook secret incident'],
    due_date:    datetime() - duration('P4D'),
    start_date:  datetime() - duration('P9D'),
    created_at:  datetime() - duration('P12D'),
    updated_at:  datetime() - duration('P3D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainn50', created_at: datetime() - duration('P12D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainn5g', created_at: datetime() - duration('P12D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainn6g', created_at: datetime() - duration('P12D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainn4g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjkg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainn4g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmjg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainn4g'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
CREATE
  (hector)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainn70', kind: 'assignee', created_at: datetime() - duration('P10D')}]->(i);

// PLAT-8 task — Done, fixed
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainn80'})
  ON CREATE SET i += {
    numeric_id:  8,
    kind:        'task',
    title:       'Spike: SSO for partner orgs',
    description: 'Time-boxed investigation of SSO for Nova Labs members on PLAT.',
    status:      'done',
    priority:    'normal',
    resolution:  'fixed',
    links:       [],
    due_date:    datetime() - duration('P1D'),
    start_date:  datetime() - duration('P7D'),
    created_at:  datetime() - duration('P10D'),
    updated_at:  datetime() - duration('P1D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainn8g', created_at: datetime() - duration('P10D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainn90', created_at: datetime() - duration('P10D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainna0', created_at: datetime() - duration('P10D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainn80'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjlg'})
CREATE (i)-[:HAS_LABEL]->(l);

// PLAT-9 bug — Closed, cannot reproduce
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm10'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainnag'})
  ON CREATE SET i += {
    numeric_id:  9,
    kind:        'bug',
    title:       'Empty board flash on first load',
    description: 'Board briefly rendered empty before issues appeared. Could not reproduce after cache fix.',
    status:      'closed',
    priority:    'lowest',
    resolution:  'cannot reproduce',
    links:       [],
    due_date:    null,
    start_date:  null,
    created_at:  datetime() - duration('P15D'),
    updated_at:  datetime() - duration('P6D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainnb0', created_at: datetime() - duration('P15D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainnbg', created_at: datetime() - duration('P15D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainncg', created_at: datetime() - duration('P15D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnag'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjkg'})
CREATE (i)-[:HAS_LABEL]->(l);

// PLAT-10 story — related to Nova INTEG work
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: '9bsv0s314mtg02goaimg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainnd0'})
  ON CREATE SET i += {
    numeric_id:  10,
    kind:        'story',
    title:       'Publish partner-visible status mapping',
    description: 'Document how ACME issue statuses appear on the Nova delivery board.',
    status:      'open',
    priority:    'normal',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P8D'),
    start_date:  null,
    created_at:  datetime() - duration('P4D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainndg', created_at: datetime() - duration('P4D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainne0', created_at: datetime() - duration('P4D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainnf0', created_at: datetime() - duration('P4D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnd0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjlg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnd0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmhg'})
CREATE (i)-[:HAS_LABEL]->(l);

// ============================================================================
// 9. Issues — MOB (Mobile App)
// ============================================================================

// MOB-1 epic
MATCH (p:Project {id: 'd9tcjmf92rs8isainjag'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm0g'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainnfg'})
  ON CREATE SET i += {
    numeric_id:  1,
    kind:        'epic',
    title:       'Offline-first mobile shell',
    description: 'Cache the work board and queue mutations while the device is offline.',
    status:      'open',
    priority:    'high',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P21D'),
    start_date:  null,
    created_at:  datetime() - duration('P7D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainng0', created_at: datetime() - duration('P7D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainngg', created_at: datetime() - duration('P7D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainnhg', created_at: datetime() - duration('P7D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnfg'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjl0'})
CREATE (i)-[:HAS_LABEL]->(l);

// MOB-2 story — subtask of MOB-1
MATCH (p:Project {id: 'd9tcjmf92rs8isainjag'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm10'})
MATCH (parent:Issue {id: 'd9tcjmf92rs8isainnfg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainni0'})
  ON CREATE SET i += {
    numeric_id:  2,
    kind:        'story',
    title:       'Push notification settings',
    description: 'Let people choose assignment, mention, and due-date alerts on iOS and Android.',
    status:      'in progress',
    priority:    'normal',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P6D'),
    start_date:  datetime() - duration('P2D'),
    created_at:  datetime() - duration('P6D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainnig', created_at: datetime() - duration('P6D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainnj0', created_at: datetime() - duration('P6D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainnk0', created_at: datetime() - duration('P6D')}]->(p),
  (i)-[:RELATED_TO {id: 'd9tcjmf92rs8isainnkg', kind: 'subtask of', created_at: datetime() - duration('P6D')}]->(parent);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainni0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmj0'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainni0'})
MATCH (luis:User {id: 'd9tcjmf92rs8isainm0g'})
CREATE
  (luis)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainnl0', kind: 'assignee', created_at: datetime() - duration('P2D')}]->(i);

// MOB-3 bug
MATCH (p:Project {id: 'd9tcjmf92rs8isainjag'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm0g'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainnm0'})
  ON CREATE SET i += {
    numeric_id:  3,
    kind:        'bug',
    title:       'iOS login keyboard overlaps the submit button',
    description: 'On smaller iPhones the keyboard covers the primary action after focusing the password field.',
    status:      'open',
    priority:    'highest',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P2D'),
    start_date:  null,
    created_at:  datetime() - duration('P3D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainnmg', created_at: datetime() - duration('P3D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainnn0', created_at: datetime() - duration('P3D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainno0', created_at: datetime() - duration('P3D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnm0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjkg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnm0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmh0'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (bug:Issue {id: 'd9tcjmf92rs8isainnm0'})
MATCH (task:Issue {id: 'd9tcjmf92rs8isainmq0'})
CREATE (bug)-[:RELATED_TO {id: 'd9tcjmf92rs8isainnog', kind: 'related to', created_at: datetime() - duration('P3D')}]->(task);

// MOB-4 task — Done
MATCH (p:Project {id: 'd9tcjmf92rs8isainjag'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm0g'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainnp0'})
  ON CREATE SET i += {
    numeric_id:  4,
    kind:        'task',
    title:       'Android biometric unlock',
    description: 'Unlock the session with the device biometric after a successful password login.',
    status:      'done',
    priority:    'low',
    resolution:  'fixed',
    links:       [],
    due_date:    datetime() - duration('P1D'),
    start_date:  datetime() - duration('P8D'),
    created_at:  datetime() - duration('P9D'),
    updated_at:  datetime() - duration('P1D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainnpg', created_at: datetime() - duration('P9D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainnq0', created_at: datetime() - duration('P9D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainnr0', created_at: datetime() - duration('P9D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnp0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjl0'})
CREATE (i)-[:HAS_LABEL]->(l);

// ============================================================================
// 10. Issues — INF (Infrastructure)
// ============================================================================

// INF-1 task — overdue, in progress
MATCH (p:Project {id: 'd9tcjmf92rs8isainjbg'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm00'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainnrg'})
  ON CREATE SET i += {
    numeric_id:  1,
    kind:        'task',
    title:       'Review delayed notification deliveries',
    description: 'Correlate delivery latency with the most recent deployment.',
    status:      'in progress',
    priority:    'highest',
    resolution:  'none',
    links:       ['https://example.com/dashboards/notifications\tNotification dashboard'],
    due_date:    datetime() - duration('P1D'),
    start_date:  datetime() - duration('P4D'),
    created_at:  datetime() - duration('P5D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainns0', created_at: datetime() - duration('P5D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainnsg', created_at: datetime() - duration('P5D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainntg', created_at: datetime() - duration('P5D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnrg'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmi0'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnrg'})
MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
CREATE (priya)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainnu0', kind: 'assignee', created_at: datetime() - duration('P4D')}]->(i);

// INF-2 story — done
MATCH (p:Project {id: 'd9tcjmf92rs8isainjbg'})
MATCH (reporter:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainnug'})
  ON CREATE SET i += {
    numeric_id:  2,
    kind:        'story',
    title:       'Publish incident follow-up',
    description: 'Capture owners and due dates for the last guest-auth incident.',
    status:      'done',
    priority:    'low',
    resolution:  'fixed',
    links:       [],
    due_date:    datetime() - duration('P3D'),
    start_date:  datetime() - duration('P8D'),
    created_at:  datetime() - duration('P9D'),
    updated_at:  datetime() - duration('P3D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainnv0', created_at: datetime() - duration('P9D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainnvg', created_at: datetime() - duration('P9D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaino0g', created_at: datetime() - duration('P9D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnug'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmhg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnug'})
MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
CREATE
  (priya)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaino10', kind: 'assignee', created_at: datetime() - duration('P8D')}]->(i);

// INF-3 bug — blocked
MATCH (p:Project {id: 'd9tcjmf92rs8isainjbg'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm00'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isaino20'})
  ON CREATE SET i += {
    numeric_id:  3,
    kind:        'bug',
    title:       'Staging Redis memory spike',
    description: 'Staging Redis hit eviction during the last load test and dropped session keys.',
    status:      'blocked',
    priority:    'high',
    resolution:  'none',
    links:       ['https://example.com/metrics/redis-staging\tRedis staging metrics'],
    due_date:    datetime() + duration('P4D'),
    start_date:  datetime() - duration('P2D'),
    created_at:  datetime() - duration('P3D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isaino2g', created_at: datetime() - duration('P3D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isaino30', created_at: datetime() - duration('P3D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaino40', created_at: datetime() - duration('P3D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaino20'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjkg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaino20'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmi0'})
CREATE (i)-[:HAS_LABEL]->(l);

// INF-4 task — open
MATCH (p:Project {id: 'd9tcjmf92rs8isainjbg'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm00'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isaino4g'})
  ON CREATE SET i += {
    numeric_id:  4,
    kind:        'task',
    title:       'Rotate guest collaborator secrets',
    description: 'Replace the shared partner credentials used by Jordan on PLAT.',
    status:      'open',
    priority:    'normal',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P7D'),
    start_date:  null,
    created_at:  datetime() - duration('P2D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isaino50', created_at: datetime() - duration('P2D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isaino5g', created_at: datetime() - duration('P2D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaino6g', created_at: datetime() - duration('P2D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaino4g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmjg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (secret:Issue {id: 'd9tcjmf92rs8isaino4g'})
MATCH (auth:Issue {id: 'd9tcjmf92rs8isaink0g'})
CREATE (secret)-[:RELATED_TO {id: 'd9tcjmf92rs8isaino70', kind: 'related to', created_at: datetime() - duration('P2D')}]->(auth);

// INF-5 bug — closed duplicate of INF-3
MATCH (p:Project {id: 'd9tcjmf92rs8isainjbg'})
MATCH (reporter:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (canonical:Issue {id: 'd9tcjmf92rs8isaino20'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isaino7g'})
  ON CREATE SET i += {
    numeric_id:  5,
    kind:        'bug',
    title:       'Staging cache misses after load test',
    description: 'Follow-up report of the same Redis eviction. Closed as a duplicate of INF-3.',
    status:      'closed',
    priority:    'normal',
    resolution:  'duplicate',
    links:       [],
    due_date:    null,
    start_date:  null,
    created_at:  datetime() - duration('P2D'),
    updated_at:  datetime() - duration('P1D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isaino80', created_at: datetime() - duration('P2D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isaino8g', created_at: datetime() - duration('P2D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaino9g', created_at: datetime() - duration('P2D')}]->(p),
  (i)-[:RELATED_TO {id: 'd9tcjmf92rs8isainoa0', kind: 'duplicates', created_at: datetime() - duration('P1D')}]->(canonical);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaino7g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjkg'})
CREATE (i)-[:HAS_LABEL]->(l);

// INF-6 task — closed won't fix
MATCH (p:Project {id: 'd9tcjmf92rs8isainjbg'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainm00'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainoag'})
  ON CREATE SET i += {
    numeric_id:  6,
    kind:        'task',
    title:       'Custom pager for staging only',
    description: 'Separate staging paging rotation. Deferred; staging stays on the shared SRE rotation.',
    status:      'closed',
    priority:    'lowest',
    resolution:  "won't fix",
    links:       [],
    due_date:    null,
    start_date:  null,
    created_at:  datetime() - duration('P16D'),
    updated_at:  datetime() - duration('P10D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainob0', created_at: datetime() - duration('P16D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainobg', created_at: datetime() - duration('P16D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainocg', created_at: datetime() - duration('P16D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainoag'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmi0'})
CREATE (i)-[:HAS_LABEL]->(l);

// ============================================================================
// 11. Issues — INTEG (Nova ACME Integration)
// ============================================================================

// INTEG-1 task — in review
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainivg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isaink3g'})
  ON CREATE SET i += {
    numeric_id:  1,
    kind:        'task',
    title:       'Wire webhook callbacks',
    description: 'Deliver ACME event webhooks into the Nova Labs delivery pipeline.',
    status:      'review',
    priority:    'high',
    resolution:  'none',
    links:       ['https://novalabs.dev/docs/webhooks\tWebhook docs'],
    due_date:    datetime() + duration('P2D'),
    start_date:  datetime() - duration('P5D'),
    created_at:  datetime() - duration('P7D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isaink40', created_at: datetime() - duration('P7D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isaink4g', created_at: datetime() - duration('P7D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaink50', created_at: datetime() - duration('P7D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink3g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjlg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink3g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmig'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink3g'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
CREATE
  (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaink5g', kind: 'assignee', created_at: datetime() - duration('P6D')}]->(i),
  (hector)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaink60', kind: 'reviewer', created_at: datetime() - duration('P3D')}]->(i),
  (hector)-[:WATCHES {id: 'd9tcjmf92rs8isaink6g', created_at: datetime() - duration('P3D')}]->(i);

// INTEG-2 story — related to PLAT-2 and PLAT-10
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainivg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainoeg'})
  ON CREATE SET i += {
    numeric_id:  2,
    kind:        'story',
    title:       'Map ACME issue status onto the delivery board',
    description: 'Show ACME status and priority on Nova delivery cards without a second source of truth.',
    status:      'in progress',
    priority:    'high',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P9D'),
    start_date:  datetime() - duration('P1D'),
    created_at:  datetime() - duration('P4D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainof0', created_at: datetime() - duration('P4D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainofg', created_at: datetime() - duration('P4D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainogg', created_at: datetime() - duration('P4D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainoeg'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjlg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (integ:Issue {id: 'd9tcjmf92rs8isainoeg'})
MATCH (platBoard:Issue {id: 'd9tcjmf92rs8isainju0'})
MATCH (platMap:Issue {id: 'd9tcjmf92rs8isainnd0'})
CREATE
  (integ)-[:RELATED_TO {id: 'd9tcjmf92rs8isainoh0', kind: 'related to', created_at: datetime() - duration('P4D')}]->(platBoard),
  (platMap)-[:RELATED_TO {id: 'd9tcjmf92rs8isainohg', kind: 'related to', created_at: datetime() - duration('P4D')}]->(integ);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainoeg'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
CREATE
  (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainoi0', kind: 'assignee', created_at: datetime() - duration('P1D')}]->(i);

// INTEG-3 bug — blocked by INTEG-1
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (blocker:Issue {id: 'd9tcjmf92rs8isaink3g'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainoj0'})
  ON CREATE SET i += {
    numeric_id:  3,
    kind:        'bug',
    title:       'Callback retries drop Authorization',
    description: 'Retried ACME callbacks are sent without the signed Authorization header.',
    status:      'open',
    priority:    'highest',
    resolution:  'none',
    links:       [],
    due_date:    datetime() + duration('P3D'),
    start_date:  null,
    created_at:  datetime() - duration('P2D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainojg', created_at: datetime() - duration('P2D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainok0', created_at: datetime() - duration('P2D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainol0', created_at: datetime() - duration('P2D')}]->(p),
  (i)-[:RELATED_TO {id: 'd9tcjmf92rs8isainolg', kind: 'blocked by', created_at: datetime() - duration('P2D')}]->(blocker);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainoj0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjkg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainoj0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmig'})
CREATE (i)-[:HAS_LABEL]->(l);

// INTEG-4 task — lowest, unassigned
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
MATCH (reporter:User {id: 'd9tcjmf92rs8isainivg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isainom0'})
  ON CREATE SET i += {
    numeric_id:  4,
    kind:        'task',
    title:       'Partner runbook for ACME cutover',
    description: 'Write the Nova-side cutover checklist once webhooks are stable.',
    status:      'open',
    priority:    'lowest',
    resolution:  'none',
    links:       [],
    due_date:    null,
    start_date:  null,
    created_at:  datetime() - duration('P1D')
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainomg', created_at: datetime() - duration('P1D')}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainon0', created_at: datetime() - duration('P1D')}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainoo0', created_at: datetime() - duration('P1D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainom0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmhg'})
CREATE (i)-[:HAS_LABEL]->(l);

// ============================================================================
// 12. Comments and attachments
// ============================================================================

// Comments on PLAT-3
MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MERGE (c:Comment {id: 'd9tcjmf92rs8isaink70'})
  ON CREATE SET c += {
    content:    'Reproduced as a guest with write on PLAT — redirect lands on ACME org home.',
    created_by: 'd9tcjmf92rs8isainj00',
    created_at: datetime() - duration('P6D')
  }
CREATE
  (i)-[:HAS_COMMENT {id: 'd9tcjmf92rs8isaink7g', created_at: datetime() - duration('P6D')}]->(c),
  (jordan)-[:COMMENTED {id: 'd9tcjmf92rs8isaink80', created_at: datetime() - duration('P6D')}]->(c);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (demo:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (c:Comment {id: 'd9tcjmf92rs8isaink90'})
  ON CREATE SET c += {
    content:    'Thanks — we will keep this on the partner board until the redirect is fixed.',
    created_by: '9bsv0s46s6s002p9ltq0',
    created_at: datetime() - duration('P5D')
  }
CREATE
  (i)-[:HAS_COMMENT {id: 'd9tcjmf92rs8isaink9g', created_at: datetime() - duration('P5D')}]->(c),
  (demo)-[:COMMENTED {id: 'd9tcjmf92rs8isainka0', created_at: datetime() - duration('P5D')}]->(c);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
MERGE (c:Comment {id: 'd9tcjmf92rs8isainoog'})
  ON CREATE SET c += {
    content:    'Secret rotation is tracked on INF-4. Do not ship partner access until that lands.',
    created_by: 'd9tcjmf92rs8isainm00',
    created_at: datetime() - duration('P1D')
  }
CREATE
  (i)-[:HAS_COMMENT {id: 'd9tcjmf92rs8isainop0', created_at: datetime() - duration('P1D')}]->(c),
  (priya)-[:COMMENTED {id: 'd9tcjmf92rs8isainopg', created_at: datetime() - duration('P1D')}]->(c);

// Comment on INTEG-1
MATCH (i:Issue {id: 'd9tcjmf92rs8isaink3g'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
MERGE (c:Comment {id: 'd9tcjmf92rs8isainoqg'})
  ON CREATE SET c += {
    content:    'Reviewing the callback signatures now. Please keep the retry bug on INTEG-3 in the same PR.',
    created_by: '9bsv0s314mtg02goaimg',
    created_at: datetime() - duration('P1D')
  }
CREATE
  (i)-[:HAS_COMMENT {id: 'd9tcjmf92rs8isainor0', created_at: datetime() - duration('P1D')}]->(c),
  (hector)-[:COMMENTED {id: 'd9tcjmf92rs8isainorg', created_at: datetime() - duration('P1D')}]->(c);

// Comment on Product Handbook
MATCH (d:Document {id: 'd9tcjmf92rs8isainjm0'})
MATCH (aisha:User {id: 'd9tcjmf92rs8isainm10'})
MERGE (c:Comment {id: 'd9tcjmf92rs8isainosg'})
  ON CREATE SET c += {
    content:    'Added a short partner onboarding section. Please review the screenshot of the shared board.',
    created_by: 'd9tcjmf92rs8isainm10',
    created_at: datetime() - duration('P2D')
  }
CREATE
  (d)-[:HAS_COMMENT {id: 'd9tcjmf92rs8isainot0', created_at: datetime() - duration('P2D')}]->(c),
  (aisha)-[:COMMENTED {id: 'd9tcjmf92rs8isainotg', created_at: datetime() - duration('P2D')}]->(c);

// Attachment on PLAT-3
MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MERGE (a:Attachment {id: 'd9tcjmf92rs8isainoug'})
  ON CREATE SET a += {
    name:       'guest-redirect.png',
    file_id:    'demo/guest-redirect.png',
    created_by: 'd9tcjmf92rs8isainj00',
    created_at: datetime() - duration('P6D')
  }
CREATE
  (i)-[:HAS_ATTACHMENT {id: 'd9tcjmf92rs8isainov0', created_at: datetime() - duration('P6D')}]->(a),
  (jordan)-[:CREATED {id: 'd9tcjmf92rs8isainovg', created_at: datetime() - duration('P6D')}]->(a);

// Attachment on Product Handbook
MATCH (d:Document {id: 'd9tcjmf92rs8isainjm0'})
MATCH (aisha:User {id: 'd9tcjmf92rs8isainm10'})
MERGE (a:Attachment {id: 'd9tcjmf92rs8isainp00'})
  ON CREATE SET a += {
    name:       'partner-board.png',
    file_id:    'demo/partner-board.png',
    created_by: 'd9tcjmf92rs8isainm10',
    created_at: datetime() - duration('P2D')
  }
CREATE
  (d)-[:HAS_ATTACHMENT {id: 'd9tcjmf92rs8isainp0g', created_at: datetime() - duration('P2D')}]->(a),
  (aisha)-[:CREATED {id: 'd9tcjmf92rs8isainp10', created_at: datetime() - duration('P2D')}]->(a);

// Attachment on INF-3
MATCH (i:Issue {id: 'd9tcjmf92rs8isaino20'})
MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
MERGE (a:Attachment {id: 'd9tcjmf92rs8isainp1g'})
  ON CREATE SET a += {
    name:       'redis-eviction.json',
    file_id:    'demo/redis-eviction.json',
    created_by: 'd9tcjmf92rs8isainm00',
    created_at: datetime() - duration('P3D')
  }
CREATE
  (i)-[:HAS_ATTACHMENT {id: 'd9tcjmf92rs8isainp20', created_at: datetime() - duration('P3D')}]->(a),
  (priya)-[:CREATED {id: 'd9tcjmf92rs8isainp2g', created_at: datetime() - duration('P3D')}]->(a);

// ============================================================================
// 13. Private todos
// ============================================================================

MATCH (demo:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainkb0'})
  ON CREATE SET t += {
    title:       'Review partner access on PLAT',
    description: 'Confirm Jordan can see shared issues without ACME membership.',
    priority:    'important',
    completed:   false,
    due_date:    datetime() + duration('P1D'),
    created_at:  datetime() - duration('P2D')
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainkbg', created_at: datetime() - duration('P2D')}]->(demo),
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainkc0', created_at: datetime() - duration('P2D')}]->(t);

MATCH (demo:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainp30'})
  ON CREATE SET t += {
    title:       'Prep MOB sprint goals',
    description: 'Offline shell and iOS keyboard bug should be in the next cut.',
    priority:    'normal',
    completed:   false,
    due_date:    datetime() + duration('P3D'),
    created_at:  datetime() - duration('P1D')
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainp3g', created_at: datetime() - duration('P1D')}]->(demo),
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainp40', created_at: datetime() - duration('P1D')}]->(t);

MATCH (demo:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainp50'})
  ON CREATE SET t += {
    title:       'Follow up on staging Redis',
    description: 'INF-3 is still blocked; check whether eviction policy changed.',
    priority:    'urgent',
    completed:   false,
    due_date:    datetime() - duration('P1D'),
    created_at:  datetime() - duration('P3D')
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainp5g', created_at: datetime() - duration('P3D')}]->(demo),
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainp60', created_at: datetime() - duration('P3D')}]->(t);

MATCH (demo:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainp70'})
  ON CREATE SET t += {
    title:       'Archive completed onboarding notes',
    description: 'Move the old contractor checklist out of the Product handbook.',
    priority:    'critical',
    completed:   true,
    due_date:    datetime() - duration('P5D'),
    created_at:  datetime() - duration('P8D')
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainp7g', created_at: datetime() - duration('P8D')}]->(demo),
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainp80', created_at: datetime() - duration('P8D')}]->(t);

MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainkd0'})
  ON CREATE SET t += {
    title:       'Finish webhook callback wiring',
    description: 'Close out INTEG-1 after Hector reviews the ACME callbacks.',
    priority:    'urgent',
    completed:   false,
    due_date:    datetime() + duration('P2D'),
    created_at:  datetime() - duration('P4D')
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainkdg', created_at: datetime() - duration('P4D')}]->(jordan),
  (jordan)-[:CREATED {id: 'd9tcjmf92rs8isainke0', created_at: datetime() - duration('P4D')}]->(t);

MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainp90'})
  ON CREATE SET t += {
    title:       'Draft Redis eviction RCA',
    description: 'Keep it short; attach the staging dump from INF-3.',
    priority:    'important',
    completed:   false,
    due_date:    datetime() + duration('P1D'),
    created_at:  datetime() - duration('P2D')
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainp9g', created_at: datetime() - duration('P2D')}]->(priya),
  (priya)-[:CREATED {id: 'd9tcjmf92rs8isainpa0', created_at: datetime() - duration('P2D')}]->(t);

MATCH (maya:User {id: 'd9tcjmf92rs8isainivg'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainpb0'})
  ON CREATE SET t += {
    title:       'Schedule ACME cutover review',
    description: 'Wait for INTEG-1 review before inviting Demo and Hector.',
    priority:    'normal',
    completed:   false,
    due_date:    datetime() + duration('P5D'),
    created_at:  datetime() - duration('P1D')
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainpbg', created_at: datetime() - duration('P1D')}]->(maya),
  (maya)-[:CREATED {id: 'd9tcjmf92rs8isainpc0', created_at: datetime() - duration('P1D')}]->(t);

// ============================================================================
// 14. Authorization: principals, role templates, GRANTED, IN_SCOPE_OF
// ============================================================================

MATCH (n)
WHERE n:User OR n:Team OR n:Organization
SET n:Principal;

// Direct Installation grant so demo can create orgs. Not a role. Not org membership.
MATCH (u:User {email: 'demo@elemo.example'})
MATCH (i:Installation {id: '00000000000000000000'})
SET u:Principal
MERGE (u)-[g:GRANTED {id: 'd9tcjmf92rs8isait00'}]->(i)
ON CREATE SET g.actions = ['organization.create', 'plugin.install'], g.created_at = datetime();

// Copy RoleTemplates onto each org. None include organization.create.
UNWIND [
  {
    org_id: '9bsv0s4vl6gg02sv7jrg',
    roles: [
      {id: 'd9tcjmf92rs8isainr00', key: 'org-admin', name: 'Organization admin', description: 'Full authority within an organization scope, excluding organization.create.', actions: [
        'organization.read', 'organization.update', 'organization.delete', 'organization.members.manage',
        'namespace.create', 'namespace.read', 'namespace.update', 'namespace.delete',
        'project.create', 'project.read', 'project.update', 'project.delete', 'project.members.manage',
        'issue.create', 'issue.read', 'issue.update', 'issue.delete', 'issue.assign',
        'document.create', 'document.read', 'document.update', 'document.delete', 'folder.create',
        'role.manage', 'team.manage', 'permission.manage', 'custom_field.manage', 'plugin.manage',
        'extension.create', 'extension.read', 'extension.update', 'extension.delete'
      ]},
      {id: 'd9tcjmf92rs8isainr10', key: 'org-member', name: 'Organization member', description: 'Read the organization they belong to, including plugin graph nodes.', actions: ['organization.read', 'extension.read']},
      {id: 'd9tcjmf92rs8isainr20', key: 'namespace-admin', name: 'Namespace admin', description: 'Administer a namespace and its descendants.', actions: [
        'namespace.read', 'namespace.update', 'namespace.delete',
        'project.create', 'project.read', 'project.update', 'project.delete', 'project.members.manage',
        'issue.create', 'issue.read', 'issue.update', 'issue.delete', 'issue.assign',
        'document.create', 'document.read', 'document.update', 'document.delete', 'folder.create',
        'team.manage', 'permission.manage', 'custom_field.manage', 'plugin.manage',
        'extension.create', 'extension.read', 'extension.update', 'extension.delete'
      ]},
      {id: 'd9tcjmf92rs8isainr30', key: 'project-maintainer', name: 'Project maintainer', description: 'Maintain a project and its issues and documents.', actions: [
        'project.read', 'project.update', 'project.delete', 'project.members.manage',
        'issue.create', 'issue.read', 'issue.update', 'issue.delete', 'issue.assign',
        'document.create', 'document.read', 'document.update', 'document.delete', 'folder.create',
        'team.manage', 'permission.manage', 'custom_field.manage', 'plugin.manage',
        'extension.create', 'extension.read', 'extension.update', 'extension.delete'
      ]},
      {id: 'd9tcjmf92rs8isainr40', key: 'project-viewer', name: 'Project viewer', description: 'Read a project and its issues and documents.', actions: [
        'project.read', 'issue.read', 'document.read', 'extension.read'
      ]},
      {id: 'd9tcjmf92rs8isainr50', key: 'issue-maintainer', name: 'Issue maintainer', description: 'Update and assign an issue.', actions: [
        'issue.read', 'issue.update', 'issue.delete', 'issue.assign'
      ]},
      {id: 'd9tcjmf92rs8isainr60', key: 'document-maintainer', name: 'Document maintainer', description: 'Update a document or folder.', actions: [
        'document.read', 'document.update', 'document.delete', 'folder.create'
      ]}
    ]
  },
  {
    org_id: 'd9tcjmf92rs8isainj10',
    roles: [
      {id: 'd9tcjmf92rs8isainr80', key: 'org-admin', name: 'Organization admin', description: 'Full authority within an organization scope, excluding organization.create.', actions: [
        'organization.read', 'organization.update', 'organization.delete', 'organization.members.manage',
        'namespace.create', 'namespace.read', 'namespace.update', 'namespace.delete',
        'project.create', 'project.read', 'project.update', 'project.delete', 'project.members.manage',
        'issue.create', 'issue.read', 'issue.update', 'issue.delete', 'issue.assign',
        'document.create', 'document.read', 'document.update', 'document.delete', 'folder.create',
        'role.manage', 'team.manage', 'permission.manage', 'custom_field.manage', 'plugin.manage',
        'extension.create', 'extension.read', 'extension.update', 'extension.delete'
      ]},
      {id: 'd9tcjmf92rs8isainr90', key: 'org-member', name: 'Organization member', description: 'Read the organization they belong to, including plugin graph nodes.', actions: ['organization.read', 'extension.read']},
      {id: 'd9tcjmf92rs8isainra0', key: 'namespace-admin', name: 'Namespace admin', description: 'Administer a namespace and its descendants.', actions: [
        'namespace.read', 'namespace.update', 'namespace.delete',
        'project.create', 'project.read', 'project.update', 'project.delete', 'project.members.manage',
        'issue.create', 'issue.read', 'issue.update', 'issue.delete', 'issue.assign',
        'document.create', 'document.read', 'document.update', 'document.delete', 'folder.create',
        'team.manage', 'permission.manage', 'custom_field.manage', 'plugin.manage',
        'extension.create', 'extension.read', 'extension.update', 'extension.delete'
      ]},
      {id: 'd9tcjmf92rs8isainrb0', key: 'project-maintainer', name: 'Project maintainer', description: 'Maintain a project and its issues and documents.', actions: [
        'project.read', 'project.update', 'project.delete', 'project.members.manage',
        'issue.create', 'issue.read', 'issue.update', 'issue.delete', 'issue.assign',
        'document.create', 'document.read', 'document.update', 'document.delete', 'folder.create',
        'team.manage', 'permission.manage', 'custom_field.manage', 'plugin.manage',
        'extension.create', 'extension.read', 'extension.update', 'extension.delete'
      ]},
      {id: 'd9tcjmf92rs8isainrc0', key: 'project-viewer', name: 'Project viewer', description: 'Read a project and its issues and documents.', actions: [
        'project.read', 'issue.read', 'document.read', 'extension.read'
      ]},
      {id: 'd9tcjmf92rs8isainrd0', key: 'issue-maintainer', name: 'Issue maintainer', description: 'Update and assign an issue.', actions: [
        'issue.read', 'issue.update', 'issue.delete', 'issue.assign'
      ]},
      {id: 'd9tcjmf92rs8isainre0', key: 'document-maintainer', name: 'Document maintainer', description: 'Update a document or folder.', actions: [
        'document.read', 'document.update', 'document.delete', 'folder.create'
      ]}
    ]
  }
] AS spec
MATCH (o:Organization {id: spec.org_id})
WITH o, spec
UNWIND spec.roles AS tmpl
MERGE (r:Role {id: tmpl.id})
ON CREATE SET
  r.key = tmpl.key,
  r.name = tmpl.name,
  r.description = tmpl.description,
  r.actions = tmpl.actions,
  r.created_at = datetime()
ON MATCH SET r.actions = tmpl.actions
MERGE (o)-[d:DEFINES_ROLE {id: tmpl.id}]->(r)
ON CREATE SET d.created_at = datetime()
MERGE (r)-[s:IN_SCOPE_OF]->(o)
ON CREATE SET s.id = tmpl.id, s.created_at = datetime();

// Additive: orgs created before extension.* existed keep template roles current.
UNWIND [
  {key: 'org-admin', extra: ['extension.create', 'extension.read', 'extension.update', 'extension.delete']},
  {key: 'namespace-admin', extra: ['extension.create', 'extension.read', 'extension.update', 'extension.delete']},
  {key: 'project-maintainer', extra: ['extension.create', 'extension.read', 'extension.update', 'extension.delete']},
  {key: 'org-member', extra: ['extension.read']},
  {key: 'project-viewer', extra: ['extension.read']}
] AS patch
MATCH (r:Role {key: patch.key})
SET r.actions = r.actions + [a IN patch.extra WHERE NOT a IN r.actions];

// Intra-org: demo is ACME org-admin; ACME members inherit org-member via the org principal.
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MATCH (o)-[:DEFINES_ROLE]->(admin:Role {key: 'org-admin'})
MERGE (u)-[g:GRANTED {id: '9bsv0s613svg02gik0r0'}]->(o)
ON CREATE SET g.role_id = admin.id, g.actions = [], g.created_at = datetime();

MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MATCH (o)-[:DEFINES_ROLE]->(member:Role {key: 'org-member'})
MERGE (o)-[g:GRANTED {id: '9bsv0s314mtg02goaing'}]->(o)
ON CREATE SET g.role_id = member.id, g.actions = [], g.created_at = datetime();

// Intra-org: Maya is Nova org-admin; Nova members inherit org-member via the org principal.
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MATCH (o:Organization {id: 'd9tcjmf92rs8isainj10'})
MATCH (o)-[:DEFINES_ROLE]->(admin:Role {key: 'org-admin'})
MERGE (u)-[g:GRANTED {id: 'd9tcjmf92rs8isainj20'}]->(o)
ON CREATE SET g.role_id = admin.id, g.actions = [], g.created_at = datetime();

MATCH (o:Organization {id: 'd9tcjmf92rs8isainj10'})
MATCH (o)-[:DEFINES_ROLE]->(member:Role {key: 'org-member'})
MERGE (o)-[g:GRANTED {id: 'd9tcjmf92rs8isainj30'}]->(o)
ON CREATE SET g.role_id = member.id, g.actions = [], g.created_at = datetime();

// Team grants (HAS_TEAM remains structural ownership).
MATCH (t:Team {id: 'd9tcjmf92rs8isainjdg'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MATCH (:Organization {id: '9bsv0s4vl6gg02sv7jrg'})-[:DEFINES_ROLE]->(r:Role {key: 'namespace-admin'})
MERGE (t)-[g:GRANTED {id: 'd9tcjmf92rs8isainjg0'}]->(ns)
ON CREATE SET g.role_id = r.id, g.actions = [], g.created_at = datetime();

MATCH (t:Team {id: 'd9tcjmf92rs8isainm70'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj6g'})
MATCH (:Organization {id: '9bsv0s4vl6gg02sv7jrg'})-[:DEFINES_ROLE]->(r:Role {key: 'namespace-admin'})
MERGE (t)-[g:GRANTED {id: 'd9tcjmf92rs8isainm9g'}]->(ns)
ON CREATE SET g.role_id = r.id, g.actions = [], g.created_at = datetime();

MATCH (t:Team {id: 'd9tcjmf92rs8isainmag'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MATCH (:Organization {id: '9bsv0s4vl6gg02sv7jrg'})-[:DEFINES_ROLE]->(r:Role {key: 'project-viewer'})
MERGE (t)-[g:GRANTED {id: 'd9tcjmf92rs8isainmd0'}]->(ns)
ON CREATE SET g.role_id = r.id, g.actions = [], g.created_at = datetime();

MATCH (t:Team {id: 'd9tcjmf92rs8isainjgg'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (:Organization {id: '9bsv0s4vl6gg02sv7jrg'})-[:DEFINES_ROLE]->(r:Role {key: 'project-maintainer'})
MERGE (t)-[g:GRANTED {id: 'd9tcjmf92rs8isainjj0'}]->(p)
ON CREATE SET g.role_id = r.id, g.actions = [], g.created_at = datetime();

// Cross-org: Nova Labs (not ACME members) can view ACME PLAT.
MATCH (nova:Organization {id: 'd9tcjmf92rs8isainj10'})
MATCH (plat:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (nova)-[:DEFINES_ROLE]->(r:Role {key: 'project-viewer'})
MERGE (nova)-[g:GRANTED {id: 'd9tcjmf92rs8isainjjg'}]->(plat)
ON CREATE SET g.role_id = r.id, g.actions = [], g.created_at = datetime();

// Nova engineers need more than org-member on INTEG.
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
MATCH (:Organization {id: 'd9tcjmf92rs8isainj10'})-[:DEFINES_ROLE]->(r:Role {key: 'project-maintainer'})
MERGE (jordan)-[g:GRANTED {id: 'd9tcjmf92rs8isainjd0'}]->(p)
ON CREATE SET g.role_id = r.id, g.actions = [], g.created_at = datetime();

MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
MATCH (:Organization {id: 'd9tcjmf92rs8isainj10'})-[:DEFINES_ROLE]->(r:Role {key: 'project-viewer'})
MERGE (hector)-[g:GRANTED {id: 'd9tcjmf92rs8isainjk0'}]->(p)
ON CREATE SET g.role_id = r.id, g.actions = [], g.created_at = datetime();

// IN_SCOPE_OF in addition to domain edges (HAS_NAMESPACE, BELONGS_TO, SCOPED_TO, ...).
MATCH (ns:Namespace)<-[:HAS_NAMESPACE]-(o:Organization)
MERGE (ns)-[rel:IN_SCOPE_OF]->(o)
ON CREATE SET rel.id = ns.id, rel.created_at = datetime();

MATCH (p:Project)<-[:HAS_PROJECT]-(ns:Namespace)
MERGE (p)-[rel:IN_SCOPE_OF]->(ns)
ON CREATE SET rel.id = p.id, rel.created_at = datetime();

MATCH (i:Issue)-[:BELONGS_TO]->(p:Project)
MERGE (i)-[rel:IN_SCOPE_OF]->(p)
ON CREATE SET rel.id = i.id, rel.created_at = datetime();

MATCH (d:Document)-[:SCOPED_TO]->(lib)
MERGE (d)-[rel:IN_SCOPE_OF]->(lib)
ON CREATE SET rel.id = d.id, rel.created_at = datetime();

MATCH (f:Folder)-[:SCOPED_TO]->(lib)
MERGE (f)-[rel:IN_SCOPE_OF]->(lib)
ON CREATE SET rel.id = f.id, rel.created_at = datetime();

MATCH (t:Todo)-[:BELONGS_TO]->(u:User)
MERGE (t)-[rel:IN_SCOPE_OF]->(u)
ON CREATE SET rel.id = t.id, rel.created_at = datetime();

MATCH (parent)-[:HAS_COMMENT]->(c:Comment)
MERGE (c)-[rel:IN_SCOPE_OF]->(parent)
ON CREATE SET rel.id = c.id, rel.created_at = datetime();
