import { useRef, useState } from "react";

import { fetchTopologySources } from "../lib/api";

export function useTopologySourcesData({ onError }) {
  const [topologySources, setTopologySources] = useState([]);
  const inFlightRef = useRef(null);
  const requestIdRef = useRef(0);

  async function loadTopologySources() {
    if (inFlightRef.current) {
      return inFlightRef.current;
    }

    const requestID = requestIdRef.current + 1;
    requestIdRef.current = requestID;

    const request = (async () => {
      try {
        onError?.("");
        const payload = await fetchTopologySources();
        if (requestIdRef.current === requestID) {
          setTopologySources(payload ?? []);
        }
        return payload;
      } catch (requestError) {
        if (requestIdRef.current === requestID) {
          onError?.(requestError.message);
        }
        return [];
      } finally {
        if (inFlightRef.current === request) {
          inFlightRef.current = null;
        }
      }
    })();

    inFlightRef.current = request;
    return request;
  }

  return { loadTopologySources, topologySources };
}
