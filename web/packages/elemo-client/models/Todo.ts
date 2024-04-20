/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { TodoPriority } from './TodoPriority';
/**
 * A todo item belonging to a user.
 */
export type Todo = {
    /**
     * Unique identifier of the todo.
     */
    id: string;
    /**
     * Title of the todo item.
     */
    title: string;
    /**
     * Description of the todo item.
     */
    description: string;
    priority: TodoPriority;
    /**
     * Status of the todo item.
     */
    completed: boolean;
    /**
     * ID of the user who owns the todo item.
     */
    owned_by: string;
    /**
     * ID of the user who created the todo item.
     */
    created_by: string;
    /**
     * Completion due date of the todo item.
     */
    due_date: string | null;
    /**
     * Date when the todo item was created.
     */
    created_at: string;
    /**
     * Date when the todo item was updated.
     */
    updated_at: string | null;
};

