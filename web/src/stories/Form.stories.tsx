import { zodResolver } from "@hookform/resolvers/zod";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";

const formSchema = z.object({
  username: z.string().min(2, "Username must be at least 2 characters."),
  email: z.string().email("Enter a valid email address."),
});

type FormValues = z.infer<typeof formSchema>;

function ExampleForm({
  defaultValues,
  validateOnMount = false,
}: {
  defaultValues?: Partial<FormValues>;
  validateOnMount?: boolean;
}) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      email: "",
      ...defaultValues,
    },
    mode: validateOnMount ? "onChange" : "onSubmit",
  });

  useEffect(() => {
    if (validateOnMount) {
      void form.trigger();
    }
  }, [form, validateOnMount]);

  return (
    <FieldProvider {...form}>
      <form
        onSubmit={form.handleSubmit(() => undefined)}
        className="w-90 space-y-4"
        noValidate
      >
        <FieldGroup>
          <ControlledField
            control={form.control}
            name="username"
            render={({ field }) => (
              <Field>
                <FieldLabel>Username</FieldLabel>
                <FieldControl>
                  <Input placeholder="johndoe" {...field} />
                </FieldControl>
                <FieldDescription>Your public display name.</FieldDescription>
                <FieldError />
              </Field>
            )}
          />
          <ControlledField
            control={form.control}
            name="email"
            render={({ field }) => (
              <Field>
                <FieldLabel>Email</FieldLabel>
                <FieldControl>
                  <Input
                    type="email"
                    placeholder="you@example.com"
                    {...field}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />
        </FieldGroup>
        <Button type="submit">Submit</Button>
      </form>
    </FieldProvider>
  );
}

const meta: Meta = {
  title: "UI/Field",
  parameters: {
    a11y: { test: "error" },
    layout: "centered",
    docs: {
      description: {
        component:
          "Field composition wired to React Hook Form with accessible descriptions, invalid state, and alert messages.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <ExampleForm />,
};

export const Invalid: Story = {
  render: () => (
    <ExampleForm
      defaultValues={{ username: "a", email: "not-an-email" }}
      validateOnMount
    />
  ),
};

export const ResponsiveOrientation: Story = {
  render: () => {
    const form = useForm<FormValues>({
      resolver: zodResolver(formSchema),
      defaultValues: { username: "ada", email: "ada@example.test" },
    });

    return (
      <FieldProvider {...form}>
        <form
          onSubmit={form.handleSubmit(() => undefined)}
          className="w-105 space-y-4"
          noValidate
        >
          <FieldGroup>
            <ControlledField
              control={form.control}
              name="username"
              render={({ field }) => (
                <Field orientation="responsive">
                  <FieldLabel>Username</FieldLabel>
                  <FieldControl>
                    <Input {...field} />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />
          </FieldGroup>
          <Button type="submit">Save</Button>
        </form>
      </FieldProvider>
    );
  },
};
