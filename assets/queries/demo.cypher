// ============================================================================
// Overview
//
// Demo seed for local development and UI exploration.
//
// Organizations
//   ACME Inc.   — Product (PLAT, MOB) and Operations (INF)
//   Nova Labs   — Delivery (INTEG), partner org collaborating on PLAT
//
// Logins (password AppleTree123 for every account)
//   demo@elemo.app       ACME owner, system Owner
//   hector@elemo.app     ACME engineering; also Nova Labs member
//   priya@elemo.app      ACME operations
//   luis@elemo.app       ACME mobile
//   aisha@elemo.app      ACME design (read)
//   maya@novalabs.dev    Nova Labs owner
//   jordan@novalabs.dev  Nova engineer; guest write on ACME PLAT
//   sam@elemo.app        pending ACME invite (not a member yet)
//
// Covers: all issue kinds/statuses/priorities, common resolutions, parent
// issues, relation kinds (except reserved "depends on"), comments,
// attachments, labels, documents, teams, invitations, and todos.
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

// Hector — ACME engineer; also joins Nova Labs later
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
    bio:        'Platform engineer on the ACME product team.',
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

// Jordan — Nova Labs engineer; guest collaborator on ACME PLAT
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

// Priya — ACME SRE
MERGE (u:User {id: 'd9tcjmf92rs8isainm00'})
  ON CREATE SET u += {
    username:   'priya-shah',
    email:      'priya@elemo.app',
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
MERGE (u:User {id: 'd9tcjmf92rs8isainm0g'})
  ON CREATE SET u += {
    username:   'luis-ortega',
    email:      'luis@elemo.app',
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
MERGE (u:User {id: 'd9tcjmf92rs8isainm10'})
  ON CREATE SET u += {
    username:   'aisha-khan',
    email:      'aisha@elemo.app',
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
MERGE (u:User {id: 'd9tcjmf92rs8isainm1g'})
  ON CREATE SET u += {
    username:   'riley-nash',
    email:      'riley@elemo.app',
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

// Priya → ACME (write)
MATCH (u:User {id: 'd9tcjmf92rs8isainm00'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm20', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainm2g', created_at: datetime(), kind: 'write'}]->(o);

// Luis → ACME (write)
MATCH (u:User {id: 'd9tcjmf92rs8isainm0g'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm30', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainm3g', created_at: datetime(), kind: 'write'}]->(o);

// Aisha → ACME (read)
MATCH (u:User {id: 'd9tcjmf92rs8isainm10'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm40', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainm4g', created_at: datetime(), kind: 'read'}]->(o);

// Riley → ACME (read, inactive account)
MATCH (u:User {id: 'd9tcjmf92rs8isainm1g'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm50', created_at: datetime() - duration('P90D')}]->(o),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainm5g', created_at: datetime() - duration('P90D'), kind: 'read'}]->(o);

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
    key:           'PLAT',
    name:          'Elemo Platform',
    description:   'Core product platform shared with partner teams.',
    logo:          'https://picsum.photos/id/201/200/200.jpg',
    status:        'active',
    next_issue_id: 10,
    created_at:    datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainja0', created_at: datetime(), kind: '*'}]->(p);

// ACME Product → MOB project
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjag'})
  ON CREATE SET p += {
    key:           'MOB',
    name:          'Mobile App',
    description:   'Native mobile clients for ACME customers.',
    logo:          '',
    status:        'active',
    next_issue_id: 4,
    created_at:    datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjb0', created_at: datetime(), kind: '*'}]->(p);

// ACME Operations → INF project
MATCH (u:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj6g'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjbg'})
  ON CREATE SET p += {
    key:           'INF',
    name:          'Infrastructure',
    description:   'Cloud infrastructure and reliability workstreams.',
    logo:          '',
    status:        'active',
    next_issue_id: 6,
    created_at:    datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjc0', created_at: datetime(), kind: '*'}]->(p);

// Nova Delivery → INTEG project
MATCH (u:User {id: 'd9tcjmf92rs8isainivg'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj80'})
MERGE (p:Project {id: 'd9tcjmf92rs8isainjcg'})
  ON CREATE SET p += {
    key:           'INTEG',
    name:          'ACME Integration',
    description:   'Partner delivery project for ACME platform integrations.',
    logo:          'https://picsum.photos/id/250/200/200.jpg',
    status:        'active',
    next_issue_id: 4,
    created_at:    datetime()
  }
CREATE
  (ns)-[:HAS_PROJECT]->(p),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjd0', created_at: datetime(), kind: '*'}]->(p);

// ============================================================================
// 4. Teams / roles
// ============================================================================

// ACME org team: Engineering (demo, hector, luis); write+read on Product
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
MATCH (luis:User {id: 'd9tcjmf92rs8isainm0g'})
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
  (luis)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm60', created_at: datetime()}]->(r),
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjg0', created_at: datetime(), kind: 'write'}]->(ns),
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainm6g', created_at: datetime(), kind: 'read'}]->(ns);

// ACME org team: Operations (demo, priya); write+read on Operations
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj6g'})
MERGE (r:Role {id: 'd9tcjmf92rs8isainm70'})
  ON CREATE SET r += {
    name:        'SRE',
    description: 'ACME site reliability and infrastructure.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_TEAM {id: 'd9tcjmf92rs8isainm7g', created_at: datetime()}]->(r),
  (owner)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm80', created_at: datetime()}]->(r),
  (owner)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainm8g', created_at: datetime(), kind: '*'}]->(r),
  (priya)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainm90', created_at: datetime()}]->(r),
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainm9g', created_at: datetime(), kind: 'write'}]->(ns),
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainma0', created_at: datetime(), kind: 'read'}]->(ns);

// ACME org team: Design (demo, aisha); read on Product
MATCH (owner:User {id: '9bsv0s46s6s002p9ltq0'})
MATCH (aisha:User {id: 'd9tcjmf92rs8isainm10'})
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
MATCH (ns:Namespace {id: 'd9tcjmf92rs8isainj50'})
MERGE (r:Role {id: 'd9tcjmf92rs8isainmag'})
  ON CREATE SET r += {
    name:        'Design',
    description: 'ACME product design.',
    created_at:  datetime()
  }
CREATE
  (o)-[:HAS_TEAM {id: 'd9tcjmf92rs8isainmb0', created_at: datetime()}]->(r),
  (owner)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainmbg', created_at: datetime()}]->(r),
  (owner)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmc0', created_at: datetime(), kind: '*'}]->(r),
  (aisha)-[:MEMBER_OF {id: 'd9tcjmf92rs8isainmcg', created_at: datetime()}]->(r),
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmd0', created_at: datetime(), kind: 'read'}]->(ns);

// ACME PLAT project team: Contractors (demo + jordan); write+read on PLAT
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
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjj0', created_at: datetime(), kind: 'write'}]->(p),
  (r)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmdg', created_at: datetime(), kind: 'read'}]->(p);

// ============================================================================
// 5. Cross-org collaboration grants
// ============================================================================

// Jordan guest write+read on ACME PLAT (direct grant; not an ACME org member)
MATCH (jordan:User {id: 'd9tcjmf92rs8isainj00'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainj9g'})
CREATE
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjjg', created_at: datetime(), kind: 'write'}]->(p),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainme0', created_at: datetime(), kind: 'read'}]->(p);

// Hector read on Nova INTEG
MATCH (hector:User {id: '9bsv0s314mtg02goaimg'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainjcg'})
CREATE (hector)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainjk0', created_at: datetime(), kind: 'read'}]->(p);

// Priya write+read on INF (in addition to Operations namespace via SRE)
MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainjbg'})
CREATE
  (priya)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmeg', created_at: datetime(), kind: 'write'}]->(p),
  (priya)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmf0', created_at: datetime(), kind: 'read'}]->(p);

// Luis write+read on MOB
MATCH (luis:User {id: 'd9tcjmf92rs8isainm0g'})
MATCH (p:Project {id: 'd9tcjmf92rs8isainjag'})
CREATE
  (luis)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmfg', created_at: datetime(), kind: 'write'}]->(p),
  (luis)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmg0', created_at: datetime(), kind: 'read'}]->(p);

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
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjn0', created_at: datetime() - duration('P20D')}]->(d),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainq70', created_at: datetime() - duration('P20D'), kind: '*'}]->(d);

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
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjog', created_at: datetime() - duration('P14D')}]->(d),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainqa0', created_at: datetime() - duration('P14D'), kind: '*'}]->(d);

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
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainjq0', created_at: datetime() - duration('P10D')}]->(d),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainqc0', created_at: datetime() - duration('P10D'), kind: '*'}]->(d);

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
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainml0', created_at: datetime() - duration('P8D')}]->(d),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainqd0', created_at: datetime() - duration('P8D'), kind: '*'}]->(d);

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
  (u)-[:CREATED {id: 'd9tcjmf92rs8isainmmg', created_at: datetime() - duration('P6D')}]->(d),
  (u)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainqg0', created_at: datetime() - duration('P6D'), kind: '*'}]->(d);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmn0', created_at: datetime() - duration('P18D'), kind: '*'}]->(i),
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
  (jordan)-[:WATCHES {id: 'd9tcjmf92rs8isainjtg', created_at: datetime() - duration('P16D')}]->(i),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmng', created_at: datetime() - duration('P16D'), kind: '*'}]->(i),
  (hector)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmo0', created_at: datetime() - duration('P16D'), kind: 'read'}]->(i);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmog', created_at: datetime() - duration('P12D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmp0', created_at: datetime() - duration('P8D'), kind: '*'}]->(i),
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
  (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaink30', kind: 'assignee', created_at: datetime() - duration('P7D')}]->(i),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmpg', created_at: datetime() - duration('P7D'), kind: '*'}]->(i);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmrg', created_at: datetime() - duration('P9D'), kind: '*'}]->(i),
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
  (aisha)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainmtg', kind: 'reviewer', created_at: datetime() - duration('P2D')}]->(i),
  (hector)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmu0', created_at: datetime() - duration('P5D'), kind: '*'}]->(i),
  (aisha)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainmug', created_at: datetime() - duration('P2D'), kind: 'read'}]->(i);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainn0g', created_at: datetime() - duration('P11D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainn3g', created_at: datetime() - duration('P5D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainn60', created_at: datetime() - duration('P12D'), kind: '*'}]->(i),
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
  (hector)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainn70', kind: 'assignee', created_at: datetime() - duration('P10D')}]->(i),
  (hector)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainn7g', created_at: datetime() - duration('P10D'), kind: '*'}]->(i);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainn9g', created_at: datetime() - duration('P10D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainnc0', created_at: datetime() - duration('P15D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainneg', created_at: datetime() - duration('P4D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainnh0', created_at: datetime() - duration('P7D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainnjg', created_at: datetime() - duration('P6D'), kind: '*'}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isainnk0', created_at: datetime() - duration('P6D')}]->(p),
  (i)-[:RELATED_TO {id: 'd9tcjmf92rs8isainnkg', kind: 'subtask of', created_at: datetime() - duration('P6D')}]->(parent);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainni0'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmj0'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainni0'})
MATCH (luis:User {id: 'd9tcjmf92rs8isainm0g'})
CREATE
  (luis)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainnl0', kind: 'assignee', created_at: datetime() - duration('P2D')}]->(i),
  (luis)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainnlg', created_at: datetime() - duration('P2D'), kind: '*'}]->(i);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainnng', created_at: datetime() - duration('P3D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainnqg', created_at: datetime() - duration('P9D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainnt0', created_at: datetime() - duration('P5D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isaino00', created_at: datetime() - duration('P9D'), kind: '*'}]->(i),
  (i)-[:BELONGS_TO {id: 'd9tcjmf92rs8isaino0g', created_at: datetime() - duration('P9D')}]->(p);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnug'})
MATCH (l:Label {id: 'd9tcjmf92rs8isainmhg'})
CREATE (i)-[:HAS_LABEL]->(l);

MATCH (i:Issue {id: 'd9tcjmf92rs8isainnug'})
MATCH (priya:User {id: 'd9tcjmf92rs8isainm00'})
CREATE
  (priya)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isaino10', kind: 'assignee', created_at: datetime() - duration('P8D')}]->(i),
  (priya)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isaino1g', created_at: datetime() - duration('P8D'), kind: '*'}]->(i);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isaino3g', created_at: datetime() - duration('P3D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isaino60', created_at: datetime() - duration('P2D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isaino90', created_at: datetime() - duration('P2D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainoc0', created_at: datetime() - duration('P16D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainod0', created_at: datetime() - duration('P7D'), kind: '*'}]->(i),
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
  (hector)-[:WATCHES {id: 'd9tcjmf92rs8isaink6g', created_at: datetime() - duration('P3D')}]->(i),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainodg', created_at: datetime() - duration('P6D'), kind: '*'}]->(i),
  (hector)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainoe0', created_at: datetime() - duration('P3D'), kind: 'read'}]->(i);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainog0', created_at: datetime() - duration('P4D'), kind: '*'}]->(i),
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
  (jordan)-[:ASSIGNED_TO {id: 'd9tcjmf92rs8isainoi0', kind: 'assignee', created_at: datetime() - duration('P1D')}]->(i),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainoig', created_at: datetime() - duration('P1D'), kind: '*'}]->(i);

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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainokg', created_at: datetime() - duration('P2D'), kind: '*'}]->(i),
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
  (reporter)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainong', created_at: datetime() - duration('P1D'), kind: '*'}]->(i),
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
  (jordan)-[:COMMENTED {id: 'd9tcjmf92rs8isaink80', created_at: datetime() - duration('P6D')}]->(c),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isaink8g', created_at: datetime() - duration('P6D'), kind: '*'}]->(c);

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
  (demo)-[:COMMENTED {id: 'd9tcjmf92rs8isainka0', created_at: datetime() - duration('P5D')}]->(c),
  (demo)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainkag', created_at: datetime() - duration('P5D'), kind: '*'}]->(c);

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
  (priya)-[:COMMENTED {id: 'd9tcjmf92rs8isainopg', created_at: datetime() - duration('P1D')}]->(c),
  (priya)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainoq0', created_at: datetime() - duration('P1D'), kind: '*'}]->(c);

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
  (hector)-[:COMMENTED {id: 'd9tcjmf92rs8isainorg', created_at: datetime() - duration('P1D')}]->(c),
  (hector)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainos0', created_at: datetime() - duration('P1D'), kind: '*'}]->(c);

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
  (aisha)-[:COMMENTED {id: 'd9tcjmf92rs8isainotg', created_at: datetime() - duration('P2D')}]->(c),
  (aisha)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainou0', created_at: datetime() - duration('P2D'), kind: '*'}]->(c);

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
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainkc0', created_at: datetime() - duration('P2D')}]->(t),
  (demo)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainkcg', created_at: datetime() - duration('P2D'), kind: '*'}]->(t);

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
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainp40', created_at: datetime() - duration('P1D')}]->(t),
  (demo)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainp4g', created_at: datetime() - duration('P1D'), kind: '*'}]->(t);

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
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainp60', created_at: datetime() - duration('P3D')}]->(t),
  (demo)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainp6g', created_at: datetime() - duration('P3D'), kind: '*'}]->(t);

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
  (demo)-[:CREATED {id: 'd9tcjmf92rs8isainp80', created_at: datetime() - duration('P8D')}]->(t),
  (demo)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainp8g', created_at: datetime() - duration('P8D'), kind: '*'}]->(t);

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
  (jordan)-[:CREATED {id: 'd9tcjmf92rs8isainke0', created_at: datetime() - duration('P4D')}]->(t),
  (jordan)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainkeg', created_at: datetime() - duration('P4D'), kind: '*'}]->(t);

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
  (priya)-[:CREATED {id: 'd9tcjmf92rs8isainpa0', created_at: datetime() - duration('P2D')}]->(t),
  (priya)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainpag', created_at: datetime() - duration('P2D'), kind: '*'}]->(t);

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
  (maya)-[:CREATED {id: 'd9tcjmf92rs8isainpc0', created_at: datetime() - duration('P1D')}]->(t),
  (maya)-[:HAS_PERMISSION {id: 'd9tcjmf92rs8isainpcg', created_at: datetime() - duration('P1D'), kind: '*'}]->(t);
