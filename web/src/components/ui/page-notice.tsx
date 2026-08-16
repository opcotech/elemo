import { Alert, AlertDescription } from "@/components/ui/alert";
import { PageHeader } from "@/components/ui/page-header";

export function PageNotice({
  title,
  message,
}: {
  title: string;
  message: string;
}) {
  return (
    <div className="space-y-6">
      <PageHeader title={title} />
      <Alert variant="destructive">
        <AlertDescription>{message}</AlertDescription>
      </Alert>
    </div>
  );
}
