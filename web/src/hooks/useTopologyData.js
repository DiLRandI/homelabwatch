import { useRef, useState } from "react";

import { fetchTopology } from "../lib/api";

export function useTopologyData({ onError }) {
  const [topology, setTopology] = useState(null);
  const inFlightRef = useRef(null);
  const requestIdRef = useRef(0);

  async function loadTopology() {
    if (inFlightRef.current) {
      return inFlightRef.current;
    }

    const requestID = requestIdRef.current + 1;
    requestIdRef.current = requestID;

    const request = (async () => {
      try {
        onError?.("");
        const payload = await fetchTopology();
        if (requestIdRef.current === requestID) {
          setTopology(payload);
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

  return { loadTopology, topology };
}
