import {
  DndContext,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import type { DragEndEvent } from "@dnd-kit/core";
import {
  SortableContext,
  arrayMove,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { GripVertical } from "lucide-react";
import { useMemo, useState } from "react";
import type { ReactNode } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { EntityMultiSelect } from "@/components/ui/entity-select";
import { FormCard } from "@/components/ui/form-card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import {
  v1CustomFieldArchiveMutation,
  v1CustomFieldDeleteMutation,
  v1CustomFieldUpdateMutation,
  v1CustomFieldsCreateMutation,
} from "@/lib/api/mutation-options";
import { v1CustomFieldsGetOptions } from "@/lib/api/query-options";
import type {
  CustomFieldDefinition,
  CustomFieldKind,
  CustomFieldOption,
  CustomFieldSchema,
  ResourceType,
} from "@/lib/api/types";
import {
  customFieldKindAllowsFullText,
  customFieldKindAllowsRange,
  customFieldKindLabels,
  customFieldKinds,
  defaultCustomFieldSchema,
  nodeResourceTypes,
} from "@/lib/custom-fields/value";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { cn } from "@/lib/utils";

export type CustomFieldScopeType = "Organization" | "Namespace" | "Project";

function scopeNoun(scopeType: CustomFieldScopeType): string {
  switch (scopeType) {
    case "Organization":
      return "organization";
    case "Namespace":
      return "namespace";
    case "Project":
      return "project";
  }
}

function capitalize(value: string): string {
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function schemaForCreate(schema: CustomFieldSchema): CustomFieldSchema {
  if (schema.kind === "url" && schema.allowed_schemes.length === 0) {
    return { ...schema, allowed_schemes: ["https"] };
  }
  return schema;
}

function Labeled({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      {children}
    </div>
  );
}

function SelectOptionsEditor({
  options,
  onChange,
}: {
  options: CustomFieldOption[];
  onChange: (options: CustomFieldOption[]) => void;
}) {
  return (
    <div className="space-y-2">
      {options.map((option, index) => (
        <div key={index} className="grid grid-cols-2 gap-2">
          <Input
            value={option.key}
            placeholder="key"
            onChange={(event) => {
              const next = [...options];
              next[index] = { ...option, key: event.target.value };
              onChange(next);
            }}
          />
          <Input
            value={option.label}
            placeholder="Label"
            onChange={(event) => {
              const next = [...options];
              next[index] = { ...option, label: event.target.value };
              onChange(next);
            }}
          />
        </div>
      ))}
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={() =>
          onChange([
            ...options,
            {
              key: `option_${options.length + 1}`,
              label: `Option ${options.length + 1}`,
              disabled: false,
            },
          ])
        }
      >
        Add option
      </Button>
    </div>
  );
}

function SchemaFields({
  schema,
  onChange,
}: {
  schema: CustomFieldSchema;
  onChange: (schema: CustomFieldSchema) => void;
}) {
  if (schema.kind === "text") {
    return (
      <Labeled label="Max length">
        <Input
          type="number"
          value={schema.max_length ?? ""}
          onChange={(event) =>
            onChange({
              ...schema,
              max_length: event.target.value
                ? Number(event.target.value)
                : undefined,
            })
          }
        />
      </Labeled>
    );
  }
  if (schema.kind === "integer") {
    return (
      <div className="grid grid-cols-2 gap-3">
        <Labeled label="Min">
          <Input
            type="number"
            value={schema.min ?? ""}
            onChange={(event) =>
              onChange({
                ...schema,
                min: event.target.value
                  ? Number(event.target.value)
                  : undefined,
              })
            }
          />
        </Labeled>
        <Labeled label="Max">
          <Input
            type="number"
            value={schema.max ?? ""}
            onChange={(event) =>
              onChange({
                ...schema,
                max: event.target.value
                  ? Number(event.target.value)
                  : undefined,
              })
            }
          />
        </Labeled>
      </div>
    );
  }
  if (schema.kind === "decimal") {
    return (
      <div className="grid grid-cols-3 gap-3">
        <Labeled label="Min">
          <Input
            value={schema.min ?? ""}
            onChange={(event) =>
              onChange({
                ...schema,
                min: event.target.value || undefined,
              })
            }
          />
        </Labeled>
        <Labeled label="Max">
          <Input
            value={schema.max ?? ""}
            onChange={(event) =>
              onChange({
                ...schema,
                max: event.target.value || undefined,
              })
            }
          />
        </Labeled>
        <Labeled label="Scale">
          <Input
            type="number"
            value={schema.scale ?? ""}
            onChange={(event) =>
              onChange({
                ...schema,
                scale: event.target.value
                  ? Number(event.target.value)
                  : undefined,
              })
            }
          />
        </Labeled>
      </div>
    );
  }
  if (schema.kind === "url") {
    return (
      <Labeled label="Allowed schemes">
        <Input
          value={schema.allowed_schemes.join(",")}
          onChange={(event) => {
            const schemes = event.target.value
              .split(",")
              .map((item) => item.trim())
              .filter(Boolean);
            onChange({
              ...schema,
              allowed_schemes: schemes,
            });
          }}
        />
      </Labeled>
    );
  }
  if (schema.kind === "single_select" || schema.kind === "multi_select") {
    return (
      <Labeled label="Options">
        <SelectOptionsEditor
          options={schema.options}
          onChange={(options) => onChange({ ...schema, options })}
        />
      </Labeled>
    );
  }
  if (schema.kind === "user_reference") {
    return (
      <label className="flex items-center gap-2 text-sm">
        <Checkbox
          checked={Boolean(schema.multiple)}
          onCheckedChange={(checked) =>
            onChange({ ...schema, multiple: Boolean(checked) })
          }
        />
        Allow multiple values
      </label>
    );
  }
  if (schema.kind === "resource_reference") {
    const types = nodeResourceTypes();
    return (
      <div className="space-y-3">
        <Labeled label="Allowed types">
          <EntityMultiSelect
            options={types.map((type) => ({
              value: type,
              title: type,
            }))}
            value={schema.allowed_types}
            placeholder="Select resource types"
            searchPlaceholder="Search types…"
            emptyMessage="No types found."
            aria-label="Allowed types"
            onValueChange={(next) =>
              onChange({
                ...schema,
                allowed_types:
                  next.length > 0 ? (next as ResourceType[]) : ["Issue"],
              })
            }
          />
        </Labeled>
        <label className="flex items-center gap-2 text-sm">
          <Checkbox
            checked={Boolean(schema.multiple)}
            onCheckedChange={(checked) =>
              onChange({ ...schema, multiple: Boolean(checked) })
            }
          />
          Allow multiple values
        </label>
      </div>
    );
  }
  return null;
}

function CreateDefinitionForm({
  scopeId,
  scopeType,
  onCreated,
}: {
  scopeId: string;
  scopeType: CustomFieldScopeType;
  onCreated: () => void;
}) {
  const [key, setKey] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [kind, setKind] = useState<CustomFieldKind>("text");
  const [required, setRequired] = useState(false);
  const [indexExact, setIndexExact] = useState(false);
  const [indexRange, setIndexRange] = useState(false);
  const [indexFullText, setIndexFullText] = useState(false);
  const [schema, setSchema] = useState<CustomFieldSchema>(
    defaultCustomFieldSchema("text")
  );
  const createMutation = useMutation(v1CustomFieldsCreateMutation());

  return (
    <FormCard
      title="Create field"
      description="The field is usable as soon as it is created."
      submitButtonText="Create field"
      isPending={createMutation.isPending}
      error={
        createMutation.error instanceof Error ? createMutation.error : null
      }
      onSubmit={(event) => {
        event.preventDefault();
        void createMutation
          .mutateAsync({
            body: {
              key,
              name,
              description: description || undefined,
              kind,
              scope_id: scopeId,
              scope_type: scopeType,
              target_type: "Issue",
              required,
              index_exact: indexExact,
              index_range: indexRange,
              index_fulltext: indexFullText,
              schema: schemaForCreate(schema),
            },
          })
          .then(() => {
            showSuccessToast(
              "Custom field created",
              `The field is available on issues in this ${scopeNoun(scopeType)}.`
            );
            setKey("");
            setName("");
            setDescription("");
            onCreated();
          })
          .catch((error: unknown) => {
            showErrorToast(
              "Failed to create custom field",
              error instanceof Error ? error : "Unknown error"
            );
          });
      }}
    >
      <div className="space-y-5">
        <Labeled label="Key">
          <Input
            value={key}
            placeholder="story_points"
            onChange={(event) => setKey(event.target.value)}
          />
        </Labeled>
        <Labeled label="Name">
          <Input
            value={name}
            placeholder="Story points"
            onChange={(event) => setName(event.target.value)}
          />
        </Labeled>
        <Labeled label="Description">
          <Textarea
            value={description}
            onChange={(event) => setDescription(event.target.value)}
          />
        </Labeled>
        <Labeled label="Type">
          <Select
            value={kind}
            onValueChange={(next) => {
              const nextKind = (next ?? "text") as CustomFieldKind;
              setKind(nextKind);
              setSchema(defaultCustomFieldSchema(nextKind));
              if (!customFieldKindAllowsRange(nextKind)) {
                setIndexRange(false);
              }
              if (!customFieldKindAllowsFullText(nextKind)) {
                setIndexFullText(false);
              }
            }}
            items={Object.fromEntries(
              customFieldKinds.map((item) => [
                item,
                customFieldKindLabels[item],
              ])
            )}
          >
            <SelectTrigger className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {customFieldKinds.map((item) => (
                <SelectItem key={item} value={item}>
                  {customFieldKindLabels[item]}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Labeled>
        <SchemaFields schema={schema} onChange={setSchema} />
        <label className="flex items-center gap-2 text-sm">
          <Checkbox
            checked={required}
            onCheckedChange={(checked) => setRequired(Boolean(checked))}
          />
          Required on create and later writes
        </label>
        <label className="flex items-center gap-2 text-sm">
          <Checkbox
            checked={indexExact}
            onCheckedChange={(checked) => setIndexExact(Boolean(checked))}
          />
          Index exact matches
        </label>
        {customFieldKindAllowsRange(kind) ? (
          <label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={indexRange}
              onCheckedChange={(checked) => setIndexRange(Boolean(checked))}
            />
            Index range queries
          </label>
        ) : null}
        {customFieldKindAllowsFullText(kind) ? (
          <label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={indexFullText}
              onCheckedChange={(checked) => setIndexFullText(Boolean(checked))}
            />
            Index full text
          </label>
        ) : null}
      </div>
    </FormCard>
  );
}

function isLocalActive(
  definition: CustomFieldDefinition,
  scopeId: string
): boolean {
  return definition.scope_id === scopeId && !definition.archived;
}

function DefinitionRowBody({
  definition,
  noun,
  inherited,
}: {
  definition: CustomFieldDefinition;
  noun: string;
  inherited: boolean;
}) {
  return (
    <div className="space-y-1">
      <div className="flex flex-wrap items-center gap-2">
        <p className="font-medium">{definition.name}</p>
        <Badge variant="secondary">
          {customFieldKindLabels[definition.kind]}
        </Badge>
        {definition.required ? <Badge>Required</Badge> : null}
        {definition.archived ? <Badge variant="outline">Archived</Badge> : null}
        {inherited ? (
          <Badge variant="outline">
            Inherited from {definition.scope_type.toLowerCase()}
          </Badge>
        ) : (
          <Badge variant="outline">This {noun}</Badge>
        )}
      </div>
      <p className="text-muted-foreground text-xs">
        {definition.key}
        {definition.description ? ` · ${definition.description}` : ""}
      </p>
    </div>
  );
}

function DefinitionRowActions({
  definition,
  canManage,
  inherited,
  onArchive,
  onDelete,
}: {
  definition: CustomFieldDefinition;
  canManage: boolean;
  inherited: boolean;
  onArchive: (definition: CustomFieldDefinition) => void;
  onDelete: (definition: CustomFieldDefinition) => void;
}) {
  if (!canManage || inherited) {
    return null;
  }
  return (
    <div className="flex gap-2">
      {!definition.archived ? (
        <Button
          variant="outline"
          size="sm"
          onClick={() => onArchive(definition)}
        >
          Archive
        </Button>
      ) : null}
      <Button
        variant="destructive"
        size="sm"
        onClick={() => onDelete(definition)}
      >
        Delete
      </Button>
    </div>
  );
}

function SortableDefinitionRow({
  definition,
  noun,
  inherited,
  disabled,
  canManage,
  onArchive,
  onDelete,
}: {
  definition: CustomFieldDefinition;
  noun: string;
  inherited: boolean;
  disabled: boolean;
  canManage: boolean;
  onArchive: (definition: CustomFieldDefinition) => void;
  onDelete: (definition: CustomFieldDefinition) => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: definition.id,
    disabled,
  });

  return (
    <li
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
      }}
      className={cn(
        "flex items-start justify-between gap-4 py-3",
        isDragging && "opacity-40"
      )}
      data-custom-field-key={definition.key}
    >
      <div className="flex min-w-0 flex-1 items-start gap-1">
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="mt-0.5 cursor-grab active:cursor-grabbing"
          aria-label={`Reorder ${definition.name}`}
          disabled={disabled}
          {...attributes}
          {...listeners}
        >
          <GripVertical />
        </Button>
        <DefinitionRowBody
          definition={definition}
          noun={noun}
          inherited={inherited}
        />
      </div>
      <DefinitionRowActions
        definition={definition}
        canManage={canManage}
        inherited={inherited}
        onArchive={onArchive}
        onDelete={onDelete}
      />
    </li>
  );
}

function StaticDefinitionRow({
  definition,
  noun,
  inherited,
  showHandleSlot,
  canManage,
  onArchive,
  onDelete,
}: {
  definition: CustomFieldDefinition;
  noun: string;
  inherited: boolean;
  showHandleSlot: boolean;
  canManage: boolean;
  onArchive: (definition: CustomFieldDefinition) => void;
  onDelete: (definition: CustomFieldDefinition) => void;
}) {
  return (
    <li
      className="flex items-start justify-between gap-4 py-3"
      data-custom-field-key={definition.key}
    >
      <div className="flex min-w-0 flex-1 items-start gap-1">
        {showHandleSlot ? <span className="mt-0.5 size-8 shrink-0" /> : null}
        <DefinitionRowBody
          definition={definition}
          noun={noun}
          inherited={inherited}
        />
      </div>
      <DefinitionRowActions
        definition={definition}
        canManage={canManage}
        inherited={inherited}
        onArchive={onArchive}
        onDelete={onDelete}
      />
    </li>
  );
}

function DefinitionList({
  definitions,
  scopeId,
  noun,
  canManage,
  reordering,
  onArchive,
  onDelete,
  onReorder,
}: {
  definitions: CustomFieldDefinition[];
  scopeId: string;
  noun: string;
  canManage: boolean;
  reordering: boolean;
  onArchive: (definition: CustomFieldDefinition) => void;
  onDelete: (definition: CustomFieldDefinition) => void;
  onReorder: (ordered: CustomFieldDefinition[]) => void;
}) {
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 8 },
    })
  );
  const sortableIds = useMemo(
    () =>
      canManage
        ? definitions
            .filter((definition) => isLocalActive(definition, scopeId))
            .map((definition) => definition.id)
        : [],
    [canManage, definitions, scopeId]
  );

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) {
      return;
    }
    const local = definitions.filter((definition) =>
      isLocalActive(definition, scopeId)
    );
    const oldIndex = local.findIndex(
      (definition) => definition.id === active.id
    );
    const newIndex = local.findIndex((definition) => definition.id === over.id);
    if (oldIndex < 0 || newIndex < 0) {
      return;
    }
    onReorder(arrayMove(local, oldIndex, newIndex));
  };

  const rows = definitions.map((definition) => {
    const inherited = definition.scope_id !== scopeId;
    const sortable = sortableIds.includes(definition.id);
    if (sortable) {
      return (
        <SortableDefinitionRow
          key={definition.id}
          definition={definition}
          noun={noun}
          inherited={inherited}
          disabled={reordering}
          canManage={canManage}
          onArchive={onArchive}
          onDelete={onDelete}
        />
      );
    }
    return (
      <StaticDefinitionRow
        key={definition.id}
        definition={definition}
        noun={noun}
        inherited={inherited}
        showHandleSlot={sortableIds.length > 0}
        canManage={canManage}
        onArchive={onArchive}
        onDelete={onDelete}
      />
    );
  });

  const list = <ul className="divide-border divide-y">{rows}</ul>;
  if (sortableIds.length === 0) {
    return list;
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <SortableContext
        items={sortableIds}
        strategy={verticalListSortingStrategy}
      >
        {list}
      </SortableContext>
    </DndContext>
  );
}

export function CustomFieldDefinitionManager({
  scopeId,
  scopeType,
  canManage,
}: {
  scopeId: string;
  scopeType: CustomFieldScopeType;
  canManage: boolean;
}) {
  const queryClient = useQueryClient();
  const noun = scopeNoun(scopeType);
  const listOptions = v1CustomFieldsGetOptions({
    query: {
      scope_id: scopeId,
      scope_type: scopeType,
      target_type: "Issue",
      include_archived: true,
    },
  });
  const { data, isPending, isError, error: listError } = useQuery(listOptions);
  const archiveMutation = useMutation(v1CustomFieldArchiveMutation());
  const deleteMutation = useMutation(v1CustomFieldDeleteMutation());
  const updateMutation = useMutation(v1CustomFieldUpdateMutation());
  const [pendingDelete, setPendingDelete] = useState<CustomFieldDefinition>();

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: listOptions.queryKey });

  const definitions = data ?? [];

  const archiveDefinition = (definition: CustomFieldDefinition) => {
    void archiveMutation
      .mutateAsync({ path: { id: definition.id } })
      .then(() => {
        showSuccessToast(
          "Custom field archived",
          "Stored values remain readable."
        );
        return invalidate();
      })
      .catch((error: unknown) => {
        showErrorToast(
          "Failed to archive custom field",
          error instanceof Error ? error : "Unknown error"
        );
      });
  };

  const reorderDefinitions = (ordered: CustomFieldDefinition[]) => {
    if (ordered.every((definition, index) => definition.order === index)) {
      return;
    }
    void Promise.all(
      ordered.map((definition, index) =>
        updateMutation.mutateAsync({
          path: { id: definition.id },
          body: { order: index },
        })
      )
    )
      .then(() => {
        showSuccessToast(
          "Custom fields reordered",
          "The display order was saved."
        );
        return invalidate();
      })
      .catch((error: unknown) => {
        showErrorToast(
          "Failed to reorder custom fields",
          error instanceof Error ? error : "Unknown error"
        );
      });
  };

  return (
    <div className="space-y-6">
      <Card data-section="custom-fields">
        <CardHeader>
          <CardTitle>Issue fields</CardTitle>
          <CardDescription>
            {capitalize(noun)} fields are editable.{" "}
            {scopeType === "Organization"
              ? "They apply to issues across this organization."
              : "Fields defined on ancestor scopes are inherited."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isPending ? (
            <div className="space-y-3">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          ) : isError ? (
            <Alert variant="destructive">
              <AlertTitle>Failed to load custom fields</AlertTitle>
              <AlertDescription>
                {listError instanceof Error
                  ? listError.message
                  : "Custom fields could not be loaded. Please try again later."}
              </AlertDescription>
            </Alert>
          ) : definitions.length === 0 ? (
            <p className="text-muted-foreground text-sm">
              No custom fields apply to issues in this {noun} yet.
            </p>
          ) : (
            <DefinitionList
              definitions={definitions}
              scopeId={scopeId}
              noun={noun}
              canManage={canManage}
              reordering={updateMutation.isPending}
              onArchive={archiveDefinition}
              onDelete={setPendingDelete}
              onReorder={reorderDefinitions}
            />
          )}
        </CardContent>
      </Card>

      {canManage ? (
        <CreateDefinitionForm
          scopeId={scopeId}
          scopeType={scopeType}
          onCreated={() => {
            void invalidate();
          }}
        />
      ) : null}

      <DeleteConfirmationDialog
        open={Boolean(pendingDelete)}
        onOpenChange={(open) => {
          if (!open) {
            setPendingDelete(undefined);
          }
        }}
        title={
          pendingDelete
            ? `Delete ${pendingDelete.name}?`
            : "Delete custom field?"
        }
        description="Hard delete is allowed only when the field has no stored values."
        consequences={[
          "The definition will be removed permanently",
          "Fields with stored values must be archived instead",
        ]}
        deleteButtonText="Delete field"
        isPending={deleteMutation.isPending}
        onConfirm={() => {
          if (!pendingDelete) {
            return;
          }
          void deleteMutation
            .mutateAsync({ path: { id: pendingDelete.id } })
            .then(() => {
              showSuccessToast(
                "Custom field deleted",
                "The definition was removed."
              );
              setPendingDelete(undefined);
              return invalidate();
            })
            .catch((error: unknown) => {
              showErrorToast(
                "Failed to delete custom field",
                error instanceof Error ? error : "Unknown error"
              );
            });
        }}
      />
    </div>
  );
}
