import type { TaskPriority, TaskStatus } from "./types";

export const statusLabels: Record<TaskStatus, string> = {
  todo: "To do",
  in_progress: "In progress",
  done: "Done",
};

export const priorityLabels: Record<TaskPriority, string> = {
  low: "Low",
  medium: "Medium",
  high: "High",
};

export const statusStyles: Record<TaskStatus, string> = {
  todo: "bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300",
  in_progress: "bg-blue-100 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300",
  done: "bg-green-100 text-green-700 dark:bg-green-950/50 dark:text-green-300",
};

export const priorityStyles: Record<TaskPriority, string> = {
  low: "bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300",
  medium: "bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300",
  high: "bg-red-100 text-red-700 dark:bg-red-950/50 dark:text-red-300",
};

// Formats an ISO timestamp as a short human date, or "" when null.
export function formatDate(iso: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

// Converts an ISO timestamp to the value expected by <input type="date">.
export function toDateInputValue(iso: string | null): string {
  if (!iso) return "";
  return new Date(iso).toISOString().slice(0, 10);
}

// Converts a yyyy-mm-dd input value to an RFC3339 timestamp (UTC midnight).
export function fromDateInputValue(value: string): string | null {
  if (!value) return null;
  return new Date(`${value}T00:00:00Z`).toISOString();
}

export function isOverdue(iso: string | null, status: TaskStatus): boolean {
  if (!iso || status === "done") return false;
  return new Date(iso).getTime() < Date.now();
}
