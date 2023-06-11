import { z } from 'zod';
/*
export function normalizeData<T extends Record<string, any>>(data: T, schema: z.AnyZodObject): T {
  const normalizedData: Partial<T> = { ...data };

  for (const [key, value] of Object.entries(data)) {
    if (schema.shape[key as keyof T]?.isOptional() && !value) {
      normalizedData[key as keyof T] = undefined;
    }
  }

  return normalizedData as T;
}
*/

export function normalizeData<T extends Record<string, any>>(data: T, schema: z.AnyZodObject): T {
  const normalizedData: Partial<T> = { ...data };

  for (const [key, value] of Object.entries(data)) {
    if (schema.shape[key as keyof T]?.isOptional() && !value) {
      normalizedData[key as keyof T] = null as T[keyof T];
    }
  }

  return normalizedData as T;
}
