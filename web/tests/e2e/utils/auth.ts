import { getSession } from "./db";
import { generateXid } from "./xid";
import { getRandomString } from "./random";
import type { User, UserStatus } from "@/lib/api";

export const USER_DEFAULT_PASSWORD = "AppleTree123";
export const USER_DEFAULT_PASSWORD_HASH =
  "$2a$10$CYRw/WtES8e8d4di2uIddevV9MO2.tI0G8R8QZEnF5dyx8S4Wqt6e"; // 'AppleTree123'

export async function createDBUser(
  status: UserStatus = "active",
  overrides?: Partial<User>
): Promise<User> {
  const user: User = {
    id: await generateXid(),
    username: getRandomString(8).toLowerCase(),
    first_name: getRandomString(),
    last_name: getRandomString(),
    email: `${getRandomString()}-test@example.com`.toLowerCase(),
    picture: "https://picsum.photos/id/64/100/100",
    title: "Senior Test User",
    bio: "I am a senior test user",
    address: "123 Main St, Anytown, USA",
    phone: "555-555-5555",
    links: ["https://example.com"],
    languages: ["en"],
    status: status,
    created_at: new Date().toISOString(),
    updated_at: null,
    ...overrides,
  };

  try {
    const resp = await getSession().executeWrite((tx: any) => {
      const query = `
      CREATE (u:User $user)
      WITH u SET u.password = $password
      RETURN u
    `;

      return tx.run(query, { user, password: USER_DEFAULT_PASSWORD_HASH });
    });

    return resp.records[0].get("u").properties;
  } finally {
    await getSession().close();
  }
}
