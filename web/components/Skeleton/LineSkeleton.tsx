import { concat } from '@/helpers';

export interface LineSkeletonProps {
  className?: string;
}

export default function LineSkeleton({ className }: LineSkeletonProps) {
  return (
    <div role="status" className={concat(className, 'w-full animate-pulse')}>
      <div className="h-2.5 bg-gray-200 rounded-full"></div>
      <span className="sr-only">Loading...</span>
    </div>
  );
}
