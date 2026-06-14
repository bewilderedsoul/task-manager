export type TaskStatus = "todo" | "in_progress" | "done";
export type TaskPriority = "low" | "medium" | "high";
export type Role = "user" | "admin";

export interface User {
  id: string;
  email: string;
  role: Role;
  createdAt: string;
  updatedAt: string;
}

export interface Task {
  id: string;
  userId: string;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  dueDate: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface TaskListResult {
  tasks: Task[];
  total: number;
  page: number;
  pageSize: number;
}

export interface TaskActivity {
  id: string;
  taskId: string;
  userId: string;
  action: string;
  detail: Record<string, unknown>;
  createdAt: string;
}

export type SortField = "created_at" | "due_date" | "priority";
export type SortOrder = "asc" | "desc";

export interface TaskQuery {
  status?: TaskStatus | "";
  search?: string;
  sort?: SortField;
  order?: SortOrder;
  page?: number;
  pageSize?: number;
  scope?: "all";
}

export interface TaskInput {
  title: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  dueDate?: string | null;
}
