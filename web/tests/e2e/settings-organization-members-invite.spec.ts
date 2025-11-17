import { createOrganization } from "./api";
import { expect, test } from "./fixtures";
import { waitForSuccessToast } from "./helpers";
import {
  OrganizationsJoinPage,
  SettingsOrganizationDetailsPage,
} from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantMembershipToUser,
  grantPermissionToUser,
  grantSystemOwnerMembershipToUser,
} from "./utils/db";
import { getInvitationTokenFromEmail, waitForEmail } from "./utils/mailpit";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api";

test.describe("@settings.organization-members-invite Organization Members Invite E2E Tests", () => {
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

  test("should allow members with write permission to invite a new member", async ({
    page,
    testConfig,
  }) => {
    const inviteeUser = await createUser(testConfig);

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    await orgDetailsPage.members.clickInviteMemberButton();
    const inviteDialog = page.getByRole("dialog", { name: "Invite Member" });
    await inviteDialog
      .getByLabel("Email Address")
      .fill(inviteeUser.email.toLowerCase());
    await inviteDialog
      .getByRole("button", { name: /send invitation/i })
      .click();

    await waitForSuccessToast(page, "Invitation sent");
    await orgDetailsPage.members.waitForLoad();
    await expect(
      orgDetailsPage.members.getRowByMemberName(getFullName(inviteeUser))
    ).toBeVisible();
  });

  test("should allow invited members to join organization using email link", async ({
    page,
    testConfig,
  }) => {
    const invitedUser = await createUser(testConfig);

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();
    await orgDetailsPage.members.clickInviteMemberButton();

    const inviteDialog = page.getByRole("dialog", { name: "Invite Member" });
    await inviteDialog
      .getByLabel("Email Address")
      .fill(invitedUser.email.toLowerCase());
    await inviteDialog
      .getByRole("button", { name: /send invitation/i })
      .click();
    await waitForSuccessToast(page, "Invitation sent");

    await waitForEmail(invitedUser.email, 20000);
    const token = await getInvitationTokenFromEmail(invitedUser.email);
    expect(token).toBeTruthy();

    const joinPage = new OrganizationsJoinPage(page);
    await joinPage.goto(organizationId, token!);
    await joinPage.acceptInvitation({ password: USER_DEFAULT_PASSWORD });

    await loginUser(page, {
      email: invitedUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    await expect(
      orgDetailsPage.members.getRowByMemberName(getFullName(invitedUser))
    ).toBeVisible();
  });

  test("should show pending status for invited members", async ({
    page,
    testConfig,
  }) => {
    const inviteeUser = await createUser(testConfig);

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    await orgDetailsPage.members.clickInviteMemberButton();
    const inviteDialog = page.getByRole("dialog", { name: "Invite Member" });
    await inviteDialog
      .getByLabel("Email Address")
      .fill(inviteeUser.email.toLowerCase());
    await inviteDialog
      .getByRole("button", { name: /send invitation/i })
      .click();
    await waitForSuccessToast(page, "Invitation sent");

    await orgDetailsPage.members.waitForLoad();
    const memberRow = orgDetailsPage.members.getRowByMemberName(
      getFullName(inviteeUser)
    );
    await expect(memberRow).toBeVisible();
    await expect(memberRow.getByText(/pending/i)).toBeVisible();
  });

  test("should allow inviting member with optional role assignment", async ({
    page,
    testConfig,
  }) => {
    const inviteeUser = await createUser(testConfig);

    await loginUser(page, {
      email: writerUser.email,
      password: USER_DEFAULT_PASSWORD,
    });
    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationId);
    await orgDetailsPage.members.waitForLoad();

    await orgDetailsPage.members.clickInviteMemberButton();
    const inviteDialog = page.getByRole("dialog", { name: "Invite Member" });
    await inviteDialog
      .getByLabel("Email Address")
      .fill(inviteeUser.email.toLowerCase());
    // Role selection is optional, so we can submit without it
    await inviteDialog
      .getByRole("button", { name: /send invitation/i })
      .click();
    await waitForSuccessToast(page, "Invitation sent");

    await orgDetailsPage.members.waitForLoad();
    await expect(
      orgDetailsPage.members.getRowByMemberName(getFullName(inviteeUser))
    ).toBeVisible();
  });
});
