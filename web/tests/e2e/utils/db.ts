import dotenv from "dotenv";
import neo4j from "neo4j-driver";
import type { Session } from "neo4j-driver";

dotenv.config({ path: ".env.test.local" });

const driver = neo4j.driver(
  process.env.NEO4J_URL || "neo4j://localhost:7687",
  neo4j.auth.basic(
    process.env.NEO4J_USER || "neo4j",
    process.env.NEO4J_PASSWORD || "neo4jsecret"
  )
);

// Get a session from the driver.
export function getSession(): Session {
  return driver.session();
}
