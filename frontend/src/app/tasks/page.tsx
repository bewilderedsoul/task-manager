"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { useTaskStream } from "@/lib/use-task-stream";
import type { Task, TaskInput, TaskListResult } from "@/lib/types";
import { Navbar } from "@/components/Navbar";
import { Toolbar, type ToolbarState } from "@/components/Toolbar";
import { TaskCard } from "@/components/TaskCard";
import { TaskForm } from "@/components/TaskForm";
import { Pagination } from "@/components/Pagination";
import { Button, Spinner } from "@/components/ui";

const PAGE_SIZE = 10;

const initialToolbar: ToolbarState = {
  search: "",
  status: "",
  sort: "created_at",
  order: "desc",
  scope: "own",
};

export default function TasksPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();

  const [toolbar, setToolbar] = useState<ToolbarState>(initialToolbar);
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [page, setPage] = useState(1);

  const [result, setResult] = useState<TaskListResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Task | null>(null);

  // Redirect unauthenticated visitors to the login page.
  useEffect(() => {
    if (!authLoading && !user) router.replace("/login");
  }, [authLoading, user, router]);

  // Debounce the search box so we don't fire a request on every keystroke.
  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(toolbar.search), 350);
    return () => clearTimeout(t);
  }, [toolbar.search]);

  // Reset to the first page whenever a filter/search/sort/scope changes.
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, toolbar.status, toolbar.sort, toolbar.order, toolbar.scope]);

  const showToast = useCallback((msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3500);
  }, []);

  const load = useCallback(async () => {
    if (!user) return;
    setLoading(true);
    setError(null);
    try {
      const data = await api.listTasks({
        search: debouncedSearch || undefined,
        status: toolbar.status || undefined,
        sort: toolbar.sort,
        order: toolbar.order,
        page,
        pageSize: PAGE_SIZE,
        scope: toolbar.scope === "all" ? "all" : undefined,
      });
      setResult(data);
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Could not load tasks.",
      );
    } finally {
      setLoading(false);
    }
  }, [
    user,
    debouncedSearch,
    toolbar.status,
    toolbar.sort,
    toolbar.order,
    toolbar.scope,
    page,
  ]);

  useEffect(() => {
    load();
  }, [load]);

  // Real-time: refetch the current view when the server pushes a change.
  // Keep a stable ref to the latest loader so the stream effect doesn't reconnect.
  const loadRef = useRef(load);
  loadRef.current = load;
  useTaskStream(Boolean(user), () => loadRef.current());

  // ── mutations ─────────────────────────────────────────
  async function handleCreate(input: TaskInput) {
    await api.createTask(input);
    await load();
  }

  async function handleEdit(input: TaskInput) {
    if (!editing) return;
    await api.updateTask(editing.id, input);
    await load();
  }

  // Optimistic complete toggle: flip the card immediately, roll back on failure.
  async function toggleComplete(task: Task) {
    if (!result) return;
    const nextStatus: Task["status"] = task.status === "done" ? "todo" : "done";
    const optimistic = result.tasks.map((t) =>
      t.id === task.id ? { ...t, status: nextStatus } : t,
    );
    setResult({ ...result, tasks: optimistic });
    try {
      await api.updateTask(task.id, { status: nextStatus });
    } catch {
      setResult(result); // rollback
      showToast("Couldn't update the task. Reverted.");
    }
  }

  // Optimistic delete: remove immediately, restore on failure.
  async function deleteTask(task: Task) {
    if (!result) return;
    if (!confirm(`Delete "${task.title}"?`)) return;
    const snapshot = result;
    setResult({
      ...result,
      tasks: result.tasks.filter((t) => t.id !== task.id),
      total: result.total - 1,
    });
    try {
      await api.deleteTask(task.id);
    } catch {
      setResult(snapshot); // rollback
      showToast("Couldn't delete the task. Restored.");
    }
  }

  function openCreate() {
    setEditing(null);
    setFormOpen(true);
  }
  function openEdit(task: Task) {
    setEditing(task);
    setFormOpen(true);
  }

  if (authLoading || !user) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <Spinner className="h-8 w-8 text-indigo-600" />
      </main>
    );
  }

  const tasks = result?.tasks ?? [];
  const isAdminScope = toolbar.scope === "all" && user.role === "admin";

  return (
    <div className="min-h-screen">
      <Navbar />

      <main className="mx-auto max-w-5xl px-4 py-6">
        <div className="mb-5 flex items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">Your tasks</h1>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              {isAdminScope
                ? "Viewing tasks across all users."
                : "Stay on top of what matters."}
            </p>
          </div>
          <Button onClick={openCreate}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 5v14M5 12h14" />
            </svg>
            New task
          </Button>
        </div>

        <div className="mb-5">
          <Toolbar
            state={toolbar}
            isAdmin={user.role === "admin"}
            onChange={(patch) => setToolbar((s) => ({ ...s, ...patch }))}
          />
        </div>

        {/* States: error, loading, empty, list */}
        {error ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-8 text-center dark:border-red-900/50 dark:bg-red-950/30">
            <p className="text-red-600 dark:text-red-400">{error}</p>
            <Button variant="secondary" className="mt-4" onClick={load}>
              Try again
            </Button>
          </div>
        ) : loading && !result ? (
          <div className="flex justify-center py-20">
            <Spinner className="h-8 w-8 text-indigo-600" />
          </div>
        ) : tasks.length === 0 ? (
          <div className="rounded-xl border border-dashed border-zinc-300 bg-white p-12 text-center dark:border-zinc-700 dark:bg-zinc-900">
            <p className="text-zinc-500 dark:text-zinc-400">
              {debouncedSearch || toolbar.status
                ? "No tasks match your filters."
                : "No tasks yet. Create your first one!"}
            </p>
            {!debouncedSearch && !toolbar.status && (
              <Button className="mt-4" onClick={openCreate}>
                New task
              </Button>
            )}
          </div>
        ) : (
          <div className={`space-y-3 ${loading ? "opacity-60" : ""}`}>
            {tasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                showOwner={isAdminScope}
                onToggleComplete={toggleComplete}
                onEdit={openEdit}
                onDelete={deleteTask}
              />
            ))}
          </div>
        )}

        {result && result.total > 0 && (
          <div className="mt-6">
            <Pagination
              page={result.page}
              pageSize={result.pageSize}
              total={result.total}
              onPageChange={setPage}
            />
          </div>
        )}
      </main>

      {formOpen && (
        <TaskForm
          task={editing}
          onClose={() => setFormOpen(false)}
          onSubmit={editing ? handleEdit : handleCreate}
        />
      )}

      {toast && (
        <div className="fixed bottom-4 left-1/2 z-50 -translate-x-1/2 rounded-lg bg-zinc-900 px-4 py-2 text-sm text-white shadow-lg dark:bg-zinc-100 dark:text-zinc-900">
          {toast}
        </div>
      )}
    </div>
  );
}
