// ============================================================================
// Overview
//
// Comprehensive demo seed for local development and UI exploration.
// Creates two organizations (ACME Inc. and Nova Labs) with namespaces,
// projects, teams, documents, issues, and cross-org collaboration.
//
// Login: demo@elemo.app / AppleTree123
//
// Requires bootstrap.cypher (ResourceTypes + system Roles) to have run first.
// ============================================================================

// ============================================================================
// 1. Users
// ============================================================================

// Demo — ACME owner + system Owner
MERGE (u:User {id: '9bsv0s46s6s002p9ltq0'})
  ON CREATE SET u += {
    username:   'demo',
    email:      'demo@elemo.app',
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
  }
WITH u
MATCH (r:Role {id: 'Owner'})
CREATE (u)-[:MEMBER_OF {id: '9bsv0s3n4ccg0329pecg', created_at: datetime()}]->(r);

// Hector — ACME member (read); also joins Nova Labs later
MERGE (u:User {id: '9bsv0s314mtg02goaimg'})
  ON CREATE SET u += {
    username:   'hector-henrik',
    email:      'hector@elemo.app',
    password:   '$2a$10$LLoJgBl7Y24MPz8smg4ruO9GARZ9SW2uZ2qI0D9AwhYpZYs/AHC/C', // AppleTree123
    status:     'active',
    first_name: 'Hector',
    last_name:  'Henrik',
    picture:    'https://picsum.photos/id/22/200/200',
    title:      'Senior Software Developer',
    bio:        "Hello. It's me!",
    phone:      '+12345678901',
    address:    '2900 S Congress Ave, Austin, TX',
    links:      ['https://example.com'],
    languages:  ['en', 'es'],
    created_at: datetime()
  };

// Maya — Nova Labs owner
MERGE (u:User {id: 'd9tcjmf92rs8isainivg'})
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

// Jordan — Nova Labs engineer; guest collaborator on ACME PLAT (no ACME membership)
MERGE (u:User {id: 'd9tcjmf92rs8isainj00'})
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
MERGE (u:User {id: 'd9tcjmf92rs8isainj0g'})
  ON CREATE SET u += {
    username:   'sam-rivera',
    email:      'sam@elemo.app',
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

// ============================================================================
// 2. Organizations, memberships, and invitations
// ============================================================================

// ACME Inc. — owned by demo
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
  ON CREATE SET o += {
    name:       'ACME Inc.',
    email:      'info@example.com',
    logo:       'https://picsum.photos/id/211/200/200.jpg',
    website:    'https://example.com',
    status:     'active',
    created_at: datetime()
  }
CREATE
  (u)-[:MEMBER_OF {id: '9m4e2mr0ui3e8a215n4g', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: '9bsv0s613svg02gik0r0', created_at: datetime(), kind: '*'}]->(o);

// Hector → ACME (read)
MATCH (u:User {id: '9bsv0s314mtg02goaimg'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: '9bsv0s314mtg02goain0', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: '9bsv0s314mtg02goaing', created_at: datetime(), kind: 'read'}]->(o);

// Sam → ACME pending invite
MATCH (u:User {id: 'd9tcjmf92rs8isainj0g'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE (u)-[:INVITED_TO {id: 'd9tcjmf92rs8isainj4g', created_at: datetime()}]->(o);

// Nova Labs — owned by Maya
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MERGE (o:Organization {id: 'd9tcjmf92rs8isainj10'})
  ON CREATE SET o += {
    name:       'Nova Labs',
    email:      'hello@novalabs.dev',
    logo:       'https://picsum.photos/id/180/200/200.jpg',
    website:    'https://novalabs.dev',
    status:     'active',
    created_at: datetime()
  }
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainj1g', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainj20', created_at: datetime(), kind: '*'}]->(o);

// Jordan → Nova Labs (write)
MATCH (u:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (o:Organization {id: 'd9tcjmf92rs8isainj10'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainj2g', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainj30', created_at: datetime(), kind: 'write'}]->(o);

// Hector → Nova Labs (read) — multi-org membership
MATCH (u:User {id: '9bsv0s314mtg02goaimg'})
MATCH (o:Organization {id: 'd9tcjmf92rs8isainj10'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainj3g', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainj40', created_at: datetime(), kind: 'read'}]->(o);

// ============================================================================
// 3. Namespaces and projects
// ============================================================================

// ACME → Product namespace
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MERGE (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
  ON CREATE SET ns += {
    name:        'Product',
    description: 'Product engineering workspaces for ACME platforms and apps.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_NAMESPACE {id: 'd9tcjmf92rs8isainj5g', created_at: datetime()}]->(ns),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainj60', created_at: datetime(), kind: '*'}]->(ns);

// ACME → Operations namespace
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MERGE (ns:Namespace {id: 'd9tcjmf92rs8isainj6g'})
  ON CREATE SET ns += {
    name:        'Operations',
    description: 'Infrastructure and internal operations for ACME.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_NAMESPACE {id: 'd9tcjmf92rs8isainj70', created_at: datetime()}]->(ns),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainj7g', created_at: datetime(), kind: '*'}]->(ns);

// Nova Labs → Delivery namespace
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MATCH (o:Organization {id: 'd9tcjmf92rs8isainj10'})
MERGE (ns:Namespace {id: 'd9tcjmf92rs8isainj80'})
  ON CREATE SET ns += {
    name:        'Delivery',
    description: 'Client delivery and partner integration projects.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_NAMESPACE {id: 'd9tcjmf92rs8isainj8g', created_at: datetime()}]->(ns),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainj90', created_at: datetime(), kind: '*'}]->(ns);

// ACME Product → PLAT project
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainj9g'})
  ON CREATE SET p += {
    key:         'PLAT',
    name:        'Elemo Platform',
    description: 'Core product platform shared with partner teams.',
    logo:        'https://picsum.photos/id/201/200/200.jpg',
    status:      'active',
    created_at:  datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainja0', created_at: datetime(), kind: '*'}]->(p);

// ACME Product → MOBILE project
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjag'})
  ON CREATE SET p += {
    key:         'MOBILE',
    name:        'Mobile App',
    description: 'Native mobile clients for ACME customers.',
    logo:        '',
    status:      'active',
    created_at:  datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjb0', created_at: datetime(), kind: '*'}]->(p);

// ACME Operations → INFRA project
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj6g'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjbg'})
  ON CREATE SET p += {
    key:         'INFRA',
    name:        'Infrastructure',
    description: 'Cloud infrastructure and reliability workstreams.',
    logo:        '',
    status:      'active',
    created_at:  datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjc0', created_at: datetime(), kind: '*'}]->(p);

// Nova Delivery → INTEG project
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj80'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjcg'})
  ON CREATE SET p += {
    key:         'INTEG',
    name:        'ACME Integration',
    description: 'Partner delivery project for ACME platform integrations.',
    logo:        'https://picsum.photos/id/250/200/200.jpg',
    status:      'active',
    created_at:  datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjd0', created_at: datetime(), kind: '*'}]->(p);

// ============================================================================
// 4. Teams / roles
// ============================================================================

// ACME org team: Engineering (demo + hector); write on Product namespace
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MERGE (r:Role {id: 'd9tcjmf92rs8isainjdg'})
  ON CREATE SET r += {
    name:        'Engineering',
    description: 'ACME product engineering team.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_TEAM {id: 'd9tcjmf92rs8isainje0', created_at: datetime()}]->(r),
  (owner)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainjeg', created_at: datetime()}]->(r),
  (owner)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjf0', created_at: datetime(), kind: '*'}]->(r),
  (hector)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainjfg', created_at: datetime()}]->(r),
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjg0', created_at: datetime(), kind: 'write'}]->(ns);

// ACME PLAT project team: Contractors (demo + jordan); write on PLAT
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MERGE (r:Role {id: 'd9tcjmf92rs8isainjgg'})
  ON CREATE SET r += {
    name:        'Contractors',
    description: 'External partners collaborating on the platform project.',
    created_at:  datetime()
  }
CREATE
  (p)-[:HAS_TEAM {id: 'd9tcjmf92rs8isainjh0', created_at: datetime()}]->(r),
  (owner)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainjhg', created_at: datetime()}]->(r),
  (owner)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainji0', created_at: datetime(), kind: '*'}]->(r),
  (jordan)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainjig', created_at: datetime()}]->(r),
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjj0', created_at: datetime(), kind: 'write'}]->(p);

// ============================================================================
// 5. Cross-org collaboration grants
// ============================================================================

// Jordan guest write on ACME PLAT (direct grant; not an ACME org member)
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
CREATE (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjjg', created_at: datetime(), kind: 'write'}]->(p);

// Hector read on Nova INTEG (beyond org-level read, explicit project grant)
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
CREATE (hector)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjk0', created_at: datetime(), kind: 'read'}]->(p);

// ============================================================================
// 6. Labels, documents, issues, comments, and todos
// ============================================================================

// Labels
MERGE (l:Label {id: 'd9tcjmf92rs8isainjkg'})
  ON CREATE SET l += {
    name:        'bug',
    description: 'Defects and regressions.',
    created_at:  datetime()
  };

MERGE (l:Label {id: 'd9tcjmf92rs8isainjl0'})
  ON CREATE SET l += {
    name:        'feature',
    description: 'New product capabilities.',
    created_at:  datetime()
  };

MERGE (l:Label {id: 'd9tcjmf92rs8isainjlg'})
  ON CREATE SET l += {
    name:        'partner',
    description: 'Work involving partner organizations.',
    created_at:  datetime()
  };

// Document: Product handbook (belongs to Product namespace)
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (d:Document {id: 'd9tcjmf92rs8isainjm0'})
  ON CREATE SET d += {
    name:       'Product Handbook',
    excerpt:    'Orientation guide for ACME product engineering and partners.',
    file_id:    '',
    created_by: '9bsv0s46s6s002p9ltq0',
    created_at: datetime()
  }
CREATE
  (d)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainjmg', created_at: datetime()}]->(ns),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjn0', created_at: datetime()}]->(d);

// Document: Platform overview (belongs to PLAT)
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (d:Document {id: 'd9tcjmf92rs8isainjng'})
  ON CREATE SET d += {
    name:       'Platform Overview',
    excerpt:    'Architecture notes and onboarding for the Elemo platform project.',
    file_id:    '',
    created_by: '9bsv0s46s6s002p9ltq0',
    created_at: datetime()
  }
CREATE
  (d)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainjo0', created_at: datetime()}]->(p),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjog', created_at: datetime()}]->(d);

// Document: Integration guide (belongs to INTEG)
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MERGE (d:Document {id: 'd9tcjmf92rs8isainjp0'})
  ON CREATE SET d += {
    name:       'Integration Guide',
    excerpt:    'How Nova Labs integrates with ACME platform APIs and workflows.',
    file_id:    '',
    created_by: 'd9tcjmf92rs8isainivg',
    created_at: datetime()
  }
CREATE
  (d)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainjpg', created_at: datetime()}]->(p),
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjq0', created_at: datetime()}]->(d);

// Issue: PLAT-1 epic — Partner collaboration rollout
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
    links:       [],
    due_date:    null,
    created_at:  datetime()
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainjr0', created_at: datetime()}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainjrg', created_at: datetime()}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainjs0', created_at: datetime()}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainjqg'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjlg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainjqg'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
CREATE
  (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainjsg', kind: 'assignee', created_at: datetime()}]->(i),
  (hector)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainjt0', kind: 'reviewer', created_at: datetime()}]->(i),
  (jordan)-[:WATCHES {id: 'd9tcjmf92rs8isainjtg', created_at: datetime()}]->(i);

// Issue: PLAT-2 story — Shared issue board for partners (subtask of epic)
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
    priority:    'medium',
    resolution:  'none',
    links:       [],
    due_date:    null,
    created_at:  datetime()
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isainjug', created_at: datetime()}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isainjv0', created_at: datetime()}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainjvg', created_at: datetime()}]->(p),
  (i)-[:RELATED_TO {id: 'd9tcjmf92rs8isaink00', kind: 'subtask of', created_at: datetime()}]->(parent);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainju0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjl0'})
CREATE (i)-[:HAS_LABEL]->(l);

// Issue: PLAT-3 bug — Auth redirect for guest collaborators
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
MATCH (reporter:User {id: '9bsv0s314mtg02goaimg'})
MERGE (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
  ON CREATE SET i += {
    numeric_id:  3,
    kind:        'bug',
    title:       'Auth redirect for guest collaborators',
    description: 'Guest users with project write permission hit an unexpected org redirect.',
    status:      'blocked',
    priority:    'critical',
    resolution:  'none',
    links:       ['https://example.com/incidents/guest-auth'],
    due_date:    null,
    created_at:  datetime()
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isaink10', created_at: datetime()}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isaink1g', created_at: datetime()}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaink20', created_at: datetime()}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjkg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (bug:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (story:Issue {id: 'd9tcjmf92rs8isainju0'})
CREATE (bug)-[:RELATED_TO {id: 'd9tcjmf92rs8isaink2g', kind: 'blocks', created_at: datetime()}]->(story);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
CREATE (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaink30', kind: 'assignee', created_at: datetime()}]->(i);

// Issue: INTEG-1 task — Wire webhook callbacks
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
    links:       [],
    due_date:    null,
    created_at:  datetime()
  }
CREATE
  (reporter)-[:CREATED {id: 'd9tcjmf92rs8isaink40', created_at: datetime()}]->(i),
  (reporter)-[:WATCHES {id: 'd9tcjmf92rs8isaink4g', created_at: datetime()}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaink50', created_at: datetime()}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink3g'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainjlg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink3g'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
CREATE
  (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaink5g', kind: 'assignee', created_at: datetime()}]->(i),
  (hector)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaink60', kind: 'reviewer', created_at: datetime()}]->(i),
  (hector)-[:WATCHES {id: 'd9tcjmf92rs8isaink6g', created_at: datetime()}]->(i);

// Comments on PLAT-3 bug
MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MERGE (c:Comment {id: 'd9tcjmf92rs8isaink70'})
  ON CREATE SET c += {
    content:    'Reproduced as a guest with write on PLAT — redirect lands on ACME org home.',
    created_by: 'd9tcjmf92rs8isainj00',
    created_at: datetime()
  }
CREATE
  (i)-[:HAS_COMMENT {id: 'd9tcjmf92rs8isaink7g', created_at: datetime()}]->(c),
  (jordan)-[:COMMENTED {id: 'd9tcjmf92rs8isaink80', created_at: datetime()}]->(c),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isaink8g', created_at: datetime(), kind: '*'}]->(c);

MATCH (i:Issue {id: 'd9tcjmf92rs8isaink0g'})
MATCH (demo:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (c:Comment {id: 'd9tcjmf92rs8isaink90'})
  ON CREATE SET c += {
    content:    'Thanks — we will keep this on the partner board until the redirect is fixed.',
    created_by: '9bsv0s46s6s002p9ltq0',
    created_at: datetime()
  }
CREATE
  (i)-[:HAS_COMMENT {id: 'd9tcjmf92rs8isaink9g', created_at: datetime()}]->(c),
  (demo)-[:COMMENTED {id: 'd9tcjmf92rs8isainka0', created_at: datetime()}]->(c),
  (demo)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainkag', created_at: datetime(), kind: '*'}]->(c);

// Private todos
MATCH (demo:User {id: '9bsv0s46s6s002p9ltq0'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainkb0'})
  ON CREATE SET t += {
    title:       'Review partner access on PLAT',
    description: 'Confirm Jordan can see shared issues without ACME membership.',
    priority:    'important',
    completed:   false,
    due_date:    null,
    created_at:  datetime()
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainkbg', created_at: datetime()}]->(demo),
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainkc0', created_at: datetime()}]->(t),
  (demo)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainkcg', created_at: datetime(), kind: '*'}]->(t);

MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MERGE (t:Todo {id: 'd9tcjmf92rs8isainkd0'})
  ON CREATE SET t += {
    title:       'Finish webhook callback wiring',
    description: 'Close out INTEG-1 after Hector reviews the ACME callbacks.',
    priority:    'urgent',
    completed:   false,
    due_date:    null,
    created_at:  datetime()
  }
CREATE
  (t)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainkdg', created_at: datetime()}]->(jordan),
  (jordan)-[:CREATED {id: 'd9tcjmf92rs8isainke0', created_at: datetime()}]->(t),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainkeg', created_at: datetime(), kind: '*'}]->(t);
