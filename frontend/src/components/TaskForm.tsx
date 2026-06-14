"use client";

import { useState } from "react";
import type { Task, TaskInput, TaskPriority, TaskStatus } from "@/lib/types";
import { fromDateInputValue, toDateInputValue } from "@/lib/format";
import {
  Button,
  FieldError,
  Input,
  Label,
  Select,
  Textarea,
} from "./ui";

interface Props {
  task?: Task | null; // present => edit mode
  onClose: () => void;
  onSubmit: (input: TaskInput) => Promise<void>;
}

export function TaskForm({ task, onClose, onSubmit }: Props) {
  const editing = Boolean(task);
  const [title, setTitle] = useState(task?.title ?? "");
  const [description, setDescription] = useState(task?.description ?? "");
  const [status, setStatus] = useState<TaskStatus>(task?.status ?? "todo");
  const [priority, setPriority] = useState<TaskPriority>(
    task?.priority ?? "medium",
  );
  const [dueDate, setDueDate] = useState(toDateInputValue(task?.dueDate ?? null));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setFormError(null);

    const errs: Record<string, string> = {};
    const trimmed = title.trim();
    if (!trimmed) errs.title = "Title is required";
    else if (trimmed.length > 200) errs.title = "Title is too long (max 200)";
    if (description.length > 5000) errs.description = "Description is too long";
    if (Object.keys(errs).length) {
      setErrors(errs);
      return;
    }
    setErrors({});

    setSubmitting(true);
    try {
      await onSubmit({
        title: trimmed,
        description,
        status,
        priority,
        dueDate: fromDateInputValue(dueDate),
      });
      onClose();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "Could not save task");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-xl dark:bg-zinc-900"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold">
          {editing ? "Edit task" : "New task"}
        </h2>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4" noValidate>
          <div>
            <Label htmlFor="title">Title</Label>
            <Input
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="What needs doing?"
              autoFocus
            />
            <FieldError message={errors.title} />
          </div>

          <div>
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional details…"
            />
            <FieldError message={errors.description} />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="status">Status</Label>
              <Select
                id="status"
                value={status}
                onChange={(e) => setStatus(e.target.value as TaskStatus)}
              >
                <option value="todo">To do</option>
                <option value="in_progress">In progress</option>
                <option value="done">Done</option>
              </Select>
            </div>
            <div>
              <Label htmlFor="priority">Priority</Label>
              <Select
                id="priority"
                value={priority}
                onChange={(e) => setPriority(e.target.value as TaskPriority)}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </Select>
            </div>
          </div>

          <div>
            <Label htmlFor="dueDate">Due date</Label>
            <Input
              id="dueDate"
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
            />
          </div>

          {formError && (
            <div className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
              {formError}
            </div>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? "Saving…" : editing ? "Save changes" : "Create task"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
