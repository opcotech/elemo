import { defineElemoPlugin } from "@elemo/plugin-sdk";

import { LoggedTime } from "./activity";
import { TimeReport } from "./report";
import { TimeTracker } from "./sidebar";

export default defineElemoPlugin({
  id: "com.elemo.timetracking",
  activate(elemo) {
    const sidebar = elemo.slots.register("issue.sidebar", (props) => (
      <TimeTracker {...props} elemo={elemo} />
    ));
    const activity = elemo.slots.register(
      "issue.activity",
      (props) => <LoggedTime {...props} elemo={elemo} />,
      { title: "Logged time" }
    );
    const report = elemo.routes.register("report", (props) => (
      <TimeReport {...props} elemo={elemo} />
    ));
    return () => {
      sidebar();
      activity();
      report();
    };
  },
});
