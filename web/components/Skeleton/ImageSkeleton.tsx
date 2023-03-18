import Icon from '@/components/Icon';
import { concat } from '@/helpers';

export interface ImageSkeletonProps {
  className?: string;
}

export default function ImageSkeleton({ className }: ImageSkeletonProps) {
  return (
    <div role="status" className={concat(className, 'flex justify-center items-center bg-gray-300 animate-pulse')}>
      <Icon variant="PhotoIcon" className="w-12 h-12 text-gray-200" />
      <span className="sr-only">Loading...</span>
    </div>
  );
}
