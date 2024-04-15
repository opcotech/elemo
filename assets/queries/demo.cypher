// ============================================================================
// Overview
//
// This script creates a demo organization with some users and roles for testing
// purposes.
// ============================================================================

// Demo Organization's owner
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

// Demo Organization
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
  (u)-[:MEMBER_OF {id: '9bsv0s3n4ccg0329pecg', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: '9bsv0s613svg02gik0r0', created_at: datetime(), kind: '*'}]->(o);

// Add random user with read permission
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
  }
WITH u
MATCH (o:Organization {id: '9bsv0s4vl6gg02sv7jrg'})
CREATE
  (u)-[:MEMBER_OF {id: '9bsv0s314mtg02goain0', created_at: datetime()}]->(o),
  (u)-[:HAS_PERMISSION {id: '9bsv0s314mtg02goaing', created_at: datetime(), kind: 'read'}]->(o);
