import {concat} from '@/helpers';

export interface ListSkeletonProps {
  className?: string;
  withBorder?: boolean;
  fullWidth?: boolean;
  count: number;
}

export default function ListSkeleton({ className, withBorder, fullWidth, count }: ListSkeletonProps) {
  return (
    <div
      role="status"
      className={concat(
        className,
        'space-y-4 divide-y divide-gray-200 animate-pulse',
        withBorder ? 'rounded border border-gray-200 shadow' : '',
        fullWidth ? 'w-full' : 'max-w-md'
      )}
    >
      {Array.from({ length: count }, (_, i) => (
        <div key={i} className={`${i > 0 && 'pt-4'} flex justify-between items-center`}>
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
