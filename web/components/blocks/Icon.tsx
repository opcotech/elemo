import * as HeroIcons from '@heroicons/react/24/outline';
import type { HTMLAttributes } from 'react';
import { createElement } from 'react';

import type { IconVariant } from '@/types/heroicon';
import { concat } from '@/lib/helpers';

const SIZES = {
  xs: 'w-4 h-4',
  sm: 'w-5 h-5',
  md: 'w-6 h-6',
  lg: 'w-8 h-8'
};

export interface IconProps extends HTMLAttributes<HTMLSpanElement> {
  size?: keyof typeof SIZES;
  variant: IconVariant;
  children?: React.ReactNode;
}

export function Icon({ size = 'md', variant, children, className }: IconProps) {
  // eslint-disable-next-line import/namespace
  const heroIcon = HeroIcons[variant];

  if (typeof heroIcon === 'undefined') {
    return <></>;
  }

  return (
    <span className="icon" aria-hidden="true">
      {createElement(heroIcon, { className: concat(className, SIZES[size]) })}
      {children}
    </span>
  );
}
