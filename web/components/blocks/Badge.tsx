import { concat } from '@/lib/helpers';

const SIZES = {
  sm: 'px-1 py-0.5 text-xs',
  md: 'px-2 py-1 text-xs',
  lg: 'px-2.5 py-1.5 text-sm'
};

const VARIANTS = {
  neutral: 'bg-gray-50 text-gray-600 ring-gray-500/10',
  info: 'bg-blue-50 text-blue-700 ring-blue-700/10',
  success: 'bg-green-50 text-green-700 ring-green-600/20',
  warning: 'bg-yellow-50 text-yellow-800 ring-yellow-600/10',
  danger: 'bg-red-50 text-red-700 ring-red-600/10'
};

const DISMISS_BUTTON_VARIANTS = {
  neutral: 'hover:bg-gray-500/20',
  info: 'hover:bg-blue-700/20',
  success: 'hover:bg-green-600/20',
  warning: 'hover:bg-yellow-600/20',
  danger: 'hover:bg-red-600/20'
};

export interface BadgeProps {
  title: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'neutral' | 'info' | 'success' | 'warning' | 'danger';
  dismissible?: boolean;
  onDismiss?: () => void;
  className?: string;
}

export function Badge({ title, size = 'md', variant = 'neutral', dismissible, onDismiss, className }: BadgeProps) {
  return (
    <span
      className={concat(
        'inline-flex items-center rounded-md font-medium ring-1 ring-inset',
        className,
        SIZES[size],
        VARIANTS[variant]
      )}
    >
      {title}

      {dismissible && (
        <button
          type="button"
          className={concat('group relative -mr-0.5 ml-1 h-3.5 w-3.5 rounded-sm', DISMISS_BUTTON_VARIANTS[variant])}
          onClick={onDismiss}
        >
          <span className="sr-only">Remove</span>
          <svg viewBox="0 0 14 14" className="h-3.5 w-3.5 stroke-gray-600/50 group-hover:stroke-gray-600/75">
            <path d="M4 4l6 6m0-6l-6 6" />
          </svg>
          <span className="absolute -inset-1" />
        </button>
      )}
    </span>
  );
}
