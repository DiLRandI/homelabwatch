import { useRef, useState } from "react";

import { fetchDashboard } from "../lib/api";

export function useDashboardData({ onError }) {
  const [dashboard, setDashboard] = useState(null);
  const inFlightRef = useRef(null);
  const requestIdRef = useRef(0);

  async function loadDashboard() {
    if (inFlightRef.current) {
      return inFlightRef.current;
    }

    const requestID = requestIdRef.current + 1;
    requestIdRef.current = requestID;

    const request = (async () => {
      try {
        onError?.("");
        const payload = await fetchDashboard();
        if (requestIdRef.current === requestID) {
          setDashboard(payload);
        }
        return payload;
      } catch (requestError) {
        if (requestIdRef.current === requestID) {
          onError?.(requestError.message);
        }
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

  return {
    dashboard,
    loadDashboard,
  };
}
