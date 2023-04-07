import type {ReactNode} from 'react';

export default function ContentSkeleton({ children }: { children?: ReactNode }) {
  return (
    <div className="mx-auto max-w-7xl">
      <div className="h-96 rounded-lg border-2 border-dashed border-gray-200">{children}</div>
    </div>
  );
}
