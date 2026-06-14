"use client";

import { useEffect, useRef } from "react";
import { api, tokenStore } from "./api";

// useTaskStream opens a Server-Sent Events connection to the backend and calls
// `onEvent` whenever a task change is pushed. We use fetch streaming rather than
// the native EventSource because EventSource cannot send an Authorization header.
//
// Real-time is a progressive enhancement: if the stream fails for any reason the
// app keeps working through normal request/response, so errors are swallowed.
export function useTaskStream(enabled: boolean, onEvent: () => void) {
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  useEffect(() => {
    if (!enabled) return;
    const token = tokenStore.get();
    if (!token) return;

    const controller = new AbortController();

    (async () => {
      try {
        const res = await fetch(`${api.baseUrl}/api/tasks/stream`, {
          headers: { Authorization: `Bearer ${token}` },
          signal: controller.signal,
        });
        if (!res.body) return;

        const reader = res.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;
          buffer += decoder.decode(value, { stream: true });

          // SSE frames are separated by a blank line.
          const frames = buffer.split("\n\n");
          buffer = frames.pop() ?? "";
          for (const frame of frames) {
            const line = frame
              .split("\n")
              .find((l) => l.startsWith("data:"));
            if (line) onEventRef.current();
          }
        }
      } catch {
        // Aborted on unmount or network hiccup — ignore.
      }
    })();

    return () => controller.abort();
  }, [enabled]);
}
