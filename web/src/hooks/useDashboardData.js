import { useState } from "react";

import { fetchDashboard } from "../lib/api";

export function useDashboardData({ onError }) {
  const [dashboard, setDashboard] = useState(null);

  async function loadDashboard() {
    try {
      onError?.("");
      const payload = await fetchDashboard();
      setDashboard(payload);
      return payload;
    } catch (requestError) {
      onError?.(requestError.message);
      return null;
    }
  }

  return {
    dashboard,
    loadDashboard,
  };
}
