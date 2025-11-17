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

test.describe("@settings.organization-members-remove Organization Members Remove E2E Tests", () => {
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

  test("should allow members with write permission to remove an existing member", async ({
    page,
    testConfig,
  }) => {
    const removableUser = await createUser(testConfig);
    await grantMembershipToUser(
      testConfig,
      removableUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      removableUser.email,
      "Organization",
      organizationId,
      "read"
    );

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    await expect(
      orgDetailsPage.members.getRowByMemberName(getFullName(removableUser))
    ).toBeVisible();

    await orgDetailsPage.members.removeMember(getFullName(removableUser));
    await expect(
      orgDetailsPage.members.getRowByMemberName(getFullName(removableUser))
    ).not.toBeVisible();
  });

  test("should show remove button only for members with write permission", async ({
    page,
    testConfig,
  }) => {
    const removableUser = await createUser(testConfig);
    await grantMembershipToUser(
      testConfig,
      removableUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      removableUser.email,
      "Organization",
      organizationId,
      "read"
    );

    // Writer with write permission should see remove button
    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    const memberRow = orgDetailsPage.members.getRowByMemberName(
      getFullName(removableUser)
    );
    await expect(memberRow).toBeVisible();
    // Verify remove button is visible by checking if we can open the dialog
    const removeButton = memberRow.getByRole("button", {
      name: /remove member/i,
    });
    await expect(removeButton).toBeVisible();

    // Reader with read-only permission should not see remove button
    await loginUser(page, {
      email: readerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    const readerMemberRow = orgDetailsPage.members.getRowByMemberName(
      getFullName(removableUser)
    );
    await expect(readerMemberRow).toBeVisible();
    await expect(
      readerMemberRow.getByRole("button", { name: /remove member/i })
    ).not.toBeVisible();
  });

  test("should display confirmation dialog when removing a member", async ({
    page,
    testConfig,
  }) => {
    const removableUser = await createUser(testConfig);
    await grantMembershipToUser(
      testConfig,
      removableUser.email,
      "Organization",
      organizationId
    );
    await grantPermissionToUser(
      testConfig,
      removableUser.email,
      "Organization",
      organizationId,
      "read"
    );

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    await orgDetailsPage.members.openRemoveMemberDialog(
      getFullName(removableUser)
    );

    const dialog = page.getByRole("alertdialog", {
      name: /Remove Member from Organization/i,
    });
    await expect(dialog).toBeVisible();
    await expect(
      dialog.getByText(/Are you sure you want to remove this member/i)
    ).toBeVisible();
  });
});
