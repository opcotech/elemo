import { LinkIcon, PlusIcon } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { InputGroupAddon } from "@/components/ui/input-group";
import { LinkAddDialog } from "@/components/ui/link-add-dialog";
import {
  RemovableInputGroup,
  RemovableInputGroupContent,
  RemovableInputGroupRemove,
} from "@/components/ui/removable-input-group";
import { Section } from "@/components/ui/section";
import type { IssueLink } from "@/lib/api/types";
import {
  issueLinkFaviconSrc,
  issueLinkHostname,
  parseIssueLink,
} from "@/lib/work/issue-links";

function LinkFavicon({ url }: { url: string }) {
  const hostname = issueLinkHostname(url);
  const [hidden, setHidden] = useState(false);

  if (!hostname || hidden) {
    return null;
  }

  return (
    <img
      src={issueLinkFaviconSrc(hostname)}
      alt=""
      width={16}
      height={16}
      className="size-4 shrink-0"
      onError={() => {
        setHidden(true);
      }}
    />
  );
}

interface IssueLinksProps {
  links: readonly IssueLink[];
  disabled?: boolean;
  onPatch: (
    patch: { links: IssueLink[] },
    description?: string
  ) => Promise<void>;
}

export function IssueLinks({
  links,
  disabled = false,
  onPatch,
}: IssueLinksProps) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingUrl, setEditingUrl] = useState<string | null>(null);
  const [draftUrl, setDraftUrl] = useState("");
  const [draftLabel, setDraftLabel] = useState("");
  const [error, setError] = useState<string | null>(null);

  const closeDialog = (open: boolean) => {
    setDialogOpen(open);
    if (!open) {
      setEditingUrl(null);
      setDraftUrl("");
      setDraftLabel("");
      setError(null);
    }
  };

  const openAddDialog = () => {
    setEditingUrl(null);
    setDraftUrl("");
    setDraftLabel("");
    setError(null);
    setDialogOpen(true);
  };

  const openEditDialog = (link: IssueLink) => {
    setEditingUrl(link.url);
    setDraftUrl(link.url);
    setDraftLabel(link.label);
    setError(null);
    setDialogOpen(true);
  };

  const saveLink = () => {
    if (disabled) {
      return;
    }

    const parsed = parseIssueLink(draftUrl, draftLabel);
    if (!parsed.ok) {
      setError(parsed.error);
      return;
    }

    const next = links.filter(
      (link) => link.url !== parsed.url && link.url !== editingUrl
    );

    closeDialog(false);
    void onPatch(
      { links: [...next, { url: parsed.url, label: parsed.label }] },
      editingUrl ? "Link updated" : "Link added"
    );
  };

  const removeLink = (url: string) => {
    if (disabled) {
      return;
    }

    closeDialog(false);
    void onPatch(
      { links: links.filter((link) => link.url !== url) },
      "Link removed"
    );
  };

  return (
    <Section
      title="Links"
      data-section="issue-links"
      action={
        <Button
          type="button"
          variant="ghost"
          size="xs"
          disabled={disabled}
          onClick={openAddDialog}
        >
          <PlusIcon />
          Add
        </Button>
      }
    >
      {links.length === 0 ? (
        <EmptyState
          compact
          icon={<LinkIcon />}
          title="No links"
          description="External URLs will appear here."
        />
      ) : (
        <div className="space-y-2">
          {links.map((link) => {
            const hostname = issueLinkHostname(link.url);
            const displayLabel =
              link.label && link.label !== link.url
                ? link.label
                : (hostname ?? link.url);
            return (
              <RemovableInputGroup key={link.url}>
                <InputGroupAddon align="inline-start" className="pl-2">
                  <LinkFavicon url={link.url} />
                </InputGroupAddon>
                <RemovableInputGroupContent className="gap-2">
                  <button
                    type="button"
                    disabled={disabled}
                    aria-label={`Edit link ${displayLabel}`}
                    title="Edit link"
                    className="min-w-0 truncate text-left font-medium"
                    onClick={() => {
                      openEditDialog(link);
                    }}
                  >
                    {displayLabel}
                  </button>
                  <a
                    href={link.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-muted-foreground min-w-0 truncate underline-offset-4 hover:underline"
                  >
                    {hostname ?? link.url}
                  </a>
                </RemovableInputGroupContent>
                <RemovableInputGroupRemove
                  disabled={disabled}
                  aria-label={`Remove link ${displayLabel}`}
                  title="Remove link"
                  onClick={() => {
                    removeLink(link.url);
                  }}
                />
              </RemovableInputGroup>
            );
          })}
        </div>
      )}
      <LinkAddDialog
        open={dialogOpen}
        onOpenChange={closeDialog}
        url={draftUrl}
        onUrlChange={(next) => {
          setDraftUrl(next);
          if (error) {
            setError(null);
          }
        }}
        label={draftLabel}
        onLabelChange={(next) => {
          setDraftLabel(next);
          if (error) {
            setError(null);
          }
        }}
        showLabel
        error={error}
        title={editingUrl ? "Edit link" : "Add link"}
        description="Set the URL and the visible label for this link."
        submitLabel={editingUrl ? "Save" : "Add link"}
        onSubmit={saveLink}
        onRemove={
          editingUrl
            ? () => {
                removeLink(editingUrl);
              }
            : undefined
        }
      />
    </Section>
  );
}
