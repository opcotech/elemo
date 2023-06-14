'use client';

import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { FormInput } from '@/components/blocks/Form/FormInput';
import { Button } from '@/components/blocks/Button';
import { getErrorMessage, UserService } from '@/lib/api';
import useStore from '@/store';

type ChangePasswordData = {
  oldPassword: string;
  newPassword: string;
  newPasswordConfirm: string;
};

const PASSWORD_SCHEMA = z
  .string()
  .min(8, 'The password must be at least 8 characters long')
  .max(64, 'The password must be at most 64 characters long');

const UPDATE_PASSWORD_SCHEMA = z.object({
  oldPassword: PASSWORD_SCHEMA,
  newPassword: PASSWORD_SCHEMA,
  newPasswordConfirm: PASSWORD_SCHEMA
});

const UPDATE_PASSWORD_SCHEMA_REFINED = UPDATE_PASSWORD_SCHEMA.refine(
  (data) => data.newPassword === data.newPasswordConfirm,
  {
    message: 'The new password and the confirmation do not match',
    path: ['newPasswordConfirm']
  }
).refine((data) => data.oldPassword !== data.newPassword, {
  message: 'The new password must be different from the old password',
  path: ['newPassword']
});

export interface ChangePasswordFormProps {
  userId: string;
}

export function ChangePasswordForm({ userId }: ChangePasswordFormProps) {
  const addMessage = useStore((state) => state.addMessage);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset
  } = useForm<ChangePasswordData>({
    resolver: zodResolver(UPDATE_PASSWORD_SCHEMA_REFINED)
  });

  async function onSubmit(data: ChangePasswordData) {
    try {
      await UserService.v1UserUpdate(userId, {
        password: data.oldPassword,
        new_password: data.newPassword
      });

      addMessage({ type: 'success', title: 'Password updated', message: `Your password has been updated.` });
      reset();
    } catch (e) {
      addMessage({ type: 'error', title: 'Failed to update password', message: getErrorMessage(e) });
    }
  }

  return (
    <form action={'#'} onSubmit={handleSubmit(onSubmit)}>
      <div className={'space-y-6'}>
        <FormInput
          type="password"
          name="oldPassword"
          label="Current password"
          placeholder="Current password"
          register={register}
          errors={errors}
          required={!UPDATE_PASSWORD_SCHEMA.shape.oldPassword.isOptional()}
        />

        <FormInput
          type="password"
          name="newPassword"
          label="New password"
          placeholder="New password"
          register={register}
          errors={errors}
          required={!UPDATE_PASSWORD_SCHEMA.shape.newPassword.isOptional()}
        />

        <FormInput
          type="password"
          name="newPasswordConfirm"
          label="Confirm new password"
          placeholder="New password again"
          register={register}
          errors={errors}
          required={!UPDATE_PASSWORD_SCHEMA.shape.newPasswordConfirm.isOptional()}
        />
      </div>

      <div className="pt-5 flex justify-end">
        <Button type={'submit'} variant="primary" loading={isSubmitting}>
          Save
        </Button>
      </div>
    </form>
  );
}
