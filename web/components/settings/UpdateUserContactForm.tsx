'use client';

import { z } from 'zod';
import { $User, getErrorMessage, UserService } from '@/lib/api';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { Button } from '@/components/blocks/Button';
import { FormInput } from '@/components/blocks/Form/FormInput';
import useStore from '@/store';
import { Badge } from '@/components/blocks/Badge';
import { normalizeData } from '@/lib/helpers/schema';

type UpdateUserContactData = {
  email: string;
  phone?: string;
  links?: Array<string>;
};

const UPDATE_CONTACT_SCHEMA = z.object({
  email: z.string().email('Invalid email address'),
  phone: z.string().max($User.properties.phone.maxLength, 'Phone number is too long').optional().or(z.literal('')),
  links: z.array(z.string().url('Invalid URL')).optional()
});

export interface UpdateUserContactFormProps {
  userId: string;
  defaultValues?: UpdateUserContactData;
}

export function UpdateUserContactForm({ userId, defaultValues }: UpdateUserContactFormProps) {
  const addMessage = useStore((state) => state.addMessage);
  const [link, setLink] = useState<string>('');
  const [links, setLinks] = useState<string[]>(defaultValues?.links || []);

  const {
    register,
    handleSubmit,
    setValue,
    reset,
    formState: { errors, isSubmitting }
  } = useForm({
    defaultValues,
    resolver: zodResolver(UPDATE_CONTACT_SCHEMA)
  });

  function handleAddLink(link: string) {
    if (!link) return;
    if (links.includes(link)) return;

    setLink('');

    const newLinks = [...links, link];
    setLinks(newLinks);
    setValue('links', newLinks);
  }

  function handleRemoveLink(link: string) {
    setLinks(links.filter((l) => l !== link));
    setValue(
      'links',
      links.filter((l) => l !== link)
    );
  }

  async function onSubmit(data: UpdateUserContactData) {
    try {
      await UserService.v1UserUpdate(userId, normalizeData(data, UPDATE_CONTACT_SCHEMA));
      addMessage({
        type: 'success',
        title: 'Contact info updated',
        message: 'Your contact information has been updated successfully.'
      });
    } catch (e) {
      addMessage({ type: 'error', title: 'Failed to update contact info', message: getErrorMessage(e) });
      reset();
    }
  }

  return (
    <form action={'#'} onSubmit={handleSubmit(onSubmit)}>
      <div className={'space-y-6'}>
        <FormInput
          type="email"
          name="email"
          label="Email"
          placeholder="username@example.com"
          register={register}
          errors={errors}
          required={!UPDATE_CONTACT_SCHEMA.shape.email.isOptional()}
        />
        <FormInput
          type="tel"
          name="phone"
          label="Phone number"
          placeholder="+1 555 555 5555"
          register={register}
          errors={errors}
          required={!UPDATE_CONTACT_SCHEMA.shape.phone.isOptional()}
        />
        <FormInput
          type="url"
          name="link"
          label="Links"
          placeholder="https://example.com"
          value={link}
          register={register}
          errors={errors}
          errorField={'links'}
          onChange={(e) => setLink(e.target.value)}
          addon={
            <Button variant={'link'} size={'sm'} onClick={() => handleAddLink(link)}>
              Add
            </Button>
          }
          addonPosition={'right'}
          addonClassName={'hover:bg-gray-50'}
        >
          {links && (
            <div className="flex mt-4 space-x-2">
              {links.map((link: string, i: number) => (
                <Badge key={i} title={link} className={'mb-2'} dismissible onDismiss={() => handleRemoveLink(link)} />
              ))}
            </div>
          )}
        </FormInput>
      </div>
      <div className="pt-5 flex justify-end">
        <Button type={'submit'} variant="primary" loading={isSubmitting}>
          Save
        </Button>
      </div>
    </form>
  );
}
