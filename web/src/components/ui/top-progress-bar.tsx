"use client";

import { useEffect, useState } from "react";
import { useRouter } from "@tanstack/react-router";
import { Progress } from "@/components/ui/progress";

export function TopProgressBar() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const unsubStart = router.subscribe("onBeforeNavigate", () => {
      setLoading(true);
    });

    const unsubDone = router.subscribe("onResolved", () => {
      setLoading(false);
    });

    return () => {
      unsubStart();
      unsubDone();
    };
  }, [router]);

  if (!loading) return null;

  return (
    <div className="fixed top-0 left-0 w-full z-50">
      <Progress value={100} className="h-1 bg-primary" />
    </div>
  );
}
