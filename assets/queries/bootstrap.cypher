// ============================================================================
// Overview
//
// This script creates the initial database schema for the system. It should be
// run once when the system is first installed.
//
// Some resources are system resources, which means they are not created by
// users. They are created by this script and have intentionally invalid IDs to
// prevent users from reading or writing them.
// ============================================================================

// ============================================================================
// Create system resource types
// ============================================================================
UNWIND [
  'Attachment',
  'Comment',
  'Document',
  'Issue',
  'Label',
  'Namespace',
  'Organization',
  'Project',
  'Role',
  'Todo',
  'User'
] AS rt
CREATE (:ResourceType {id: rt, system: true, created_at: datetime()});

// ============================================================================
// Create system roles to manage resources
// ============================================================================

// Create roles
UNWIND ['Owner', 'Admin', 'Support'] AS r
CREATE (sr:Role {id: r, name: r, system: true, created_at: datetime()});

// Create role bindings
UNWIND [
  'Attachment',
  'Comment',
  'Document',
  'Issue',
  'Label',
  'Namespace',
  'Organization',
  'Project',
  'Role',
  'Todo',
  'User'
] AS t
UNWIND [
  ['Owner', '*'],
  ['Admin', 'create', 'read', 'write'],
  ['Support', 'read', 'write']
] AS bindings
WITH t, bindings[0] AS role, bindings[1..] AS permissions
MATCH (rt:ResourceType {id: t}), (r:Role {id: role})
WITH rt, r, permissions
UNWIND permissions AS permission
CREATE (r)-[:HAS_PERMISSION {kind: permission, created_at: datetime()}]->(rt);

// ============================================================================
// Non-system resources
// ============================================================================

// Resource type index
CREATE TEXT INDEX resource_type_id_idx IF NOT EXISTS FOR (n:ResourceType) ON (n.id);

// Role index
CREATE TEXT INDEX role_id_idx IF NOT EXISTS FOR (n:Role) ON (n.id);

// User index
CREATE TEXT INDEX user_id_idx IF NOT EXISTS FOR (n:User) ON (n.id);

CREATE TEXT INDEX user_email_idx IF NOT EXISTS FOR (n:User) ON (n.email);
