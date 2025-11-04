import { z } from "zod";

/**
 * Creates a form schema that allows empty strings for optional fields.
 * Empty strings are transformed to undefined before validation.
 *
 * This is useful for form inputs where users can clear optional fields.
 *
 * @param schema - The base Zod schema
 * @returns A new schema with empty string handling for optional fields
 */
export function createFormSchema<T extends z.ZodObject<any>>(schema: T): T {
  const shape: Record<string, z.ZodTypeAny> = {};

  for (const [key, fieldSchema] of Object.entries(schema.shape)) {
    const zodField = fieldSchema as z.ZodTypeAny;

    // If the field is optional, allow empty strings and transform them to undefined
    if (zodField.isOptional()) {
      // Get the inner type (unwrap the optional wrapper)
      // Access Zod's internal structure to unwrap optional types
      const def = zodField.def as any;
      const innerType = def.innerType || def.schema;

      if (innerType) {
        // Handle union types (e.g., z.union([z.string().url(), z.null()]))
        const innerDef = innerType.def as any;
        if (innerDef.typeName === "ZodUnion") {
          const options = innerDef.options as z.ZodTypeAny[];
          // Find the non-null option
          const nonNullOption = options.find(
            (opt: z.ZodTypeAny) => (opt.def as any).typeName !== "ZodNull"
          );

          if (nonNullOption) {
            // Create a union that includes empty string literal, then transform to undefined
            shape[key] = z
              .union([nonNullOption, z.literal(""), z.null()])
              .optional()
              .transform((val) => {
                if (val === "" || val === null) return undefined;
                return val;
              });
          } else {
            shape[key] = zodField;
          }
        } else {
          // Handle simple optional types (e.g., z.optional(z.string().url()))
          // Create a union that allows empty string, then transform to undefined
          shape[key] = z
            .union([innerType, z.literal("")])
            .optional()
            .transform((val) => {
              if (val === "") return undefined;
              return val;
            });
        }
      } else {
        shape[key] = zodField;
      }
    } else {
      // For required fields, keep as-is
      shape[key] = zodField;
    }
  }

  return z.object(shape) as T;
}

/**
 * Normalizes form data for API submission by:
 * 1. Converting empty strings to undefined for optional fields
 * 2. Only including fields that are explicitly set (not undefined)
 * 3. Removing null values
 *
 * @param schema - The Zod schema to check which fields are optional
 * @param data - The data to normalize
 * @returns The normalized data ready for API submission
 */
export function normalizeFormData<T extends Record<string, any>>(
  schema: z.ZodObject<any>,
  data: T
): Partial<T> {
  const normalizedData: Partial<T> = {};

  function isEmpty(value: any) {
    return (
      value === null ||
      value === undefined ||
      (typeof value === "string" && value.trim() === "")
    );
  }

  for (const [key, value] of Object.entries(data)) {
    // Skip undefined values (field not set)
    if (value === undefined) {
      continue;
    }

    // For optional fields, convert empty strings to undefined
    const fieldSchema = schema.shape[key as keyof T];
    if (fieldSchema?.isOptional() && isEmpty(value)) {
      // Skip empty optional fields entirely
      continue;
    }

    // Include the field if it has a value
    normalizedData[key as keyof T] = value;
  }

  return normalizedData;
}

/**
 * Normalizes form data for PATCH operations.
 * For patch operations, if an optional field was cleared (changed from a value to empty),
 * it should be included as `null` to explicitly clear it on the backend.
 *
 * @param schema - The Zod schema to check which fields are optional
 * @param data - The current form data (may have undefined for cleared fields due to schema transform)
 * @param originalData - The original data before edits (for comparison)
 * @returns The normalized data ready for API submission
 */
export function normalizePatchData<T extends Record<string, any>>(
  schema: z.ZodObject<any>,
  data: T,
  originalData: Partial<T>
): Partial<T> {
  const normalizedData: Partial<T> = {};

  function isEmpty(value: any) {
    return (
      value === null ||
      value === undefined ||
      (typeof value === "string" && value.trim() === "")
    );
  }

  // Check all fields that exist in either form data or original data
  const allKeys = new Set([...Object.keys(data), ...Object.keys(originalData)]);

  for (const key of allKeys) {
    const value = data[key as keyof T];
    const originalValue = originalData[key as keyof T];
    const fieldSchema = schema.shape[key];

    if (!fieldSchema) {
      continue;
    }

    if (fieldSchema.isOptional()) {
      if (
        isEmpty(value) &&
        originalValue !== undefined &&
        originalValue !== null &&
        !isEmpty(originalValue)
      ) {
        normalizedData[key as keyof T] = null as any;
      } else if (value !== undefined && !isEmpty(value)) {
        normalizedData[key as keyof T] = value;
      }
    } else {
      if (value !== undefined && !isEmpty(value)) {
        normalizedData[key as keyof T] = value;
      }
    }
  }

  return normalizedData;
}

/**
 * Helper to create onChange handler for optional text fields.
 * Converts empty strings to undefined to properly handle optional fields.
 *
 * @param field - The form field from react-hook-form
 * @returns onChange handler function
 */
export function createOptionalFieldHandler<
  T extends string | undefined,
>(field: { onChange: (value: T) => void; onBlur: () => void }) {
  return (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const value = e.target.value;
    field.onChange((value === "" ? undefined : value) as T);
  };
}

/**
 * Returns the value of the field if not empty, otherwise returns the default value.
 *
 * @param value - The value of the field
 * @param defaultValue - The default value of the field
 * @returns The value of the field if not empty, otherwise returns the default value
 */
export function getFieldValue<T extends string | undefined | null>(
  value: T,
  defaultValue: string = ""
): string {
  return value || defaultValue;
}
