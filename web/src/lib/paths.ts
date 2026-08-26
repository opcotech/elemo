export interface OrganizationPathInput {
  readonly organizationSlug: string;
}

export interface NamespacePathInput extends OrganizationPathInput {
  readonly namespaceSlug: string;
}

export interface ProjectPathInput extends NamespacePathInput {
  readonly projectKey: string;
}

export interface WorkItemPathInput extends NamespacePathInput {
  readonly issueKey: string;
}

export function organizationPath({
  organizationSlug,
}: OrganizationPathInput): string {
  return `/organizations/${organizationSlug}`;
}

export function organizationDocumentsPath(
  input: OrganizationPathInput
): string {
  return `${organizationPath(input)}/documents`;
}

export function namespacePath({
  organizationSlug,
  namespaceSlug,
}: NamespacePathInput): string {
  return `/organizations/${organizationSlug}/namespaces/${namespaceSlug}`;
}

export function namespaceProjectsPath(input: NamespacePathInput): string {
  return `${namespacePath(input)}/projects`;
}

export function namespaceWorkPath(input: NamespacePathInput): string {
  return `${namespacePath(input)}/work`;
}

export function namespaceDocumentsPath(input: NamespacePathInput): string {
  return `${namespacePath(input)}/documents`;
}

export function namespaceAdministrationPath(input: NamespacePathInput): string {
  return `${namespacePath(input)}/administration`;
}

export function projectPath({
  organizationSlug,
  namespaceSlug,
  projectKey,
}: ProjectPathInput): string {
  return `${namespacePath({ organizationSlug, namespaceSlug })}/projects/${projectKey}`;
}

export function projectWorkPath(input: ProjectPathInput): string {
  return `${projectPath(input)}/work`;
}

export function projectDocumentsPath(input: ProjectPathInput): string {
  return `${projectPath(input)}/documents`;
}

export function projectActivityPath(input: ProjectPathInput): string {
  return `${projectPath(input)}/activity`;
}

export function workItemPath({
  organizationSlug,
  namespaceSlug,
  issueKey,
}: WorkItemPathInput): string {
  return `/work/${organizationSlug}/${namespaceSlug}/${issueKey}`;
}

export function documentPath(documentId: string): string {
  return `/documents/${documentId}`;
}

export function settingsOrganizationPath({
  organizationSlug,
}: OrganizationPathInput): string {
  return `/settings/organizations/${organizationSlug}`;
}

export function settingsOrganizationEditPath(
  input: OrganizationPathInput
): string {
  return `${settingsOrganizationPath(input)}/edit`;
}

export function settingsNamespaceNewPath(input: OrganizationPathInput): string {
  return `${settingsOrganizationPath(input)}/namespaces/new`;
}

export function settingsNamespacePath(input: NamespacePathInput): string {
  return `${settingsOrganizationPath(input)}/namespaces/${input.namespaceSlug}`;
}

export function settingsNamespaceEditPath(input: NamespacePathInput): string {
  return `${settingsNamespacePath(input)}/edit`;
}

export function settingsProjectNewPath(input: NamespacePathInput): string {
  return `${settingsNamespacePath(input)}/projects/new`;
}

export function settingsProjectPath(input: ProjectPathInput): string {
  return `${settingsNamespacePath(input)}/projects/${input.projectKey}`;
}

export function settingsProjectEditPath(input: ProjectPathInput): string {
  return `${settingsProjectPath(input)}/edit`;
}

export function organizationRouteParams({
  organizationSlug,
}: OrganizationPathInput) {
  return { organizationSlug };
}

export function namespaceRouteParams({
  organizationSlug,
  namespaceSlug,
}: NamespacePathInput) {
  return { organizationSlug, namespaceSlug };
}

export function projectRouteParams({
  organizationSlug,
  namespaceSlug,
  projectKey,
}: ProjectPathInput) {
  return { organizationSlug, namespaceSlug, projectKey };
}

export function workItemRouteParams({
  organizationSlug,
  namespaceSlug,
  issueKey,
}: WorkItemPathInput) {
  return { organizationSlug, namespaceSlug, issueKey };
}
