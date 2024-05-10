/* ============================================================================
 * Overview
 *
 * This script creates the initial relational database schema. It should be run
 * once when the system is first installed.
 */
-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id          VARCHAR(20) PRIMARY KEY,
    title       VARCHAR     NOT NULL CONSTRAINT notifications_title_length CHECK (LENGTH(title) <= 120),
    description TEXT        CONSTRAINT notifications_description_length CHECK (LENGTH(title) <= 500),
    recipient   VARCHAR(20) NOT NULL,
    read        BOOLEAN     NOT NULL,
    created_at  TIMESTAMP   NOT NULL,
    updated_at  TIMESTAMP
);
