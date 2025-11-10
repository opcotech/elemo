import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getInitials } from "@/lib/utils";

const avatarSizeVariants = cva("", {
  variants: {
    size: {
      sm: "h-8 w-8",
      md: "h-10 w-10",
      lg: "h-12 w-12",
    },
  },
  defaultVariants: {
    size: "md",
  },
});

interface UserAvatarProps {
  firstName: string;
  lastName: string;
  email?: string;
  picture?: string | null;
  size?: VariantProps<typeof avatarSizeVariants>["size"];
  showEmail?: boolean;
  className?: string;
}

export function UserAvatar({
  firstName,
  lastName,
  email,
  picture,
  size = "md",
  showEmail = false,
  className,
}: UserAvatarProps) {
  const fullName = `${firstName} ${lastName}`;
  const initials = getInitials(firstName, lastName);

  return (
    <div className={`flex items-center gap-3 ${className || ""}`}>
      <Avatar className={avatarSizeVariants({ size })}>
        <AvatarImage src={picture || undefined} alt={fullName} />
        <AvatarFallback>{initials}</AvatarFallback>
      </Avatar>
      <div className="flex flex-col gap-0.5">
        <span className="font-medium">{fullName}</span>
        {showEmail && email && (
          <span className="text-muted-foreground text-sm">{email}</span>
        )}
      </div>
    </div>
  );
}

interface UserAvatarCompactProps {
  firstName: string;
  lastName: string;
  picture?: string | null;
  size?: VariantProps<typeof avatarSizeVariants>["size"];
}

export function UserAvatarCompact({
  firstName,
  lastName,
  picture,
  size = "md",
}: UserAvatarCompactProps) {
  const fullName = `${firstName} ${lastName}`;
  const initials = getInitials(firstName, lastName);

  return (
    <Avatar className={avatarSizeVariants({ size })}>
      <AvatarImage src={picture || undefined} alt={fullName} />
      <AvatarFallback>{initials}</AvatarFallback>
    </Avatar>
  );
}
