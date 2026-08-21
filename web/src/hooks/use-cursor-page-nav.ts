import { useCallback, useState } from "react";

export function useCursorPageNav({
  resetKey = "",
}: { resetKey?: string } = {}) {
  const [pageToken, setPageToken] = useState<string | undefined>();
  const [previousTokens, setPreviousTokens] = useState<string[]>([]);
  const [previousResetKey, setPreviousResetKey] = useState(resetKey);

  if (resetKey !== previousResetKey) {
    setPreviousResetKey(resetKey);
    setPageToken(undefined);
    setPreviousTokens([]);
  }

  const canGoPrevious = previousTokens.length > 0 || Boolean(pageToken);

  const goNext = useCallback(
    (nextPageToken: string) => {
      setPreviousTokens((stack) => [...stack, pageToken ?? ""]);
      setPageToken(nextPageToken);
    },
    [pageToken]
  );

  const goPrevious = useCallback(() => {
    const previous = previousTokens.at(-1);
    setPreviousTokens((stack) => stack.slice(0, -1));
    setPageToken(previous || undefined);
  }, [previousTokens]);

  return {
    pageToken,
    canGoPrevious,
    goNext,
    goPrevious,
  };
}
