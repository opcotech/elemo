import {
  addTeamMember,
  createOrganization,
  createTeam,
  listTeamMembers,
  removeTeamMember,
} from "./api";
import { expect, test } from "./fixtures";
import {
  clickUntilURL,
  clickUntilVisible,
  waitForSuccessToast,
} from "./helpers";
import { SettingsOrganizationDetailsPage } from "./pages";
import { USER_DEFAULT_PASSWORD, loginUser } from "./utils/auth";
import {
  createUser,
  grantMembershipToUser,
  grantOrganizationCreateToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import { v1OrganizationTeamsCreate } from "@/lib/api/sdk";
import type { User } from "@/lib/api/types";

test.describe("@settings.organization-teams Organization Team Members E2E Tests", () => {
  let owner: User;
  let member1: User;
  let member2: User;
  let organizationId: string;
  let organizationSlug: string;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    owner = await createUser(testConfig);
    member1 = await createUser(testConfig);
    member2 = await createUser(testConfig);

    await grantOrganizationCreateToUser(testConfig, owner.email);

    const uniqueId = getRandomString(8);
    const apiClient = await createApiClient(owner.email, USER_DEFAULT_PASSWORD);
    const organization = await createOrganization(apiClient, {
      name: `Test Org Teams ${uniqueId}`,
      email: `test-teams-${uniqueId}@example.com`,
    });
    organizationId = organization.id;
    organizationSlug = organization.slug;

    await grantMembershipToUser(
      testConfig,
      member1.email,
      "Organization",
      organizationId
    );
    await grantMembershipToUser(
      testConfig,
      member2.email,
      "Organization",
      organizationId
    );
  });

  test("should create a team and add members via the organization teams API", async ({
    createApiClient,
  }) => {
    const apiClient = await createApiClient(owner.email, USER_DEFAULT_PASSWORD);
    const team = await createTeam(apiClient, organizationId, {
      name: `Platform ${getRandomString(8)}`,
      description: "Platform engineering team.",
    });

    expect(team.id).toBeTruthy();
    expect(team.name).toContain("Platform");

    await addTeamMember(apiClient, organizationId, team.id, member1.id);
    await addTeamMember(apiClient, organizationId, team.id, member2.id);

    const members = await listTeamMembers(apiClient, organizationId, team.id);
    const memberIds = members.map((user) => user.id);
    expect(memberIds).toContain(member1.id);
    expect(memberIds).toContain(member2.id);
  });

  test("should remove a team member", async ({ createApiClient }) => {
    const apiClient = await createApiClient(owner.email, USER_DEFAULT_PASSWORD);
    const team = await createTeam(apiClient, organizationId, {
      name: `Remove ${getRandomString(8)}`,
    });

    await addTeamMember(apiClient, organizationId, team.id, member1.id);
    let members = await listTeamMembers(apiClient, organizationId, team.id);
    expect(members.map((user) => user.id)).toContain(member1.id);

    await removeTeamMember(apiClient, organizationId, team.id, member1.id);
    members = await listTeamMembers(apiClient, organizationId, team.id);
    expect(members.map((user) => user.id)).not.toContain(member1.id);
  });

  test("should deny team management without team.manage", async ({
    createApiClient,
  }) => {
    const memberClient = await createApiClient(
      member1.email,
      USER_DEFAULT_PASSWORD
    );

    const result = await v1OrganizationTeamsCreate({
      client: memberClient,
      path: { organizationRef: organizationId },
      body: { name: `Denied ${getRandomString(8)}` },
    });

    expect(result.error).toBeTruthy();
    expect(result.response?.status).toBe(403);
  });

  test("should create a team from organization settings and add a member", async ({
    page,
    createApiClient,
  }) => {
    await loginUser(page, {
      email: owner.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);

    const teams = page.locator("[data-section='teams']");
    await expect(teams).toBeVisible();
    await clickUntilURL(
      teams.getByRole("link", { name: /create team/i }).first(),
      new RegExp(`/settings/organizations/${organizationSlug}/teams/new`)
    );

    const teamName = `UI Team ${getRandomString(8)}`;
    const form = page.locator("[data-section='team-create-form']");
    await form.getByRole("textbox", { name: "Name" }).fill(teamName);
    await form.getByRole("button", { name: "Create Team" }).click();
    await waitForSuccessToast(page, "created");
    await expect(page).toHaveURL(
      new RegExp(`/settings/organizations/${organizationSlug}/teams/[^/]+/edit`)
    );

    const members = page.locator("[data-section='team-members']");
    await expect(members).toBeVisible();
    const dialog = page.getByRole("dialog");
    await clickUntilVisible(
      members.getByRole("button", { name: /add member/i }).first(),
      dialog
    );
    await dialog.getByRole("combobox").click();
    const member1Name = `${member1.first_name} ${member1.last_name}`;
    await page.getByRole("option", { name: member1Name }).click();
    await dialog.getByRole("button", { name: /add member/i }).click();
    await waitForSuccessToast(page, "added");

    await expect(members.getByText(member1Name)).toBeVisible();

    const apiClient = await createApiClient(owner.email, USER_DEFAULT_PASSWORD);
    const listed = await listTeamMembers(
      apiClient,
      organizationId,
      page.url().split("/teams/")[1].split("/")[0]
    );
    expect(listed.map((user) => user.id)).toContain(member1.id);
  });

  test("should not show create team without team.manage", async ({ page }) => {
    await loginUser(page, {
      email: member1.email,
      password: USER_DEFAULT_PASSWORD,
    });

    const orgDetailsPage = new SettingsOrganizationDetailsPage(page);
    await orgDetailsPage.goto(organizationSlug);
    const teams = page.locator("[data-section='teams']");
    await expect(teams).toBeVisible();
    await expect(
      teams.getByRole("link", { name: /create team/i })
    ).not.toBeVisible();
  });
});
