import { useEffect, useRef } from "react";

const EVENT_TYPES = [
  "bootstrap",
  "service",
  "device",
  "check",
  "bookmark",
  "folder",
  "docker-endpoint",
  "scan-target",
  "discovered-service",
  "service-definition",
  "status-page",
  "status-page-announcement",
];

export function useServerEvents(enabled, handlers = {}, debounceMs = 200) {
  const handlersRef = useRef(handlers);
  const pendingTypesRef = useRef(new Set());
  const timerRef = useRef(null);

  useEffect(() => {
    handlersRef.current = handlers;
  }, [handlers]);

  useEffect(() => {
    if (!enabled) {
      return undefined;
    }

    const events = new EventSource("/api/ui/v1/events");
    const listeners = new Map();

    function flushPending() {
      timerRef.current = null;
      const pendingTypes = Array.from(pendingTypesRef.current);
      pendingTypesRef.current.clear();

      const uniqueHandlers = new Set();
      for (const type of pendingTypes) {
        const handler =
          handlersRef.current[type] || handlersRef.current["*"] || null;
        if (handler) {
          uniqueHandlers.add(handler);
        }
      }

      for (const handler of uniqueHandlers) {
        handler({ types: pendingTypes });
      }
    }

    function scheduleFlush(type) {
      pendingTypesRef.current.add(type);
      if (timerRef.current != null) {
        return;
      }
      timerRef.current = window.setTimeout(flushPending, debounceMs);
    }

    for (const type of EVENT_TYPES) {
      const listener = () => scheduleFlush(type);
      listeners.set(type, listener);
      events.addEventListener(type, listener);
    }

    return () => {
      if (timerRef.current != null) {
        window.clearTimeout(timerRef.current);
        timerRef.current = null;
      }
      pendingTypesRef.current.clear();
      for (const [type, listener] of listeners.entries()) {
        events.removeEventListener(type, listener);
      }
      events.close();
    };
  }, [debounceMs, enabled]);
}
