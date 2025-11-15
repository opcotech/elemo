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
MERGE (r:ResourceType {id: rt, system: true})
ON CREATE SET r.created_at = datetime();

// ============================================================================
// Create system roles to manage resources
//
// The users with system roles assigned are considered superusers. However,
// they shouldn't access __ALL__ resources. Some resources may contain private
// or sensitive data to the individual user. For example, a user may create a
// TODO item for themselves that should not be visible to other users.
//
// All other resources are considered public within the platform and can be
// accessed by all users (with necessary permissions).
// ============================================================================

// Create roles
UNWIND ['Owner', 'Admin', 'Support'] AS r
MERGE (sr:Role {id: r, name: r, system: true})
ON CREATE SET sr.created_at = datetime();

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
  'User'
] AS t
UNWIND [
  ['Owner', '*'],
  ['Admin', 'create', 'read', 'write'],
  ['Support', 'read', 'write']
] AS bindings
WITH t, bindings[0] AS role, bindings[1..] AS permissions
MATCH (rt:ResourceType {id: t})
OPTIONAL MATCH (r:Role {id: role, system: true})
WITH rt, r, permissions
UNWIND permissions AS permission
MERGE (r)-[p:HAS_PERMISSION {kind: permission}]->(rt)
ON CREATE SET p.created_at = datetime();

// ============================================================================
// Non-system resources
// ============================================================================

// Resource type index
CREATE TEXT INDEX resource_type_id_idx IF NOT EXISTS FOR (n:ResourceType) ON (n.id);
CREATE CONSTRAINT resource_type_id_unique IF NOT EXISTS FOR (n:ResourceType) REQUIRE n.id IS UNIQUE;

// Role index
CREATE TEXT INDEX role_id_idx IF NOT EXISTS FOR (n:Role) ON (n.id);
CREATE TEXT INDEX role_name_idx IF NOT EXISTS FOR (n:Role) ON (n.name);
CREATE CONSTRAINT role_id_unique IF NOT EXISTS FOR (n:Role) REQUIRE n.id IS UNIQUE;

// Permission index
CREATE INDEX has_permission_kind_idx IF NOT EXISTS FOR (r:HAS_PERMISSION) ON (r.kind);

// User index
CREATE TEXT INDEX user_id_idx IF NOT EXISTS FOR (n:User) ON (n.id);
CREATE TEXT INDEX user_email_idx IF NOT EXISTS FOR (n:User) ON (n.email);
CREATE CONSTRAINT user_id_unique IF NOT EXISTS FOR (n:User) REQUIRE n.id IS UNIQUE;
CREATE CONSTRAINT user_username_unique IF NOT EXISTS FOR (n:User) REQUIRE n.username IS UNIQUE;
CREATE CONSTRAINT user_email_unique IF NOT EXISTS FOR (n:User) REQUIRE n.email IS UNIQUE;

// Organization index
CREATE TEXT INDEX organization_id_idx IF NOT EXISTS FOR (n:Organization) ON (n.id);
CREATE TEXT INDEX organization_name_idx IF NOT EXISTS FOR (n:Organization) ON (n.name);
CREATE TEXT INDEX organization_email_idx IF NOT EXISTS FOR (n:Organization) ON (n.email);
CREATE CONSTRAINT organization_id_unique IF NOT EXISTS FOR (n:Organization) REQUIRE n.id IS UNIQUE;
CREATE CONSTRAINT organization_name_unique IF NOT EXISTS FOR (n:Organization) REQUIRE n.name IS UNIQUE;
CREATE CONSTRAINT organization_email_unique IF NOT EXISTS FOR (n:Organization) REQUIRE n.email IS UNIQUE;
