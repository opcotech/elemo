import { getRandomString } from "./random";

export interface TodoTestData {
  title: string;
  description?: string;
  priority?: "normal" | "important" | "urgent" | "critical";
  dueDate?: Date;
}

export function generateTodoData(overrides?: Partial<TodoTestData>): TodoTestData {
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  
  return {
    title: `Test TODO ${getRandomString(8)}`,
    description: `Description for test TODO ${getRandomString(8)}`,
    priority: "normal",
    dueDate: tomorrow,
    ...overrides,
  };
}

export function generateTitleOnly(): TodoTestData {
  return {
    title: `TODO ${getRandomString(8)}`,
  };
}

export function generateWithDescription(): TodoTestData {
  return {
    title: `TODO ${getRandomString(8)}`,
    description: `Description ${getRandomString(12)}`,
  };
}

export function generateWithPriority(priority: "normal" | "important" | "urgent" | "critical"): TodoTestData {
  return {
    title: `TODO ${getRandomString(8)}`,
    priority,
  };
}

export function generateComplete(): TodoTestData {
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  
  return {
    title: `TODO ${getRandomString(8)}`,
    description: `Description ${getRandomString(12)}`,
    priority: "important",
    dueDate: tomorrow,
  };
}