import { defineElemoPlugin } from "@elemo/plugin-sdk";

import { AccountsPage } from "./accounts";
import { BudgetsPage } from "./budgets";
import { BudgetPicker } from "./charge";
import { OrganizationStatus } from "./org-settings";
import { ReportPage } from "./report";

export default defineElemoPlugin({
  id: "com.elemo.accounting",
  activate(elemo) {
    const orgSettings = elemo.slots.register(
      "organization.settings",
      (props) => <OrganizationStatus {...props} elemo={elemo} />,
    );
    const projectSettings = elemo.slots.register(
      "project.settings",
      (props) => (
        <BudgetPicker
          {...props}
          elemo={elemo}
          parentId={String(props.projectId ?? "")}
          parentType="Project"
          title="Project budget"
          testId="accounting-project-settings"
        />
      ),
    );
    const projectSidebar = elemo.slots.register("project.sidebar", (props) => (
      <BudgetPicker
        {...props}
        elemo={elemo}
        parentId={String(props.projectId ?? "")}
        parentType="Project"
        title="Budget"
        testId="accounting-project-sidebar"
      />
    ));
    const issueSidebar = elemo.slots.register("issue.sidebar", (props) => (
      <BudgetPicker
        {...props}
        elemo={elemo}
        parentId={String(props.issueId ?? "")}
        parentType="Issue"
        title="Work item budget"
        testId="accounting-issue-sidebar"
      />
    ));
    const accounts = elemo.routes.register("accounts", (props) => (
      <AccountsPage {...props} elemo={elemo} />
    ));
    const budgets = elemo.routes.register("budgets", (props) => (
      <BudgetsPage {...props} elemo={elemo} />
    ));
    const report = elemo.routes.register("report", (props) => (
      <ReportPage {...props} elemo={elemo} />
    ));
    return () => {
      orgSettings();
      projectSettings();
      projectSidebar();
      issueSidebar();
      accounts();
      budgets();
      report();
    };
  },
});
