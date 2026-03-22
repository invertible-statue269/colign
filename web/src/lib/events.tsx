"use client";

import { createContext, useContext, useEffect, useRef, useCallback } from "react";
import { notificationClient } from "./notification";

type EventHandler = (event: { type: string; changeId: bigint; payload: string }) => void;

interface EventContextValue {
  on: (handler: EventHandler) => () => void;
}

const EventContext = createContext<EventContextValue>({
  on: () => () => {},
});

export function EventProvider({ children }: { children: React.ReactNode }) {
  const handlersRef = useRef<Set<EventHandler>>(new Set());

  const on = useCallback((handler: EventHandler) => {
    handlersRef.current.add(handler);
    return () => {
      handlersRef.current.delete(handler);
    };
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function subscribe() {
      while (!cancelled) {
        try {
          if (!notificationClient?.subscribe) {
            await new Promise((r) => setTimeout(r, 5000));
            continue;
          }
          const stream = notificationClient.subscribe({ changeId: BigInt(0) });
          for await (const event of stream) {
            if (cancelled) break;
            for (const handler of handlersRef.current) {
              handler({
                type: event.type,
                changeId: event.changeId,
                payload: event.payload,
              });
            }
          }
        } catch {
          // Connection lost, retry after delay
          if (!cancelled) {
            await new Promise((r) => setTimeout(r, 3000));
          }
        }
      }
    }

    subscribe();
    return () => {
      cancelled = true;
    };
  }, []);

  return <EventContext.Provider value={{ on }}>{children}</EventContext.Provider>;
}

/**
 * Hook to listen for real-time events.
 * Returns a function to register a handler; call the returned cleanup function to unsubscribe.
 *
 * Usage:
 * ```
 * const { on } = useEvents();
 * useEffect(() => on((e) => { if (e.type === "task_created") refetch(); }), [on]);
 * ```
 */
export function useEvents() {
  return useContext(EventContext);
}
