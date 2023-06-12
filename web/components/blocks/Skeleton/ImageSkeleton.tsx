import { Icon } from '@/components/blocks/Icon';
import { concat } from '@/lib/helpers';

export interface ImageSkeletonProps {
  className?: string;
}

export function ImageSkeleton({ className }: ImageSkeletonProps) {
  return (
    <div
      role="status"
      className={concat(
        className,
        'flex justify-center items-center p-4 bg-gray-200 rounded-lg animate-pulse motion-reduce:animate-none'
      )}
    >
      <Icon variant="PhotoIcon" size="lg" className="text-gray-100" />
      <span className="sr-only">Loading...</span>
    </div>
  );
}
