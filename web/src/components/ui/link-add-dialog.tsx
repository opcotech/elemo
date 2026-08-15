import { useId } from "react";
import type { FormEvent } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function LinkAddDialog({
  open,
  onOpenChange,
  url,
  onUrlChange,
  label = "",
  onLabelChange,
  showLabel = false,
  error,
  title = "Add link",
  description,
  submitLabel = "Add link",
  submitDisabled = false,
  onSubmit,
  onRemove,
  removeLabel = "Remove link",
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  url: string;
  onUrlChange: (url: string) => void;
  label?: string;
  onLabelChange?: (label: string) => void;
  showLabel?: boolean;
  error?: string | null;
  title?: string;
  description?: string;
  submitLabel?: string;
  submitDisabled?: boolean;
  onSubmit: () => void;
  onRemove?: () => void;
  removeLabel?: string;
}) {
  const id = useId();
  const urlId = `${id}-url`;
  const labelId = `${id}-label`;
  const canSubmit =
    !submitDisabled &&
    Boolean(url.trim()) &&
    (!showLabel || Boolean(label.trim()));

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!canSubmit) {
      return;
    }
    onSubmit();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={handleSubmit} className="flex flex-col gap-y-6">
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            {description ? (
              <DialogDescription>{description}</DialogDescription>
            ) : null}
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <Label htmlFor={urlId}>URL</Label>
              <Input
                id={urlId}
                type="text"
                inputMode="url"
                autoComplete="url"
                value={url}
                placeholder="https://"
                autoFocus
                aria-invalid={error != null}
                onChange={(event) => onUrlChange(event.target.value)}
              />
              {error ? (
                <p className="text-destructive text-xs">{error}</p>
              ) : null}
            </div>
            {showLabel ? (
              <div className="flex flex-col gap-2">
                <Label htmlFor={labelId}>Label</Label>
                <Input
                  id={labelId}
                  type="text"
                  value={label}
                  placeholder="Link text"
                  onChange={(event) => onLabelChange?.(event.target.value)}
                />
              </div>
            ) : null}
          </div>

          <DialogFooter>
            {onRemove ? (
              <Button
                type="button"
                variant="destructive-ghost"
                className="sm:mr-auto"
                onClick={onRemove}
              >
                {removeLabel}
              </Button>
            ) : null}
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={!canSubmit}>
              {submitLabel}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
