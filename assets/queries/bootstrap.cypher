// ============================================================================
// Overview
//
// This script creates the initial database schema for the system. It should be
// run once when the system is first installed.
//
// Authorization is scoped ReBAC. There are no system roles and no wildcard
// actions. The Installation node is a singleton scope for installation-level
// grants such as organization.create. Bootstrap does not attach any grants
// to it. Those are created explicitly (for example in demo seed or first-run).
// ============================================================================

// ============================================================================
// Installation
//
// Stable nil xid. Not an Organization: no members, no IN_SCOPE_OF children,
// no default grants (including no organization.create).
// ============================================================================
MERGE (i:Installation {id: '00000000000000000000'})
ON CREATE SET i.created_at = datetime();

// ============================================================================
// Resource types (documentation only)
//
// These nodes describe known labels. They MUST NOT carry permission bindings.
// ============================================================================
UNWIND [
  'Attachment',
  'Comment',
  'Document',
  'Folder',
  'Installation',
  'Issue',
  'Label',
  'Namespace',
  'Organization',
  'Permission',
  'Project',
  'Role',
  'Team',
  'Todo',
  'User'
] AS rt
MERGE (r:ResourceType {id: rt, system: true})
ON CREATE SET r.created_at = datetime();

// ============================================================================
// Constraints and indexes
// ============================================================================

// Resource type
CREATE TEXT INDEX resource_type_id_idx IF NOT EXISTS FOR (n:ResourceType) ON (n.id);
CREATE CONSTRAINT resource_type_id_unique IF NOT EXISTS FOR (n:ResourceType) REQUIRE n.id IS UNIQUE;

// Installation
CREATE TEXT INDEX installation_id_idx IF NOT EXISTS FOR (n:Installation) ON (n.id);
CREATE CONSTRAINT installation_id_unique IF NOT EXISTS FOR (n:Installation) REQUIRE n.id IS UNIQUE;

// Principal (User, Team, Organization)
CREATE TEXT INDEX principal_id_idx IF NOT EXISTS FOR (n:Principal) ON (n.id);
CREATE CONSTRAINT principal_id_unique IF NOT EXISTS FOR (n:Principal) REQUIRE n.id IS UNIQUE;

// Role (per-organization templates, no intrinsic authority)
CREATE TEXT INDEX role_id_idx IF NOT EXISTS FOR (n:Role) ON (n.id);
CREATE TEXT INDEX role_name_idx IF NOT EXISTS FOR (n:Role) ON (n.name);
CREATE TEXT INDEX role_key_idx IF NOT EXISTS FOR (n:Role) ON (n.key);
CREATE CONSTRAINT role_id_unique IF NOT EXISTS FOR (n:Role) REQUIRE n.id IS UNIQUE;

// Team
CREATE TEXT INDEX team_id_idx IF NOT EXISTS FOR (n:Team) ON (n.id);
CREATE TEXT INDEX team_name_idx IF NOT EXISTS FOR (n:Team) ON (n.name);
CREATE CONSTRAINT team_id_unique IF NOT EXISTS FOR (n:Team) REQUIRE n.id IS UNIQUE;

// GRANTED / IN_SCOPE_OF
CREATE CONSTRAINT granted_id_unique IF NOT EXISTS FOR ()-[g:GRANTED]-() REQUIRE g.id IS UNIQUE;
CREATE INDEX granted_role_id_idx IF NOT EXISTS FOR ()-[g:GRANTED]-() ON (g.role_id);
CREATE CONSTRAINT in_scope_of_id_unique IF NOT EXISTS FOR ()-[r:IN_SCOPE_OF]-() REQUIRE r.id IS UNIQUE;

// User
CREATE TEXT INDEX user_id_idx IF NOT EXISTS FOR (n:User) ON (n.id);
CREATE TEXT INDEX user_email_idx IF NOT EXISTS FOR (n:User) ON (n.email);
CREATE CONSTRAINT user_id_unique IF NOT EXISTS FOR (n:User) REQUIRE n.id IS UNIQUE;
CREATE CONSTRAINT user_username_unique IF NOT EXISTS FOR (n:User) REQUIRE n.username IS UNIQUE;
CREATE CONSTRAINT user_email_unique IF NOT EXISTS FOR (n:User) REQUIRE n.email IS UNIQUE;

// Organization
CREATE TEXT INDEX organization_id_idx IF NOT EXISTS FOR (n:Organization) ON (n.id);
CREATE TEXT INDEX organization_name_idx IF NOT EXISTS FOR (n:Organization) ON (n.name);
CREATE TEXT INDEX organization_email_idx IF NOT EXISTS FOR (n:Organization) ON (n.email);
CREATE CONSTRAINT organization_id_unique IF NOT EXISTS FOR (n:Organization) REQUIRE n.id IS UNIQUE;
CREATE CONSTRAINT organization_name_unique IF NOT EXISTS FOR (n:Organization) REQUIRE n.name IS UNIQUE;
CREATE CONSTRAINT organization_email_unique IF NOT EXISTS FOR (n:Organization) REQUIRE n.email IS UNIQUE;

// Namespace
CREATE TEXT INDEX namespace_id_idx IF NOT EXISTS FOR (n:Namespace) ON (n.id);
CREATE TEXT INDEX namespace_name_idx IF NOT EXISTS FOR (n:Namespace) ON (n.name);
CREATE CONSTRAINT namespace_id_unique IF NOT EXISTS FOR (n:Namespace) REQUIRE n.id IS UNIQUE;

// Project
CREATE TEXT INDEX project_id_idx IF NOT EXISTS FOR (n:Project) ON (n.id);
CREATE TEXT INDEX project_key_idx IF NOT EXISTS FOR (n:Project) ON (n.key);
CREATE TEXT INDEX project_name_idx IF NOT EXISTS FOR (n:Project) ON (n.name);
CREATE CONSTRAINT project_id_unique IF NOT EXISTS FOR (n:Project) REQUIRE n.id IS UNIQUE;

// Issue
CREATE TEXT INDEX issue_id_idx IF NOT EXISTS FOR (n:Issue) ON (n.id);
CREATE CONSTRAINT issue_id_unique IF NOT EXISTS FOR (n:Issue) REQUIRE n.id IS UNIQUE;

// Document / Folder
CREATE TEXT INDEX document_id_idx IF NOT EXISTS FOR (n:Document) ON (n.id);
CREATE CONSTRAINT document_id_unique IF NOT EXISTS FOR (n:Document) REQUIRE n.id IS UNIQUE;
CREATE TEXT INDEX folder_id_idx IF NOT EXISTS FOR (n:Folder) ON (n.id);
CREATE CONSTRAINT folder_id_unique IF NOT EXISTS FOR (n:Folder) REQUIRE n.id IS UNIQUE;

// Comment / Attachment / Label / Todo
CREATE TEXT INDEX comment_id_idx IF NOT EXISTS FOR (n:Comment) ON (n.id);
CREATE CONSTRAINT comment_id_unique IF NOT EXISTS FOR (n:Comment) REQUIRE n.id IS UNIQUE;
CREATE TEXT INDEX attachment_id_idx IF NOT EXISTS FOR (n:Attachment) ON (n.id);
CREATE CONSTRAINT attachment_id_unique IF NOT EXISTS FOR (n:Attachment) REQUIRE n.id IS UNIQUE;
CREATE TEXT INDEX label_id_idx IF NOT EXISTS FOR (n:Label) ON (n.id);
CREATE TEXT INDEX label_name_idx IF NOT EXISTS FOR (n:Label) ON (n.name);
CREATE CONSTRAINT label_id_unique IF NOT EXISTS FOR (n:Label) REQUIRE n.id IS UNIQUE;
CREATE TEXT INDEX todo_id_idx IF NOT EXISTS FOR (n:Todo) ON (n.id);
CREATE CONSTRAINT todo_id_unique IF NOT EXISTS FOR (n:Todo) REQUIRE n.id IS UNIQUE;
