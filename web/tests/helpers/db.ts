import dotenv from 'dotenv';
import neo4j, { type Session } from 'neo4j-driver';

dotenv.config({ path: '.env.test' });

let driver = neo4j.driver(
  process.env.NEO4J_URL || '',
  neo4j.auth.basic(process.env.NEO4J_USER || 'neo4j', process.env.NEO4J_PASSWORD || '')
);

// Get a session from the driver.
export function getSession(): Session {
  return driver.session();
}
