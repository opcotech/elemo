import { concat } from '@/lib/helpers';

export interface ListSkeletonProps {
  className?: string;
  withBorder?: boolean;
  fullWidth?: boolean;
  count: number;
}

export function ListSkeleton({ className, withBorder, fullWidth, count = 3 }: ListSkeletonProps) {
  return (
    <div
      role="status"
      className={concat(
        className,
        'space-y-4 divide-y divide-gray-200',
        withBorder ? 'rounded border border-gray-200 py-6 px-4 shadow' : '',
        fullWidth ? 'w-full' : 'max-w-md'
      )}
    >
      {Array.from({ length: count }, (_, i) => (
        <div
          key={i}
          className={`${i > 0 && 'pt-4'} flex justify-between items-center animate-pulse motion-reduce:animate-none`}
        >
          <div>
            <div className="h-2.5 bg-gray-300 rounded-full  w-24 mb-2.5"></div>
            <div className="w-32 h-2 bg-gray-200 rounded-full "></div>
          </div>
          <div className="h-2.5 bg-gray-300 rounded-full  w-12"></div>
        </div>
      ))}

      <span className="sr-only">Loading...</span>
    </div>
  );
}
