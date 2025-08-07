import { Link } from "@tanstack/react-router";
import {
  ChevronsUpDown,
  LogOut,
  Palette,
  SettingsIcon,
  UserIcon,
} from "lucide-react";

import { useTheme } from "@/components/theme-provider";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { Skeleton } from "@/components/ui/skeleton";
import { useLogout } from "@/hooks/use-auth";
import type { User } from "@/lib/auth/types";

export function NavUser({ user }: { user: User }) {
  const { isMobile } = useSidebar();
  const { logout, isLoading: isLoggingOut } = useLogout();
  const { theme, setTheme } = useTheme();

  const getInitials = (firstName: string | null, lastName: string | null) => {
    return `${firstName?.[0] || ""}${lastName?.[0] || ""}`.toUpperCase() || "U";
  };

  return (
    <SidebarMenuItem>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <SidebarMenuButton
            size="lg"
            className="data-[state=open]:bg-primary/5 data-[state=open]:text-primary data-[state=open]:border-primary/20"
          >
            <Avatar className="h-8 w-8 rounded-lg">
              <AvatarImage
                src={user.picture || undefined}
                alt={`${user.first_name} ${user.last_name}`}
              />
              <AvatarFallback className="rounded-lg">
                {getInitials(user.first_name, user.last_name)}
              </AvatarFallback>
            </Avatar>
            <div className="grid flex-1 text-left text-sm leading-tight">
              <span className="truncate font-medium">{`${user.first_name} ${user.last_name}`}</span>
              <span className="truncate text-xs">{user.email}</span>
            </div>
            <ChevronsUpDown className="ml-auto size-4" />
          </SidebarMenuButton>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
          side={isMobile ? "bottom" : "right"}
          align="end"
          sideOffset={4}
        >
          <DropdownMenuGroup>
            <DropdownMenuItem>
              <UserIcon />
              Account
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link to="/settings">
                <SettingsIcon />
                Settings
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => setTheme(theme === "light" ? "dark" : "light")}
            >
              <Palette />
              Toggle theme
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={logout} disabled={isLoggingOut}>
            <LogOut />
            Log out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  );
}

export function NavUserSkeleton() {
  return (
    <SidebarMenuItem>
      <SidebarMenuButton
        size="lg"
        className="data-[state=open]:bg-primary/5 data-[state=open]:text-primary data-[state=open]:border-primary/20"
      >
        <div className="flex items-center space-x-4">
          <Skeleton className="size-8 rounded-lg" />
          <div className="space-y-2">
            <Skeleton className="h-3.5 w-[130px]" />
            <Skeleton className="h-2.5 w-[100px]" />
          </div>
        </div>
        <ChevronsUpDown className="ml-auto size-4" />
      </SidebarMenuButton>
    </SidebarMenuItem>
  );
}
