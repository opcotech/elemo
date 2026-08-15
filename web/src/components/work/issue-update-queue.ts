const pending = new Map<string, Promise<unknown>>();

/** Run issue mutations one at a time so later snapshots see earlier results. */
export function enqueueIssueUpdate<T>(
  issueId: string,
  task: () => Promise<T>
): Promise<T> {
  const previous = pending.get(issueId) ?? Promise.resolve();
  const next = previous.then(() => task());
  const tracked = next.then(
    () => undefined,
    () => undefined
  );
  pending.set(issueId, tracked);
  void tracked.then(() => {
    if (pending.get(issueId) === tracked) {
      pending.delete(issueId);
    }
  });
  return next;
}
