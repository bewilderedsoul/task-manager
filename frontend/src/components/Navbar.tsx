"use client";

import { useAuth } from "@/lib/auth-context";
import { Button } from "./ui";
import { ThemeToggle } from "./ThemeToggle";

export function Navbar() {
  const { user, logout } = useAuth();
  return (
    <header className="sticky top-0 z-10 border-b border-zinc-200 bg-white/80 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/80">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
        <div className="flex items-center gap-2">
          <span className="text-lg font-bold tracking-tight">✓ Tasks</span>
          {user?.role === "admin" && (
            <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-700 dark:bg-amber-900/40 dark:text-amber-300">
              admin
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          {user && (
            <>
              <span className="hidden text-sm text-zinc-500 sm:inline dark:text-zinc-400">
                {user.email}
              </span>
              <Button variant="ghost" onClick={logout}>
                Log out
              </Button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
