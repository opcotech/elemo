import type {ButtonHTMLAttributes} from 'react';
import Spinner from '@/components/Spinner';
import {concat} from '@/helpers';

const fontSizes = {
  xs: 'text-xs h-7',
  sm: 'text-sm h-8',
  md: 'text-base h-10',
  lg: 'text-lg h-12',
  xl: 'text-xl h-14'
};

const variantClasses = {
  primary: 'bg-blue-500 text-white hover:bg-blue-600 focus:ring-blue-500',
  secondary: 'text-black bg-gray-200 hover:bg-gray-300 focus:ring-gray-300',
  accent: 'bg-gray-800 text-white hover:bg-gray-900 focus:ring-gray-500',
  danger: 'bg-red-500 text-white hover:bg-red-600 focus:ring-red-500'
};

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  loading?: boolean;
  variant?: 'primary' | 'secondary' | 'danger' | 'accent';
}

export default function Button({ className, disabled, loading, variant, size, children, ...props }: ButtonProps) {
  return (
    <button
      disabled={disabled || loading}
      className={concat(
        className,
        variantClasses[variant ?? 'primary'],
        fontSizes[size ?? 'md'],
        'inline-flex items-center justify-center rounded-md px-4 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-75'
      )}
      {...props}
    >
      {loading ? <Spinner className="w-4 h-4 text-white" /> : children}
    </button>
  );
}
