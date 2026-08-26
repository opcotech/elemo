import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";

import type { TeamFormValues } from "./team-form-fields";
import { TeamFormFields, teamFormSchema } from "./team-form-fields";

import { FieldProvider } from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { v1OrganizationTeamsGetOptions } from "@/lib/api/query-options";
import { v1OrganizationTeamsCreate } from "@/lib/api/sdk";
import type {
  Options,
  TeamCreate,
  V1OrganizationTeamsCreateData,
} from "@/lib/api/types";
import { normalizeFormData } from "@/lib/forms";
import { showSuccessToast } from "@/lib/toast";

interface TeamCreateFormProps {
  organizationId: string;
  organizationSlug: string;
}

export function TeamCreateForm({
  organizationId,
  organizationSlug,
}: TeamCreateFormProps) {
  const navigate = useNavigate();

  const form = useForm<TeamFormValues>({
    resolver: zodResolver(teamFormSchema),
    defaultValues: {
      name: "",
      description: "",
    },
  });

  const mutation = useFormMutation<
    { id: string },
    Options<V1OrganizationTeamsCreateData>,
    TeamFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationTeamsCreate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: undefined,
    errorMessagePrefix: "Failed to create team",
    queryKeysToInvalidate: [
      v1OrganizationTeamsGetOptions({
        path: { organizationRef: organizationId },
      }).queryKey,
    ],
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        teamFormSchema,
        values
      ) as TeamCreate;
      return {
        path: { organizationRef: organizationId },
        body: normalizedBody,
      };
    },
    onSuccess: (data) => {
      showSuccessToast("Team created", "The team was created successfully");
      navigate({
        to: "/settings/organizations/$organizationSlug/teams/$teamId/edit",
        params: { organizationSlug, teamId: data.id },
      });
    },
  });

  return (
    <FormCard
      data-section="team-create-form"
      description="Create a team that can hold grants as a principal."
      onSubmit={mutation.handleSubmit}
      onCancel={() =>
        navigate({
          to: "/settings/organizations/$organizationSlug",
          params: { organizationSlug },
        })
      }
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Create Team"
    >
      <FieldProvider {...form}>
        <TeamFormFields control={form.control} isPending={mutation.isPending} />
      </FieldProvider>
    </FormCard>
  );
}
