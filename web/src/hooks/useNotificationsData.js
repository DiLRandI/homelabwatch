import { useRef, useState } from "react";

import {
  fetchNotificationChannels,
  fetchNotificationDeliveries,
  fetchNotificationRules,
} from "../lib/api";

export function useNotificationsData({ onError }) {
  const [notifications, setNotifications] = useState({
    channels: [],
    deliveries: [],
    rules: [],
  });
  const inFlightRef = useRef(null);

  async function loadNotifications() {
    if (inFlightRef.current) {
      return inFlightRef.current;
    }
    const request = (async () => {
      try {
        onError?.("");
        const [channels, rules, deliveries] = await Promise.all([
          fetchNotificationChannels(),
          fetchNotificationRules(),
          fetchNotificationDeliveries(),
        ]);
        const payload = { channels, deliveries, rules };
        setNotifications(payload);
        return payload;
      } catch (requestError) {
        onError?.(requestError.message);
        return null;
      } finally {
        if (inFlightRef.current === request) {
          inFlightRef.current = null;
        }
      }
    })();
    inFlightRef.current = request;
    return request;
  }

  return { loadNotifications, notifications };
}
