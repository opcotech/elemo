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
