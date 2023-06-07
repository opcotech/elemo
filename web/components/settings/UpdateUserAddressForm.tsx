'use client';

import { z } from 'zod';
import { $User, getErrorMessage, UsersService } from '@/lib/api';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/blocks/Button';
import { FormInput } from '@/components/blocks/Form/FormInput';
import useStore from '@/store';

type UpdateUserAddressData = {
  address?: string;
};

const UPDATE_ADDRESS_SCHEMA = z.object({
  address: z.string().max($User.properties.address.maxLength, 'Address is too long').optional()
});

export interface UpdateUserAddressFormProps {
  userId: string;
  defaultValues?: UpdateUserAddressData;
}

export function UpdateUserAddressForm({ userId, defaultValues }: UpdateUserAddressFormProps) {
  const addMessage = useStore((state) => state.addMessage);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting }
  } = useForm({
    defaultValues,
    resolver: zodResolver(UPDATE_ADDRESS_SCHEMA)
  });

  async function onSubmit(data: UpdateUserAddressData) {
    try {
      await UsersService.v1UserUpdate(userId, data);
      addMessage({ type: 'success', title: 'Address updated', message: 'Your address has been updated successfully.' });
    } catch (e) {
      addMessage({ type: 'error', title: 'Failed to update address', message: getErrorMessage(e) });
      reset();
    }
  }

  return (
    <form action={'#'} onSubmit={handleSubmit(onSubmit)}>
      <div className={'space-y-6'}>
        <FormInput
          type="text"
          name="address"
          label="Address"
          placeholder="Remote"
          register={register}
          errors={errors}
          required={!UPDATE_ADDRESS_SCHEMA.shape.address.isOptional()}
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
