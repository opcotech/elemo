import { test as base } from "@playwright/test";

import { USER_DEFAULT_PASSWORD } from "../utils/auth";
import { createUser, grantOrganizationCreateToUser } from "../utils/db";
import { getTestConfig } from "../utils/test-config";

import type { User } from "@/lib/api/types";

export interface TestPersona {
  user: User;
  credentials: {
    email: string;
    password: string;
  };
}

type PersonaFixtures = {
  userPersona: TestPersona;
  ownerPersona: TestPersona;
};

async function createPersona(owner: boolean): Promise<TestPersona> {
  const testConfig = getTestConfig();
  const user = await createUser(testConfig);

  if (owner) {
    await grantOrganizationCreateToUser(testConfig, user.email);
  }

  return {
    user,
    credentials: {
      email: user.email,
      password: USER_DEFAULT_PASSWORD,
    },
  };
}

/**
 * Personas are created lazily and uniquely for each test that requests them.
 * This avoids shared authentication state while centralizing common bootstrap.
 */
export const test = base.extend<PersonaFixtures>({
  // eslint-disable-next-line no-empty-pattern
  userPersona: async ({}, use) => {
    await use(await createPersona(false));
  },
  // eslint-disable-next-line no-empty-pattern
  ownerPersona: async ({}, use) => {
    await use(await createPersona(true));
  },
});

export { expect } from "@playwright/test";
