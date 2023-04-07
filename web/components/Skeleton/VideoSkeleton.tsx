import Icon from '@/components/Icon';

export default function VideoSkeleton() {
  return (
    <div role="status" className="flex justify-center items-center max-w-sm h-56 bg-gray-300 rounded-lg animate-pulse">
      <Icon variant="VideoCameraIcon" className="w-12 h-12 text-gray-200"/>
      <span className="sr-only">Loading...</span>
    </div>
  );
}
