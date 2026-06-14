"use client";

import type { Task } from "@/lib/types";
import {
  formatDate,
  isOverdue,
  priorityLabels,
  priorityStyles,
  statusLabels,
  statusStyles,
} from "@/lib/format";

interface Props {
  task: Task;
  showOwner?: boolean; // admin view across users
  onToggleComplete: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
}

export function TaskCard({
  task,
  showOwner,
  onToggleComplete,
  onEdit,
  onDelete,
}: Props) {
  const done = task.status === "done";
  const overdue = isOverdue(task.dueDate, task.status);

  return (
    <div className="flex items-start gap-3 rounded-xl border border-zinc-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md dark:border-zinc-800 dark:bg-zinc-900">
      <button
        onClick={() => onToggleComplete(task)}
        aria-label={done ? "Mark as not done" : "Mark as done"}
        className={`mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 transition-colors ${
          done
            ? "border-green-600 bg-green-600 text-white"
            : "border-zinc-300 hover:border-green-500 dark:border-zinc-600"
        }`}
      >
        {done && (
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
            <path d="M20 6 9 17l-5-5" />
          </svg>
        )}
      </button>

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <h3
            className={`font-medium ${done ? "text-zinc-400 line-through dark:text-zinc-500" : ""}`}
          >
            {task.title}
          </h3>
          <div className="flex shrink-0 gap-1">
            <button
              onClick={() => onEdit(task)}
              aria-label="Edit task"
              className="rounded-md p-1.5 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800 dark:hover:text-zinc-200"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 20h9" />
                <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z" />
              </svg>
            </button>
            <button
              onClick={() => onDelete(task)}
              aria-label="Delete task"
              className="rounded-md p-1.5 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/40"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
              </svg>
            </button>
          </div>
        </div>

        {task.description && (
          <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
            {task.description}
          </p>
        )}

        <div className="mt-3 flex flex-wrap items-center gap-2 text-xs">
          <span className={`rounded-full px-2 py-0.5 font-medium ${statusStyles[task.status]}`}>
            {statusLabels[task.status]}
          </span>
          <span className={`rounded-full px-2 py-0.5 font-medium ${priorityStyles[task.priority]}`}>
            {priorityLabels[task.priority]} priority
          </span>
          {task.dueDate && (
            <span
              className={`rounded-full px-2 py-0.5 font-medium ${
                overdue
                  ? "bg-red-100 text-red-700 dark:bg-red-950/50 dark:text-red-300"
                  : "bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300"
              }`}
            >
              {overdue ? "Overdue · " : "Due "}
              {formatDate(task.dueDate)}
            </span>
          )}
          {showOwner && (
            <span className="rounded-full bg-indigo-50 px-2 py-0.5 font-mono text-indigo-600 dark:bg-indigo-950/40 dark:text-indigo-300">
              {task.userId.slice(0, 8)}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
