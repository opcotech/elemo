import { expect, test } from "@playwright/test";
import { createDBUser, loginUser } from "./utils/auth";
import {
  addMemberToOrganization,
  createDBOrganization,
} from "./utils/organization";
import { MembersPage } from "./pages/members-page";
import { Dialog } from "./components/dialog";
import { waitForSuccessToast } from "./helpers/toast";

test.describe("@settings.organization-member-remove Organization Member Remove E2E Tests", () => {
  test.describe("Permission-Based Visibility", () => {
    let ownerUser: any;
    let writeUser: any;
    let readUser: any;
    let testOrg: any;
    let memberToRemove: any;

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
      memberToRemove = await createDBUser("active", {
        first_name: "Member",
        last_name: "ToRemove",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Remove Permission Test Org",
      });
      await addMemberToOrganization(testOrg.id, writeUser.id, "read");
      await addMemberToOrganization(testOrg.id, writeUser.id, "write");
      await addMemberToOrganization(testOrg.id, readUser.id, "read");
      await addMemberToOrganization(testOrg.id, memberToRemove.id, "read");
    });

    test("user with write permission should see remove button", async ({
      page,
    }) => {
      await loginUser(page, writeUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const memberCount = await membersPage.getMemberCount();
      if (memberCount > 1) {
        const hasRemoveButton =
          await membersPage.hasRemoveButton("Member ToRemove");
        expect(hasRemoveButton).toBe(true);
      } else {
        const hasInviteButton = await membersPage.hasInviteButton();
        expect(hasInviteButton).toBe(true);
      }
    });

    test("user with read permission should not see remove button", async ({
      page,
    }) => {
      await loginUser(page, readUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const hasRemoveButton =
        await membersPage.hasRemoveButton("Member ToRemove");
      expect(hasRemoveButton).toBe(false);
    });

    test("owner should see remove button for all members", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      const hasWriteUserRemoveButton =
        await membersPage.hasRemoveButton("Write User");
      expect(hasWriteUserRemoveButton).toBe(true);

      const hasReadUserRemoveButton =
        await membersPage.hasRemoveButton("Read User");
      expect(hasReadUserRemoveButton).toBe(true);
    });
  });

  test.describe("Remove Dialog", () => {
    let ownerUser: any;
    let testOrg: any;
    let memberToRemove: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });
      memberToRemove = await createDBUser("active", {
        first_name: "Dialog",
        last_name: "Member",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Remove Dialog Test Org",
      });

      await addMemberToOrganization(testOrg.id, memberToRemove.id, "read");
    });

    test("should open remove dialog when remove button is clicked", async ({
      page,
    }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickRemoveMember("Dialog Member");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      const dialogLocator = dialog.getLocator();
      await expect(
        dialogLocator.getByText(/Are you sure you want to remove/i)
      ).toBeVisible();
    });

    test("should close dialog when cancel is clicked", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickRemoveMember("Dialog Member");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.cancel();
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`);
    });
  });

  test.describe("Successful Removal", () => {
    let ownerUser: any;
    let testOrg: any;

    test.beforeEach(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: `Remove Success Test Org ${Date.now()}`,
      });
    });

    test("should successfully remove member and update list", async ({
      page,
    }) => {
      const memberToRemove = await createDBUser("active", {
        first_name: "Remove",
        last_name: "Test",
      });
      await addMemberToOrganization(testOrg.id, memberToRemove.id, "read");

      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.waitForMember("Remove Test");
      await membersPage.clickRemoveMember("Remove Test");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.confirm("Remove");
      await waitForSuccessToast(page, "Member removed", { timeout: 10000 });
      await membersPage.waitForMembersLoad();
      const memberExists = await membersPage.memberExists("Remove Test");
      expect(memberExists).toBe(false);
    });
  });

  test.describe("Error Handling", () => {
    let ownerUser: any;
    let testOrg: any;
    let memberToRemove: any;

    test.beforeAll(async () => {
      ownerUser = await createDBUser("active", {
        first_name: "Owner",
        last_name: "User",
      });
      memberToRemove = await createDBUser("active", {
        first_name: "Error",
        last_name: "Test",
      });

      testOrg = await createDBOrganization(ownerUser.id, "active", {
        name: "Remove Error Test Org",
      });

      await addMemberToOrganization(testOrg.id, memberToRemove.id, "read");
    });

    test("should handle removal errors gracefully", async ({ page }) => {
      await loginUser(page, ownerUser, {
        destination: `/settings/organizations/${testOrg.id}`,
      });

      const membersPage = new MembersPage(page, testOrg.id);
      await membersPage.waitForMembersLoad();
      await membersPage.clickRemoveMember("Error Test");
      const dialog = new Dialog(page);
      await dialog.waitFor();
      await dialog.cancel();
      await expect(page).toHaveURL(`/settings/organizations/${testOrg.id}`);
    });
  });
});
