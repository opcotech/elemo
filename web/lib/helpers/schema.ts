import { z } from 'zod';

export function normalizeData<T extends Record<string, any>>(data: T, schema: z.AnyZodObject): T {
  const normalizedData: Partial<T> = { ...data };

  function isEmpty(value: any) {
    return value === null || (value !== undefined && value.trim?.() === '');
  }

  for (const [key, value] of Object.entries(data)) {
    if (schema.shape[key as keyof T]?.isOptional() && isEmpty(value)) {
      normalizedData[key as keyof T] = null as T[keyof T];
    }
  }

  return normalizedData as T;
}
