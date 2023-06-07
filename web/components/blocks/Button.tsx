import { HTMLAttributes } from 'react';
import { Spinner } from './Spinner';
import { IconVariant } from '@/types/heroicon';
import { concat } from '@/lib/helpers';
import { Icon } from '@/components/blocks/Icon';

const SIZES = {
  xs: 'text-xs px-2 py-1.5',
  sm: 'text-sm px-3 py-2',
  md: 'text-base px-4 py-2',
  lg: 'text-base px-5 py-3'
};

const INDICATOR_SIZES = {
  xs: 'w-4 h-4',
  sm: 'w-5 h-5',
  md: 'w-6 h-6',
  lg: 'w-6 h-6'
};

const VARIANTS = {
  primary:
    'bg-blue-500 text-white hover:bg-blue-600 disabled:bg-blue-500 focus:ring-blue-500 focus:ring-2 focus:ring-offset-2',
  secondary:
    'bg-gray-500 text-white hover:bg-gray-600 disabled:bg-gray-500 focus:ring-gray-500 focus:ring-2 focus:ring-offset-2',
  accent:
    'bg-green-500 text-white hover:bg-green-600 disabled:bg-green-500 focus:ring-green-500 focus:ring-2 focus:ring-offset-2',
  ghost:
    'bg-transparent text-gray-500 hover:bg-gray-100 disabled:bg-transparent focus:ring-gray-500 shadow-none focus:ring-2 focus:ring-offset-2',
  link: 'text-blue-500 hover:text-blue-600 disabled:text-blue-500 shadow-none rounded-none focus:ring-0'
};

interface ButtonProps extends HTMLAttributes<HTMLButtonElement> {
  type?: 'button' | 'submit' | 'reset';
  size?: keyof typeof SIZES;
  variant?: keyof typeof VARIANTS;
  icon?: IconVariant;
  loading?: boolean;
  disabled?: boolean;
  children?: React.ReactNode;
  className?: string;
}

export const Button = ({
  type = 'button',
  size = 'md',
  variant = 'primary',
  icon = undefined,
  loading = false,
  disabled = false,
  children,
  className,
  ...props
}: ButtonProps) => {
  const interactive = !disabled && !loading;

  return (
    <button
      {...props}
      type={type}
      disabled={!interactive}
      className={concat(
        className,
        icon
          ? 'rounded-full p-0.5 text-gray-600 hover:text-black focus:ring-gray-600'
          : `${VARIANTS[variant]} ${SIZES[size]} rounded-md shadow-sm`,
        'inline-flex items-center justify-center disabled:opacity-75 disabled:cursor-not-allowed focus:outline-none'
      )}
    >
      {loading ? (
        <Spinner className={INDICATOR_SIZES[size]} />
      ) : icon ? (
        <Icon size={size} variant={icon}>
          <span className="sr-only">{props['aria-label']}</span>
          {children}
        </Icon>
      ) : (
        children
      )}
    </button>
  );
};
