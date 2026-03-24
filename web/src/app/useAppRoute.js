import { useEffect, useState } from "react";

import { getRoute, normalizePath } from "./routes";

export function useAppRoute() {
  const [pathname, setPathname] = useState(() =>
    normalizePath(window.location.pathname),
  );

  useEffect(() => {
    function handlePopState() {
      setPathname(normalizePath(window.location.pathname));
    }

    window.addEventListener("popstate", handlePopState);
    return () => {
      window.removeEventListener("popstate", handlePopState);
    };
  }, []);

  function navigate(path) {
    const nextPath = normalizePath(path);
    if (nextPath === pathname) {
      window.scrollTo(0, 0);
      return;
    }

    window.history.pushState({}, "", nextPath);
    setPathname(nextPath);
    window.scrollTo(0, 0);
  }

  return {
    navigate,
    pathname,
    route: getRoute(pathname),
  };
}
