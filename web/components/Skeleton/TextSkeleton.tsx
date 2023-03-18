import { concat } from '@/helpers';

export interface TextSkeletonProps {
  className?: string;
}

export default function TextSkeleton({ className }: TextSkeletonProps) {
  return (
    <div role="status" className={concat(className, 'w-full animate-pulse')}>
      <div className="h-2.5 bg-gray-200 rounded-full  w-48 mb-4"></div>
      <div className="h-2 bg-gray-200 rounded-full  max-w-[360px] mb-2.5"></div>
      <div className="h-2 bg-gray-200 rounded-full  mb-2.5"></div>
      <div className="h-2 bg-gray-200 rounded-full  max-w-[330px] mb-2.5"></div>
      <div className="h-2 bg-gray-200 rounded-full  max-w-[300px] mb-2.5"></div>
      <div className="h-2 bg-gray-200 rounded-full  max-w-[360px]"></div>
      <span className="sr-only">Loading...</span>
    </div>
  );
}
