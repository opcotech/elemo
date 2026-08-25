import { InfoIcon } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export function DemoLoginAlert() {
  return (
    <Alert variant="warning" className="mb-4">
      <InfoIcon />
      <AlertTitle>Demo Instance</AlertTitle>
      <AlertDescription>
        <p>
          This instance allows you to explore Elemo without installing it
          yourself. All data is reset every 6 hours.
        </p>
        <p>
          For login details please visit the{" "}
          <a
            href="https://github.com/opcotech/elemo/tree/main/tools/workload-prefill#logins"
            target="_blank"
            rel="noopener noreferrer nofollow"
          >
            demo documentation
          </a>
          .
        </p>
      </AlertDescription>
    </Alert>
  );
}
