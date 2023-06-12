import Image from 'next/image';

import { concat } from '@/lib/helpers';

const SIZES = {
  xs: 'w-6 h-6',
  sm: 'w-8 h-8',
  md: 'w-12 h-12',
  lg: 'w-16 h-16',
  xl: 'w-20 h-20'
};

const FONT_SIZES = {
  xs: 'text-xs',
  sm: 'text-xs',
  md: 'text-lg',
  lg: 'text-xl',
  xl: 'text-2xl'
};

export interface AvatarProps extends AvatarImageProps, AvatarInitialsProps {
  size: keyof typeof SIZES;
}

export function Avatar({ size = 'md', initials = 'N/A', alt = 'Avatar', src, grayscale, className }: AvatarProps) {
  const avatarClasses = concat(className, SIZES[size]);

  if (src !== undefined && src !== null && src !== '') {
    return <AvatarImage className={avatarClasses} src={src} alt={alt} grayscale={grayscale} />;
  }

  return <AvatarInitials className={avatarClasses} textClassName={FONT_SIZES[size]} initials={initials} />;
}

export interface AvatarImageProps {
  src: string;
  className?: string;
  alt?: string;
  grayscale?: boolean;
}

function AvatarImage({ src, alt, grayscale, className }: AvatarImageProps) {
  const Component = !src.startsWith('/') ? 'img' : Image;
  return (
    <Component
      className={concat(className, 'rounded-full', grayscale ? 'grayscale' : undefined)}
      src={src}
      width={100}
      height={100}
      alt={alt || 'Avatar'}
      {...(Component !== 'img' && { priority: true })}
    />
  );
}

export interface AvatarInitialsProps {
  initials: string;
  textClassName?: string;
  className?: string;
}

function AvatarInitials({ initials, textClassName, className }: AvatarInitialsProps) {
  return (
    <span className={concat(className, 'inline-flex items-center justify-center rounded-full bg-gray-800')}>
      <span className={`leading-none text-white ${textClassName}`}>{initials}</span>
    </span>
  );
}
