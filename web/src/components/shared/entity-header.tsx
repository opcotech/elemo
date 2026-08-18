import { CopyIcon, MoreHorizontalIcon } from "lucide-react";
import { Fragment } from "react";
import type { ReactNode } from "react";

import { EntityIcon } from "@/components/shared/entity-link";
import type { AppEntityType } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { InternalLink } from "@/components/ui/internal-link";
import { PageHeader } from "@/components/ui/page-header";
import { internalPath } from "@/lib/internal-url";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

async function writeClipboardText(value: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value);
      return;
    } catch {
      // Firefox/WebKit often reject the Clipboard API without a Chromium-style
      // permission grant. Fall back to a user-gesture execCommand copy.
    }
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  const copied = document.execCommand("copy");
  textarea.remove();
  if (!copied) {
    throw new Error("Clipboard unavailable");
  }
}

export { PageHeader };

export function EntityHeader({
  type,
  eyebrow,
  title,
  description,
  meta,
  actions,
  copyValue,
  copyLabel = "Copy ID",
  showIcon = true,
  imageUrl,
}: {
  type: AppEntityType;
  eyebrow?: ReactNode;
  title: ReactNode;
  description?: ReactNode;
  meta?: ReactNode;
  actions?: ReactNode;
  copyValue?: string;
  copyLabel?: string;
  showIcon?: boolean;
  imageUrl?: string | null;
}) {
  const resolvedImageUrl = imageUrl?.trim() ? imageUrl : null;

  return (
    <PageHeader
      leading={
        showIcon ? (
          resolvedImageUrl ? (
            <span className="bg-muted size-11 shrink-0 overflow-hidden rounded-xl">
              <img
                src={resolvedImageUrl}
                alt=""
                className="size-full object-cover"
              />
            </span>
          ) : (
            <span className="bg-primary-subtle text-primary-on-subtle flex size-11 shrink-0 items-center justify-center rounded-xl">
              <EntityIcon type={type} className="size-5" />
            </span>
          )
        ) : undefined
      }
      eyebrow={
        eyebrow ? (
          <>
            <span>{eyebrow}</span>
            {copyValue && (
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                className="text-muted-foreground hover:text-foreground size-5"
                aria-label={copyLabel}
                title={copyLabel}
                onClick={() => {
                  void writeClipboardText(copyValue).then(
                    () => {
                      showSuccessToast("Copied", copyValue);
                    },
                    (error: unknown) => {
                      showErrorToast(
                        "Could not copy",
                        error instanceof Error ? error : "Clipboard unavailable"
                      );
                    }
                  );
                }}
              >
                <CopyIcon />
              </Button>
            )}
          </>
        ) : undefined
      }
      title={title}
      description={description}
      meta={meta}
      actions={actions}
    />
  );
}

export function PageActions({
  primary,
  secondary = [],
  size = "icon-sm",
}: {
  primary?: ReactNode;
  secondary?: {
    label: string;
    href?: string;
    onSelect?: () => void;
    disabled?: boolean;
    variant?: "default" | "destructive";
  }[];
  size?: "icon-sm" | "icon-xs";
}) {
  return (
    <>
      {primary}
      {secondary.length > 0 && (
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button
                variant={size === "icon-xs" ? "ghost" : "outline"}
                size={size}
                aria-label="More actions"
              />
            }
          >
            <MoreHorizontalIcon />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {secondary.map((item, index) => (
              <Fragment key={item.label}>
                {item.variant === "destructive" && index > 0 && (
                  <DropdownMenuSeparator />
                )}
                <DropdownMenuItem
                  variant={item.variant}
                  disabled={item.disabled}
                  render={
                    item.href ? (
                      <InternalLink to={internalPath(item.href)} />
                    ) : undefined
                  }
                  onClick={item.onSelect}
                >
                  {item.label}
                </DropdownMenuItem>
              </Fragment>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      )}
    </>
  );
}
