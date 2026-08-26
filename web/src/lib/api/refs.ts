export function organizationRefPath(organizationRef: string) {
  return { organizationRef };
}

export function namespaceRefPath(
  organizationRef: string,
  namespaceRef: string
) {
  return { organizationRef, namespaceRef };
}

export function projectIdPath(projectId: string) {
  return { projectId };
}
