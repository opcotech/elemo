'use client';

import { z } from 'zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { FormInput } from '@/components/blocks/Form/FormInput';
import { Button } from '@/components/blocks/Button';
import { getErrorMessage, UsersService } from '@/lib/api';
import useStore from '@/store';
import { useSession } from 'next-auth/react';

const PASSWORD_SCHEMA = z
  .string()
  .min(8, 'The password must be at least 8 characters long')
  .max(64, 'The password must be at most 64 characters long');

const UPDATE_PASSWORD_SCHEMA = z
  .object({
    oldPassword: PASSWORD_SCHEMA,
    newPassword: PASSWORD_SCHEMA,
    newPasswordConfirm: PASSWORD_SCHEMA
  })
  .refine((data) => data.newPassword === data.newPasswordConfirm, {
    message: 'The new password and the confirmation do not match',
    path: ['newPasswordConfirm']
  })
  .refine((data) => data.oldPassword !== data.newPassword, {
    message: 'The new password must be different from the old password',
    path: ['newPassword']
  });

type ChangePasswordData = {
  oldPassword: string;
  newPassword: string;
  newPasswordConfirm: string;
};

export interface ChangePasswordFormProps {}

export function ChangePasswordForm({}: ChangePasswordFormProps) {
  const { data: session } = useSession();
  const addMessage = useStore((state) => state.addMessage);
  const [loading, setLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset
  } = useForm<ChangePasswordData>({
    resolver: zodResolver(UPDATE_PASSWORD_SCHEMA)
  });

  async function onSubmit(data: ChangePasswordData) {
    setLoading(true);

    try {
      await UsersService.v1UserUpdate(session!.user!.id, {
        password: data.oldPassword,
        new_password: data.newPassword
      });

      addMessage({ type: 'success', title: 'Password updated', message: `Your password has been updated.` });
      reset();
    } catch (e) {
      addMessage({ type: 'error', title: 'Failed to update password', message: getErrorMessage(e) });
    }

    setLoading(false);
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
          required
        />

        <FormInput
          type="password"
          name="newPassword"
          label="New password"
          placeholder="New password"
          register={register}
          errors={errors}
          required
        />

        <FormInput
          type="password"
          name="newPasswordConfirm"
          label="Confirm new password"
          placeholder="New password again"
          register={register}
          errors={errors}
          required
        />
      </div>

      <div className="pt-5 flex justify-end">
        <Button type={'submit'} variant="primary" loading={loading}>
          Save
        </Button>
      </div>
    </form>
  );
}
