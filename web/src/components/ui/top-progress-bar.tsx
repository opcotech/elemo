import { useRouter as useAppRouter } from "@tanstack/react-router";
import * as React from "react";

const INITIAL_POSITION = 12;
const MAX_WAITING_POSITION = 92;
const ADVANCE_INTERVAL_MS = 240;
const COMPLETION_PAUSE_MS = 180;

function nextPosition(position: number) {
  const remaining = MAX_WAITING_POSITION - position;
  return Math.min(
    MAX_WAITING_POSITION,
    position + Math.max(1, remaining * 0.08)
  );
}

export function NavigationProgress() {
  const navigationRouter = useAppRouter();
  const intervalRef = React.useRef<ReturnType<typeof setInterval> | undefined>(
    undefined
  );
  const hideTimeoutRef = React.useRef<
    ReturnType<typeof setTimeout> | undefined
  >(undefined);
  const [position, setPosition] = React.useState<number | null>(null);

  React.useEffect(() => {
    const stopInterval = () => {
      if (intervalRef.current !== undefined) {
        clearInterval(intervalRef.current);
        intervalRef.current = undefined;
      }
    };
    const stopHideTimeout = () => {
      if (hideTimeoutRef.current !== undefined) {
        clearTimeout(hideTimeoutRef.current);
        hideTimeoutRef.current = undefined;
      }
    };
    const beginNavigation = () => {
      stopInterval();
      stopHideTimeout();
      setPosition(INITIAL_POSITION);
      intervalRef.current = setInterval(() => {
        setPosition((current) =>
          current === null ? INITIAL_POSITION : nextPosition(current)
        );
      }, ADVANCE_INTERVAL_MS);
    };
    const finishNavigation = () => {
      stopInterval();
      setPosition(100);
      hideTimeoutRef.current = setTimeout(() => {
        setPosition(null);
        hideTimeoutRef.current = undefined;
      }, COMPLETION_PAUSE_MS);
    };

    const removeBeforeListener = navigationRouter.subscribe(
      "onBeforeNavigate",
      beginNavigation
    );
    const removeResolvedListener = navigationRouter.subscribe(
      "onResolved",
      finishNavigation
    );

    const disconnect = () => {
      removeBeforeListener();
      removeResolvedListener();
      stopInterval();
      stopHideTimeout();
    };
    return disconnect;
  }, [navigationRouter]);

  if (position === null) return null;

  const indicator = (
    <div
      aria-label="Page navigation"
      aria-valuemax={100}
      aria-valuemin={0}
      aria-valuenow={Math.round(position)}
      className="bg-primary/20 pointer-events-none fixed inset-x-0 top-0 z-50 h-1 overflow-hidden"
      role="progressbar"
    >
      <div
        className="bg-primary h-full transition-[width] duration-200 ease-out"
        style={{ width: `${position}%` }}
      />
    </div>
  );

  return indicator;
}
