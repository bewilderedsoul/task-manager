"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { AuthForm } from "@/components/AuthForm";
import { useAuth } from "@/lib/auth-context";

export default function LoginPage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  // If already authenticated, skip the form.
  useEffect(() => {
    if (!loading && user) router.replace("/tasks");
  }, [loading, user, router]);

  return <AuthForm mode="login" />;
}
