import { Icon } from '@/components/blocks/Icon';
import { concat } from '@/lib/helpers';

export interface VideoSkeletonProps {
  className?: string;
}

export function VideoSkeleton({ className }: VideoSkeletonProps) {
  return (
    <div
      role="status"
      className={concat(
        className,
        'flex justify-center items-center p-4 bg-gray-200 rounded-lg animate-pulse motion-reduce:animate-none'
      )}
    >
      <Icon variant="VideoCameraIcon" size="lg" className="text-gray-100" />
      <span className="sr-only">Loading...</span>
    </div>
  );
}
