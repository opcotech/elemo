import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import { SettingsOrganizationDetailsPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantMembershipToUser,
  grantPermissionToUser,
  grantSystemOwnerMembershipToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api";

test.describe("@settings.organization-members-list Organization Members List E2E Tests", () => {
  let ownerUser: User;
  let writerUser: User;
  let readerUser: User;
  let organizationId: string;

  const getFullName = (user: User) => `${user.first_name} ${user.last_name}`;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    ownerUser = await createUser(testConfig);
    writerUser = await createUser(testConfig);
    readerUser = await createUser(testConfig);

    await grantSystemOwnerMembershipToUser(testConfig, ownerUser.email);

    const apiClient = await createApiClient(
      ownerUser.email,
      USER_DEFAULT_PASSWORD
    );
    const uniqueId = getRandomString(8);
    const organization = await createOrganization(apiClient, {
      name: `Members Org ${uniqueId}`,
      email: `members-${uniqueId}@example.com`,
    });
    organizationId = organization.id;

    await grantMembershipToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId,
      "read"
    );
    await grantPermissionToUser(
      testConfig,
      writerUser.email,
      "Organization",
      organizationId,
      "write"
    );

    await grantMembershipToUser(
      testConfig,
      readerUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      readerUser.email,
      "Organization",
      organizationId,
      "read"
    );
  });

  test("should list organization members for users with read permission", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    await expect(
      orgDetailsPage.members.getRowByMemberName(getFullName(ownerUser))
    ).toBeVisible();
    await expect(
      orgDetailsPage.members.getRowByMemberName(getFullName(writerUser))
    ).toBeVisible();
    await expect(
      orgDetailsPage.members.getRowByMemberName(getFullName(readerUser))
    ).toBeVisible();
  });

  test("should show invite member button only for members with write permission", async ({
    page,
  }) => {
    // Writer with write permission sees the invite button
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();
    expect(await orgDetailsPage.members.hasInviteMemberButton()).toBeTruthy();

    // Reader with read-only permission does not see the invite button
    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();
    expect(await orgDetailsPage.members.hasInviteMemberButton()).toBeFalsy();
  });

  test("should display member status badges correctly", async ({ page }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    // Verify active members show active status
    const ownerRow = orgDetailsPage.members.getRowByMemberName(
      getFullName(ownerUser)
    );
    await expect(ownerRow).toBeVisible();
    await expect(ownerRow.getByText("Active")).toBeVisible();
  });

  test("should display member email addresses in the list", async ({
    page,
  }) => {
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    const ownerRow = orgDetailsPage.members.getRowByMemberName(
      getFullName(ownerUser)
    );
    await expect(ownerRow.getByText(ownerUser.email)).toBeVisible();
  });
});
