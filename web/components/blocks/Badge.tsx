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

export interface BadgeProps {
  title: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'neutral' | 'info' | 'success' | 'warning' | 'danger';
  className?: string;
}

export function Badge({ title, size = 'md', variant = 'neutral', className }: BadgeProps) {
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
    </span>
  );
}
