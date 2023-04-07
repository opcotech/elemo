import Image from 'next/image';

import {concat} from '@/helpers';

const sizes = {
  xs: 'w-6 h-6',
  sm: 'w-8 h-8',
  md: 'w-12 h-12',
  lg: 'w-16 h-16',
  xl: 'w-24 h-24'
};

const fontSizes = {
  xs: 'text-xs',
  sm: 'text-xs',
  md: 'text-lg',
  lg: 'text-xl',
  xl: 'text-3xl'
};

export interface AvatarProps {
  size: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  src?: string | null;
  initials?: string;
  className?: string;
}

function AvatarImage({src, className}: { src: string; className?: string }) {
  return (
    <Image
      className={concat(className, 'rounded-full')}
      priority
      src={src}
      width={100}
      height={100}
      alt="Profile picture"
    />
  );
}

function AvatarInitials({
                          initials,
                          textClassName,
                          className
                        }: {
  initials: string;
  textClassName: string;
  className?: string;
}) {
  return (
    <span className={concat(className, 'inline-flex items-center justify-center rounded-full bg-gray-800')}>
      <span className={`leading-none text-white ${textClassName}`}>{initials}</span>
    </span>
  );
}

export default function Avatar({size, src, initials, className}: AvatarProps) {
  const avatarClasses = concat(className, sizes[size] || sizes.md);

  if (src !== undefined && src !== null && src !== '') {
    return <AvatarImage className={avatarClasses} src={src}/>;
  }

  return (
    <AvatarInitials
      className={avatarClasses}
      textClassName={fontSizes[size] || fontSizes.md}
      initials={initials ?? 'N/A'}
    />
  );
}
