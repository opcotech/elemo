import { useCallback, useState } from "react";

interface UsePendingPermissionsProps {
  initialActions?: string[];
}

export function usePendingPermissions({
  initialActions = [],
}: UsePendingPermissionsProps = {}) {
  const [pendingActions, setPendingActions] =
    useState<string[]>(initialActions);

  const addAction = useCallback((action: string) => {
    setPendingActions((prev) =>
      prev.includes(action) ? prev : [...prev, action]
    );
  }, []);

  const removeAction = useCallback((action: string) => {
    setPendingActions((prev) => prev.filter((item) => item !== action));
  }, []);

  const toggleAction = useCallback((action: string) => {
    setPendingActions((prev) =>
      prev.includes(action)
        ? prev.filter((item) => item !== action)
        : [...prev, action]
    );
  }, []);

  const clearActions = useCallback(() => {
    setPendingActions([]);
  }, []);

  return {
    pendingActions,
    addAction,
    removeAction,
    toggleAction,
    clearActions,
    hasPendingActions: pendingActions.length > 0,
  };
}
