import * as HeroIcons from '@heroicons/react/24/outline';
import type { HTMLAttributes } from 'react';
import { createElement } from 'react';

import type { IconVariant } from '@/types/heroicon';

export interface IconProps extends HTMLAttributes<HTMLSpanElement> {
  variant: IconVariant;
}

export default function Icon({ variant, className }: IconProps) {
  // eslint-disable-next-line import/namespace
  const heroIcon = HeroIcons[variant];

  if (typeof heroIcon === 'undefined') {
    return <></>;
  }

  return <span className="icon">{createElement(heroIcon, { className })}</span>;
}
