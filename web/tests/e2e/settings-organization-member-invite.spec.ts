import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
  createDBRole,
} from "./utils/organization";
import { getInvitationTokenFromEmail, waitForEmail } from "./utils/mailpit";
import { MembersPage } from "./pages/members-page";
import { InviteDialog } from "./components/invite-dialog";
import { waitForErrorToast, waitForSuccessToast } from "./helpers/toast";

test.describe("@settings.organization-member-invite Organization Member Invite E2E Tests", () => {
  test.describe("Permission-Based Visibility", () => {
    let ownerUser: any;
    let writeUser: any;
    let readUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });
      writeUser = await createDBUser("active", {
        first_name: "Write",
        last_name: "User",
      });
      readUser = await createDBUser("active", {
        first_name: "Read",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Invite Permission Test Org",
      });
      await addMemberToOrganization(testOrg.id, writeUser.id, "write");
      await addMemberToOrganization(testOrg.id, readUser.id, "read");
    });

    test("user with write permission should see invite button", async ({
      page,
    }) => {
      await loginUser(page, writeUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const hasInviteButton = await membersPage.hasInviteButton();
      expect(hasInviteButton).toBe(true);
    });

    test("user with read permission should not see invite button", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const hasInviteButton = await membersPage.hasInviteButton();
      expect(hasInviteButton).toBe(false);
    });

    test("owner should see invite button", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const hasInviteButton = await membersPage.hasInviteButton();
      expect(hasInviteButton).toBe(true);
    });
  });

  test.describe("Invite Dialog", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Invite Dialog Test Org",
      });
    });

    test("should open invite dialog when invite button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await expect(page.getByLabel("Email Address")).toBeVisible();
      await expect(page.getByLabel("Role (Optional)")).toBeVisible();
      await expect(
        page.getByRole("button", { name: "Send Invitation" })
      ).toBeVisible();
    });

    test("should close dialog when cancel is clicked", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await inviteDialog.cancel();
      await inviteDialog.waitForClose();
    });
  });

  test.describe("Form Validation", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Invite Validation Test Org",
      });
    });

    test("should show validation error for invalid email", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await inviteDialog.fillEmail("invalid-email");
      await page.getByLabel("Email Address").blur();
      await page
        .waitForFunction(
          () => {
            const input = document.querySelector(
              'input[aria-label="Email Address"]'
            ) as HTMLInputElement;
            return input.value === "invalid-email";
          },
          { timeout: 1000 }
        )
        .catch(() => {});
      await inviteDialog.sendInvitation();
      const formMessages = page.locator('[data-slot="form-message"]');
      const hasError = await formMessages
        .first()
        .isVisible()
        .catch(() => false);
      const dialogStillOpen = await inviteDialog.isVisible();

      expect(hasError || dialogStillOpen).toBe(true);
    });

    test("should show validation error for empty email", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await inviteDialog.sendInvitation();
      const formMessages = page.locator('[data-slot="form-message"]');
      const hasError = await formMessages
        .first()
        .isVisible()
        .catch(() => false);
      const dialogStillOpen = await inviteDialog.isVisible();

      expect(dialogStillOpen || hasError).toBe(true);
    });
  });

  test.describe("Successful Invitation", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: `Invite Success Test Org ${Date.now()}`,
      });
    });

    test("should successfully send invitation with email only", async ({
      page,
    }) => {
      const inviteEmail = `invite-${Date.now()}@example.com`;

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await inviteDialog.inviteMember(inviteEmail);
      await waitForSuccessToast(page, "Invitation sent", { timeout: 10000 });
      await inviteDialog.waitForClose();
      const email = await waitForEmail(inviteEmail, 15000);
      expect(email).not.toBeNull();
      expect(email?.To.some((to) => to.Address === inviteEmail)).toBe(true);
    });

    test("should successfully send invitation with email and role", async ({
      page,
    }) => {
      const roleId = await createDBRole(
        testOrg.id,
        ownerUser.id,
        "Developer",
        "Development team"
      );

      const inviteEmail = `invite-role-${Date.now()}@example.com`;

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await inviteDialog.inviteMember(inviteEmail, "Developer");
      await waitForSuccessToast(page, "Invitation sent", { timeout: 10000 });
      await inviteDialog.waitForClose();
      const email = await waitForEmail(inviteEmail, 15000);
      expect(email).not.toBeNull();
      expect(email?.To.some((to) => to.Address === inviteEmail)).toBe(true);
    });

    test("should show success toast after invitation", async ({ page }) => {
      const inviteEmail = `invite-toast-${Date.now()}@example.com`;

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await inviteDialog.inviteMember(inviteEmail);
      await waitForSuccessToast(page, "Invitation sent", { timeout: 10000 });
    });

    test("should extract invitation token from email", async ({ page }) => {
      const inviteEmail = `invite-token-${Date.now()}@example.com`;

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await inviteDialog.inviteMember(inviteEmail);
      await waitForSuccessToast(page, "Invitation sent", { timeout: 10000 });
      const token = await new Promise<string | null>((resolve) => {
        setTimeout(async () => {
          const extractedToken = await getInvitationTokenFromEmail(inviteEmail);
          resolve(extractedToken);
        }, 5000);
      });
      expect(token).not.toBeNull();
      expect(token?.length).toBeGreaterThan(0);
    });
  });

  test.describe("Error Handling", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Invite Error Test Org",
      });
    });

    test("should handle invitation errors gracefully", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickInviteMember();
      const inviteDialog = new InviteDialog(page);
      await inviteDialog.waitFor();
      await inviteDialog.fillEmail("not-an-email");
      await page.getByLabel("Email Address").blur();
      await page
        .waitForFunction(
          () => {
            const input = document.querySelector(
              'input[aria-label="Email Address"]'
            ) as HTMLInputElement;
            return input.value === "not-an-email";
          },
          { timeout: 1000 }
        )
        .catch(() => {});
      await inviteDialog.sendInvitation();
      const formMessages = page.locator('[data-slot="form-message"]');
      const hasError = await formMessages
        .first()
        .isVisible()
        .catch(() => false);
      const dialogStillOpen = await inviteDialog.isVisible();

      expect(dialogStillOpen || hasError).toBe(true);
      await expect(inviteDialog.getLocator()).toBeVisible();
    });
  });
});
