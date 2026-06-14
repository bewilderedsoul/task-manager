"use client";

import type { SortField, SortOrder, TaskStatus } from "@/lib/types";
import { Select } from "./ui";

export interface ToolbarState {
  search: string;
  status: TaskStatus | "";
  sort: SortField;
  order: SortOrder;
  scope: "own" | "all";
}

interface Props {
  state: ToolbarState;
  isAdmin: boolean;
  onChange: (patch: Partial<ToolbarState>) => void;
}

export function Toolbar({ state, isAdmin, onChange }: Props) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
      <div className="relative flex-1 sm:min-w-[200px]">
        <svg
          className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400"
          width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
        <input
          type="search"
          value={state.search}
          onChange={(e) => onChange({ search: e.target.value })}
          placeholder="Search by title…"
          className="w-full rounded-lg border border-zinc-300 bg-white py-2 pl-9 pr-3 text-sm text-zinc-900 placeholder-zinc-400 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-100"
        />
      </div>

      <Select
        aria-label="Filter by status"
        className="sm:w-40"
        value={state.status}
        onChange={(e) => onChange({ status: e.target.value as TaskStatus | "" })}
      >
        <option value="">All statuses</option>
        <option value="todo">To do</option>
        <option value="in_progress">In progress</option>
        <option value="done">Done</option>
      </Select>

      <Select
        aria-label="Sort by"
        className="sm:w-40"
        value={state.sort}
        onChange={(e) => onChange({ sort: e.target.value as SortField })}
      >
        <option value="created_at">Created date</option>
        <option value="due_date">Due date</option>
        <option value="priority">Priority</option>
      </Select>

      <Select
        aria-label="Sort order"
        className="sm:w-32"
        value={state.order}
        onChange={(e) => onChange({ order: e.target.value as SortOrder })}
      >
        <option value="desc">Descending</option>
        <option value="asc">Ascending</option>
      </Select>

      {isAdmin && (
        <Select
          aria-label="Scope"
          className="sm:w-40"
          value={state.scope}
          onChange={(e) =>
            onChange({ scope: e.target.value as "own" | "all" })
          }
        >
          <option value="own">My tasks</option>
          <option value="all">All users (admin)</option>
        </Select>
      )}
    </div>
  );
}
