'use client';

import { z } from 'zod';
import { $User, getErrorMessage, Language, UsersService } from '@/lib/api';
import { FormSelect, FormSelectOption } from '@/components/blocks/Form/FormSelect';
import { LANGUAGES } from '@/lib/constants';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { Button } from '@/components/blocks/Button';
import { FormInput } from '@/components/blocks/Form/FormInput';
import { FormTextarea } from '@/components/blocks/Form/FormTextarea';
import useStore from '@/store';
import { normalizeData } from '@/lib/helpers/schema';
import { Avatar } from '@/components/blocks/Avatar';
import { getInitials } from '@/lib/helpers';

type UpdateUserProfileData = {
  username: string;
  first_name?: string;
  last_name?: string;
  picture?: string;
  title?: string;
  bio?: string;
  languages?: Array<Language>;
};

const UPDATE_PROFILE_SCHEMA = z.object({
  username: z
    .string()
    .min($User.properties.username.minLength, 'Username is too short')
    .max($User.properties.username.maxLength, 'Username is too long'),
  first_name: z
    .string()
    .min($User.properties.first_name.minLength, 'First name is too short')
    .max($User.properties.first_name.maxLength, 'First name is too long')
    .optional()
    .or(z.literal('')),
  last_name: z
    .string()
    .min($User.properties.last_name.minLength, 'Last name is too short')
    .max($User.properties.last_name.maxLength, 'Last name is too long')
    .optional()
    .or(z.literal('')),
  title: z
    .string()
    .min($User.properties.title.minLength, 'Title is too short')
    .max($User.properties.title.maxLength, 'Title is too long')
    .optional()
    .or(z.literal('')),
  picture: z.string().url().optional().or(z.literal('')),
  bio: z.string().max($User.properties.bio.maxLength, 'Bio is too long').optional().or(z.literal('')),
  languages: z.array(z.string()).optional()
});

const LANGUAGE_CHOICES: FormSelectOption[] = LANGUAGES.map((language) => {
  return {
    label: language.name,
    value: language.code
  };
});

export interface UpdateUserProfileFormProps {
  userId: string;
  defaultValues?: UpdateUserProfileData;
}

export function UpdateUserProfileForm({ userId, defaultValues }: UpdateUserProfileFormProps) {
  const addMessage = useStore((state) => state.addMessage);
  const [languageQuery, setLanguageQuery] = useState('');
  const [selectedLanguages, setSelectedLanguages] = useState<FormSelectOption[]>(
    LANGUAGE_CHOICES.filter((language) =>
      (defaultValues?.languages || []).map((code) => code.toLowerCase()).includes(language.value.toLowerCase())
    ).map((language) => language)
  );

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors, isSubmitting }
  } = useForm<UpdateUserProfileData>({
    defaultValues: {
      ...defaultValues,
      languages: selectedLanguages.map((language) => language.value)
    },
    resolver: zodResolver(UPDATE_PROFILE_SCHEMA)
  });

  const filteredLanguages =
    languageQuery === ''
      ? LANGUAGE_CHOICES
      : LANGUAGE_CHOICES.filter((language) => {
          return language.label.toLowerCase().includes(languageQuery.toLowerCase());
        });

  async function selectLanguages(languages: FormSelectOption[]) {
    console.log('languages', languages);
    setSelectedLanguages(languages);
    setValue(
      'languages',
      languages.map((language) => language.value)
    );
  }

  async function onSubmit(data: UpdateUserProfileData) {
    try {
      await UsersService.v1UserUpdate(userId, normalizeData(data, UPDATE_PROFILE_SCHEMA));
      addMessage({ type: 'success', title: 'Profile updated', message: 'Your profile has been updated successfully.' });
    } catch (e) {
      addMessage({ type: 'error', title: 'Failed to update profile', message: getErrorMessage(e) });
      reset();
    }
  }

  return (
    <form action={'#'} onSubmit={handleSubmit(onSubmit)}>
      <div className={'space-y-6'}>
        <div className={'sm:grid sm:grid-cols-12 sm:items-start sm:gap-3'}>
          <div className={'mt-1 sm:col-span-3 sm:mt-0'}>
            <Avatar
              size={'xl'}
              src={watch('picture') || ''}
              initials={getInitials(`${watch('first_name')} ${watch('last_name')}`)}
              className={'mt-2 mr-4'}
            />
          </div>
          <div className={'mt-1 sm:col-span-9 sm:mt-0'}>
            <FormInput
              type="url"
              name="picture"
              label="Picture"
              placeholder="https://example.com/static/images/avatar.png"
              grid={false}
              register={register}
              errors={errors}
              required={!UPDATE_PROFILE_SCHEMA.shape.picture.isOptional()}
            />
          </div>
        </div>
        <FormInput
          type="text"
          name="username"
          label="Username"
          placeholder="bob"
          prefix="@"
          register={register}
          errors={errors}
          required={!UPDATE_PROFILE_SCHEMA.shape.username.isOptional()}
        />
        <FormInput
          type="text"
          name="first_name"
          label="First name"
          placeholder="Bob"
          register={register}
          errors={errors}
          required={!UPDATE_PROFILE_SCHEMA.shape.first_name.isOptional()}
        />
        <FormInput
          type="text"
          name="last_name"
          label="Last name"
          placeholder="Awesome"
          register={register}
          errors={errors}
          required={!UPDATE_PROFILE_SCHEMA.shape.last_name.isOptional()}
        />
        <FormInput
          type="text"
          name="title"
          label="Title"
          placeholder="Senior Software Engineer"
          register={register}
          errors={errors}
          required={!UPDATE_PROFILE_SCHEMA.shape.title.isOptional()}
        />
        <FormTextarea
          name="bio"
          label="Bio"
          placeholder="A few words about me..."
          rows={4}
          register={register}
          errors={errors}
          required={!UPDATE_PROFILE_SCHEMA.shape.bio.isOptional()}
        />
        <FormSelect
          label="Languages"
          name="languages"
          register={register}
          placeholder="Select languages"
          options={filteredLanguages}
          selectedOptions={selectedLanguages}
          setFilter={setLanguageQuery}
          setSelectedOptions={selectLanguages}
          errors={errors}
          multiple
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
