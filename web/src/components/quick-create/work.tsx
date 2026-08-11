import { zodResolver } from "@hookform/resolvers/zod";
import { PlusIcon } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { QuickCreateContext } from "@/components/quick-create/context-panel";
import { MoreProperties } from "@/components/quick-create/more-properties";
import type { QuickCreateKindProps } from "@/components/quick-create/types";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { Button } from "@/components/ui/button";
import { DialogFooter } from "@/components/ui/dialog";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { getDefaultValue } from "@/lib/utils";

const workQuickCreateSchema = z.object({
  title: z.string().trim().min(1, "Title is required"),
  description: z.string().optional(),
});

type WorkQuickCreateValues = z.infer<typeof workQuickCreateSchema>;

export function WorkQuickCreate({ onCancel }: QuickCreateKindProps) {
  const form = useForm<WorkQuickCreateValues>({
    resolver: zodResolver(workQuickCreateSchema),
    defaultValues: {
      title: "",
      description: "",
    },
  });

  return (
    <FieldProvider {...form}>
      <form
        onSubmit={(event) => {
          event.preventDefault();
        }}
      >
        <FieldGroup className="my-5">
          <ControlledField
            control={form.control}
            name="title"
            render={({ field }) => (
              <Field>
                <FieldLabel>Title</FieldLabel>
                <FieldControl>
                  <Input
                    autoFocus
                    placeholder="What needs to be done?"
                    {...field}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <MoreProperties>
            <ControlledField
              control={form.control}
              name="description"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Description</FieldLabel>
                  <FieldControl>
                    <Textarea
                      placeholder="Add context (optional)"
                      rows={3}
                      {...field}
                      value={getDefaultValue(field.value)}
                    />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />
          </MoreProperties>

          <QuickCreateContext />

          <MockDataAlert title="Work creation unavailable">
            This type of item cannot be created from here yet. Drafts are not
            saved.
          </MockDataAlert>
        </FieldGroup>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled>
            <PlusIcon />
            Create unavailable
          </Button>
        </DialogFooter>
      </form>
    </FieldProvider>
  );
}
