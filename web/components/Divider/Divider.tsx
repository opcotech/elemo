import {concat} from '@/helpers';

export default function Divider({className}: { className?: string }) {
  return (
    <div className={concat(className, 'relative')}>
      <div className="absolute inset-0 flex items-center" aria-hidden="true">
        <div className="w-full border-t border-gray-300"/>
      </div>
    </div>
  );
}
