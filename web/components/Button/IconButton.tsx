import type {ReactNode} from 'react';

import Icon from '@/components/Icon';
import {concat} from '@/helpers';
import type {IconVariant} from '@/types/heroicon';

export interface IconButtonProps {
  icon: IconVariant;
  disabled?: boolean;
  onClick?: () => void;
  className?: string;
  size?: number;
  children?: ReactNode;
}

export default function IconButton({icon, disabled, onClick, className, size, children}: IconButtonProps) {
  const sizeClass = size ? `h-${size} w-${size}` : 'h-5 w-5';

  return (
    <button className={className} disabled={disabled} onClick={onClick}>
      <div className={'relative'}>
        <Icon variant={icon} className={concat(sizeClass, 'cursor-pointer icon-normal')} aria-hidden="true"/>
        {children}
      </div>
    </button>
  );
}
