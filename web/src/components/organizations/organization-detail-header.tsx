export function OrganizationDetailHeader({
  title,
  description = "View organization information.",
}: {
  title: string;
  description?: string;
}) {
  return (
    <div className="mb-6">
      <h1 className="text-2xl font-bold">{title}</h1>
      <p className="mt-2 text-gray-600">{description}</p>
    </div>
  );
}
