'use client';

import { z } from 'zod';
import { $User, getErrorMessage, UserService } from '@/lib/api';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/blocks/Button';
import { FormInput } from '@/components/blocks/Form/FormInput';
import useStore from '@/store';
import { FormSwitch } from '@/components/blocks/Form/FormSwitch';
import { useState } from 'react';
import { normalizeData } from '@/lib/helpers/schema';

type UpdateUserAddressData = {
  address?: string;
};

const UPDATE_ADDRESS_SCHEMA = z.object({
  address: z
    .string()
    .min($User.properties.address.minLength, 'Address is too short')
    .max($User.properties.address.maxLength, 'Address is too long')
    .optional()
    .or(z.literal(''))
});

export interface UpdateUserAddressFormProps {
  userId: string;
  defaultValues?: UpdateUserAddressData;
}

export function UpdateUserAddressForm({ userId, defaultValues }: UpdateUserAddressFormProps) {
  const addMessage = useStore((state) => state.addMessage);
  const [previousAddress, setPreviousAddress] = useState<string | undefined>(defaultValues?.address);
  const [isRemote, setIsRemote] = useState(defaultValues?.address === 'Remote');

  const {
    register,
    setValue,
    getValues,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting }
  } = useForm({
    defaultValues,
    resolver: zodResolver(UPDATE_ADDRESS_SCHEMA)
  });

  async function onIsRemoteChange() {
    setIsRemote(!isRemote);

    // If the previous value was remote, set the address to the previous address
    // Otherwise, set the address to remote.
    if (!isRemote) {
      setPreviousAddress(getValues('address'));
      setValue('address', 'Remote');
    } else {
      setValue('address', previousAddress);
    }
  }

  async function onSubmit(data: UpdateUserAddressData) {
    try {
      await UserService.v1UserUpdate(userId, normalizeData(data, UPDATE_ADDRESS_SCHEMA));
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
          disabled={isRemote}
          register={register}
          errors={errors}
          required={!UPDATE_ADDRESS_SCHEMA.shape.address.isOptional()}
        />
        <FormSwitch
          label="Working remotely"
          description={
            'If you are working remotely, your address won&apos;t be displayed on' +
            'your profile and your location will be set to Remote.'
          }
          checked={isRemote}
          onChange={onIsRemoteChange}
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
