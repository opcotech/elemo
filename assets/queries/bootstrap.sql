/* ============================================================================
 * Overview
 *
 * This script creates the initial relational database schema. It should be run
 * once when the system is first installed.
 */
-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
  id VARCHAR(35) PRIMARY KEY,
  title VARCHAR NOT NULL CONSTRAINT notifications_title_length CHECK (LENGTH (title)<=120),
  description TEXT CONSTRAINT notifications_description_length CHECK (LENGTH (title)<=500),
  recipient VARCHAR(35) NOT NULL,
  read BOOLEAN NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP
);

-- User tokens table
CREATE TABLE IF NOT EXISTS user_tokens (
  id VARCHAR(35) PRIMARY KEY,
  user_id VARCHAR(35) NOT NULL,
  sent_to TEXT NOT NULL,
  token CHARACTER VARYING(72) NOT NULL,
  context CHARACTER VARYING(14) CHECK (context IN ('confirm', 'reset_password', 'invite')) NOT NULL,
  created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);

-- Indexes for performance optimization
CREATE INDEX IF NOT EXISTS user_tokens_sent_to_index ON user_tokens USING btree (sent_to);
CREATE INDEX IF NOT EXISTS user_tokens_context_index ON user_tokens USING btree (context);
CREATE UNIQUE INDEX IF NOT EXISTS user_tokens_sent_to_context_idx ON user_tokens (sent_to, context);

-- Custom field definitions (typed attributes scoped to org/namespace/project)
CREATE TABLE IF NOT EXISTS custom_field_definitions (
  id VARCHAR(64) PRIMARY KEY,
  field_key VARCHAR(63) NOT NULL,
  name VARCHAR(120) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  kind VARCHAR(32) NOT NULL,
  scope_id VARCHAR(64) NOT NULL,
  target_type VARCHAR(32) NOT NULL,
  required BOOLEAN NOT NULL DEFAULT FALSE,
  archived BOOLEAN NOT NULL DEFAULT FALSE,
  index_exact BOOLEAN NOT NULL DEFAULT FALSE,
  index_range BOOLEAN NOT NULL DEFAULT FALSE,
  index_fulltext BOOLEAN NOT NULL DEFAULT FALSE,
  owner_user_id VARCHAR(64) NOT NULL,
  registrar_client_id VARCHAR(128) NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ,
  CONSTRAINT custom_field_definitions_key_format CHECK (field_key ~ '^[a-z][a-z0-9_]{1,62}$'),
  CONSTRAINT custom_field_definitions_name_length CHECK (char_length(name) BETWEEN 3 AND 120),
  CONSTRAINT custom_field_definitions_kind_known CHECK (kind IN (
    'text', 'integer', 'decimal', 'boolean', 'date', 'datetime', 'url',
    'single_select', 'multi_select', 'user_reference', 'resource_reference'
  ))
);

CREATE UNIQUE INDEX IF NOT EXISTS custom_field_definitions_scope_target_key
  ON custom_field_definitions (scope_id, target_type, field_key);

CREATE INDEX IF NOT EXISTS custom_field_definitions_scope_target
  ON custom_field_definitions (scope_id, target_type);

CREATE TABLE IF NOT EXISTS custom_field_schema_text (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  min_length INTEGER,
  max_length INTEGER,
  pattern TEXT NOT NULL DEFAULT '',
  CONSTRAINT custom_field_schema_text_bounds CHECK (
    (min_length IS NULL OR min_length >= 0)
    AND (max_length IS NULL OR max_length >= 0)
    AND (min_length IS NULL OR max_length IS NULL OR min_length <= max_length)
  )
);

CREATE TABLE IF NOT EXISTS custom_field_schema_integer (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  min_value BIGINT,
  max_value BIGINT,
  CONSTRAINT custom_field_schema_integer_bounds CHECK (
    min_value IS NULL OR max_value IS NULL OR min_value <= max_value
  )
);

CREATE TABLE IF NOT EXISTS custom_field_schema_decimal (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  min_value NUMERIC,
  max_value NUMERIC,
  scale INTEGER,
  CONSTRAINT custom_field_schema_decimal_bounds CHECK (
    (scale IS NULL OR (scale >= 0 AND scale <= 38))
    AND (min_value IS NULL OR max_value IS NULL OR min_value <= max_value)
  )
);

CREATE TABLE IF NOT EXISTS custom_field_schema_boolean (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS custom_field_schema_date (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  min_value DATE,
  max_value DATE,
  CONSTRAINT custom_field_schema_date_bounds CHECK (
    min_value IS NULL OR max_value IS NULL OR min_value <= max_value
  )
);

CREATE TABLE IF NOT EXISTS custom_field_schema_datetime (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  min_value TIMESTAMPTZ,
  max_value TIMESTAMPTZ,
  CONSTRAINT custom_field_schema_datetime_bounds CHECK (
    min_value IS NULL OR max_value IS NULL OR min_value <= max_value
  )
);

CREATE TABLE IF NOT EXISTS custom_field_schema_url (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  allowed_schemes TEXT[] NOT NULL
);

CREATE TABLE IF NOT EXISTS custom_field_schema_select (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS custom_field_schema_user_reference (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  multiple BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS custom_field_schema_resource_reference (
  definition_id VARCHAR(64) PRIMARY KEY REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  allowed_types TEXT[] NOT NULL,
  multiple BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS custom_field_options (
  id VARCHAR(20) PRIMARY KEY,
  definition_id VARCHAR(64) NOT NULL REFERENCES custom_field_definitions (id) ON DELETE CASCADE,
  option_key VARCHAR(63) NOT NULL,
  label VARCHAR(120) NOT NULL,
  color VARCHAR(32) NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  disabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT custom_field_options_key_format CHECK (option_key ~ '^[a-z][a-z0-9_]{1,62}$'),
  UNIQUE (definition_id, option_key)
);

CREATE TABLE IF NOT EXISTS custom_field_values (
  id VARCHAR(20) PRIMARY KEY,
  definition_id VARCHAR(64) NOT NULL REFERENCES custom_field_definitions (id) ON DELETE RESTRICT,
  resource_id VARCHAR(64) NOT NULL,
  resource_type VARCHAR(32) NOT NULL,
  ordinal INTEGER NOT NULL DEFAULT 0,
  committed BOOLEAN NOT NULL DEFAULT TRUE,
  text_value TEXT,
  integer_value BIGINT,
  decimal_value NUMERIC,
  boolean_value BOOLEAN,
  date_value DATE,
  datetime_value TIMESTAMPTZ,
  url_value TEXT,
  option_key VARCHAR(63),
  user_id VARCHAR(64),
  ref_resource_id VARCHAR(64),
  index_exact BOOLEAN NOT NULL DEFAULT FALSE,
  index_range BOOLEAN NOT NULL DEFAULT FALSE,
  index_fulltext BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ,
  CONSTRAINT custom_field_values_one_representation CHECK (
    (
      (text_value IS NOT NULL)::INTEGER
      + (integer_value IS NOT NULL)::INTEGER
      + (decimal_value IS NOT NULL)::INTEGER
      + (boolean_value IS NOT NULL)::INTEGER
      + (date_value IS NOT NULL)::INTEGER
      + (datetime_value IS NOT NULL)::INTEGER
      + (url_value IS NOT NULL)::INTEGER
      + (option_key IS NOT NULL)::INTEGER
      + (user_id IS NOT NULL)::INTEGER
      + (ref_resource_id IS NOT NULL)::INTEGER
    ) = 1
  ),
  UNIQUE (definition_id, resource_id, ordinal)
);

CREATE INDEX IF NOT EXISTS custom_field_values_resource
  ON custom_field_values (resource_id);

CREATE INDEX IF NOT EXISTS custom_field_values_definition_resource
  ON custom_field_values (definition_id, resource_id);

CREATE INDEX IF NOT EXISTS custom_field_values_text_exact
  ON custom_field_values (definition_id, text_value)
  WHERE committed AND index_exact AND text_value IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_integer_range
  ON custom_field_values (definition_id, integer_value)
  WHERE committed AND (index_exact OR index_range) AND integer_value IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_decimal_range
  ON custom_field_values (definition_id, decimal_value)
  WHERE committed AND (index_exact OR index_range) AND decimal_value IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_boolean_exact
  ON custom_field_values (definition_id, boolean_value)
  WHERE committed AND index_exact AND boolean_value IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_date_range
  ON custom_field_values (definition_id, date_value)
  WHERE committed AND (index_exact OR index_range) AND date_value IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_datetime_range
  ON custom_field_values (definition_id, datetime_value)
  WHERE committed AND (index_exact OR index_range) AND datetime_value IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_url_exact
  ON custom_field_values (definition_id, url_value)
  WHERE committed AND index_exact AND url_value IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_option_exact
  ON custom_field_values (definition_id, option_key)
  WHERE committed AND index_exact AND option_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_user_exact
  ON custom_field_values (definition_id, user_id)
  WHERE committed AND index_exact AND user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_ref_exact
  ON custom_field_values (definition_id, ref_resource_id)
  WHERE committed AND index_exact AND ref_resource_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS custom_field_values_text_fulltext
  ON custom_field_values USING GIN (to_tsvector('simple', text_value))
  WHERE committed AND index_fulltext AND text_value IS NOT NULL;

CREATE TABLE IF NOT EXISTS custom_field_operations (
  id VARCHAR(20) PRIMARY KEY,
  kind VARCHAR(32) NOT NULL,
  status VARCHAR(16) NOT NULL,
  resource_id VARCHAR(64) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ,
  CONSTRAINT custom_field_operations_kind_known CHECK (kind IN ('stage_values', 'delete_resource')),
  CONSTRAINT custom_field_operations_status_known CHECK (status IN ('pending', 'committed', 'aborted'))
);

CREATE INDEX IF NOT EXISTS custom_field_operations_pending
  ON custom_field_operations (created_at)
  WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS custom_field_operations_resource
  ON custom_field_operations (resource_id, status);

