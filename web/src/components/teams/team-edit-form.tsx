import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { TeamFormFields, teamFormSchema } from "./team-form-fields";
import { TeamMemberAssignment } from "./team-member-assignment";

import { FieldProvider } from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { useFormMutation } from "@/hooks/use-form-mutation";
import {
  v1OrganizationTeamGetOptions,
  v1OrganizationTeamsGetOptions,
} from "@/lib/api/query-options";
import { zTeamPatch } from "@/lib/api/schemas";
import { v1OrganizationTeamUpdate } from "@/lib/api/sdk";
import type {
  Options,
  Team,
  V1OrganizationTeamUpdateData,
} from "@/lib/api/types";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const teamEditFormSchema = createFormSchema(
  zTeamPatch.extend({
    name: teamFormSchema.shape.name,
  })
);

type TeamEditFormValues = z.infer<typeof teamEditFormSchema>;

interface TeamEditFormProps {
  team: Team;
  organizationId: string;
  teamId: string;
}

export function TeamEditForm({
  team,
  organizationId,
  teamId,
}: TeamEditFormProps) {
  const navigate = useNavigate();

  const form = useForm<TeamEditFormValues>({
    resolver: zodResolver(teamEditFormSchema),
    defaultValues: {
      name: team.name,
      description: getDefaultValue(team.description),
    },
  });

  useEffect(() => {
    if (!form.formState.isDirty) {
      form.reset({
        name: team.name,
        description: getDefaultValue(team.description),
      });
    }
  }, [team.name, team.description, form]);

  const mutation = useFormMutation<
    Team,
    Options<V1OrganizationTeamUpdateData>,
    TeamEditFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationTeamUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Team updated",
    errorMessagePrefix: "Failed to update team",
    queryKeysToInvalidate: [
      v1OrganizationTeamsGetOptions({
        path: { id: organizationId },
      }).queryKey,
      v1OrganizationTeamGetOptions({
        path: {
          id: organizationId,
          team_id: teamId,
        },
      }).queryKey,
    ],
    navigateOnSuccess: (navigateTo) =>
      navigateTo({
        to: "/settings/organizations/$organizationId",
        params: { organizationId },
      }),
    transformValues: (values) => {
      const normalizedBody = normalizePatchData(teamEditFormSchema, values, {
        name: team.name,
        description: team.description,
      });
      return {
        path: {
          id: organizationId,
          team_id: teamId,
        },
        body: normalizedBody,
      };
    },
  });

  return (
    <div className="flex flex-col gap-6">
      <FormCard
        data-section="team-edit-form"
        onSubmit={mutation.handleSubmit}
        onCancel={() =>
          navigate({
            to: "/settings/organizations/$organizationId",
            params: { organizationId },
          })
        }
        isPending={mutation.isPending}
        error={mutation.error || null}
        submitButtonText="Save Changes"
        description="Update the team details below."
      >
        <FieldProvider {...form}>
          <TeamFormFields
            control={form.control}
            isPending={mutation.isPending}
          />
        </FieldProvider>
      </FormCard>
      <TeamMemberAssignment
        organizationId={organizationId}
        teamId={teamId}
        teamName={team.name}
      />
    </div>
  );
}
