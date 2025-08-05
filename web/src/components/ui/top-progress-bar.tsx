import { useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";

import { Progress } from "@/components/ui/progress";

export function TopProgressBar() {
  const [progress, setProgress] = useState(0);
  const [isVisible, setIsVisible] = useState(false);
  const router = useRouter();

  useEffect(() => {
    let timer: NodeJS.Timeout | undefined;

    const start = () => {
      setIsVisible(true);
      setProgress(10);
      timer = setInterval(() => {
        setProgress((prev) => {
          if (prev >= 90) {
            clearInterval(timer);
            return prev;
          }
          if (prev < 20) return prev + 10;
          if (prev < 50) return prev + 4;
          return prev + 2;
        });
      }, 300);
    };

    const complete = () => {
      if (timer) {
        clearInterval(timer);
      }
      setProgress(100);
      setTimeout(() => {
        setIsVisible(false);
        setTimeout(() => setProgress(0), 500);
      }, 500);
    };

    const unsubBefore = router.subscribe("onBeforeNavigate", start);
    const unsubResolved = router.subscribe("onResolved", complete);

    return () => {
      unsubBefore();
      unsubResolved();
      if (timer) {
        clearInterval(timer);
      }
    };
  }, [router]);

  if (!isVisible) {
    return null;
  }

  return (
    <div className="fixed top-0 left-0 z-50 w-full">
      <Progress value={progress} className="h-1" />
    </div>
  );
}
