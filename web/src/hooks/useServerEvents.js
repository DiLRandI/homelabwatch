import { useEffect, useRef } from "react";

const EVENT_TYPES = [
  "service",
  "device",
  "check",
  "bookmark",
  "docker-endpoint",
  "scan-target",
];

export function useServerEvents(enabled, onRefresh) {
  const refreshRef = useRef(onRefresh);

  useEffect(() => {
    refreshRef.current = onRefresh;
  }, [onRefresh]);

  useEffect(() => {
    if (!enabled) {
      return undefined;
    }

    const events = new EventSource("/api/ui/v1/events");
    const handleRefresh = () => {
      refreshRef.current?.();
    };

    for (const type of EVENT_TYPES) {
      events.addEventListener(type, handleRefresh);
    }

    return () => {
      for (const type of EVENT_TYPES) {
        events.removeEventListener(type, handleRefresh);
      }
      events.close();
    };
  }, [enabled]);
}
