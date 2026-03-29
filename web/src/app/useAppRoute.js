import { useEffect, useState } from "react";

import { getRoute, normalizePath } from "./routes";

function scrollAppToTop() {
  const scrollRoot = document.querySelector("[data-app-scroll-root]");
  if (scrollRoot instanceof HTMLElement) {
    scrollRoot.scrollTop = 0;
  }

  window.scrollTo(0, 0);
}

export function useAppRoute() {
  const [pathname, setPathname] = useState(() =>
    normalizePath(window.location.pathname),
  );

  useEffect(() => {
    function handlePopState() {
      setPathname(normalizePath(window.location.pathname));
      scrollAppToTop();
    }

    window.addEventListener("popstate", handlePopState);
    return () => {
      window.removeEventListener("popstate", handlePopState);
    };
  }, []);

  function navigate(path) {
    const nextPath = normalizePath(path);
    if (nextPath === pathname) {
      scrollAppToTop();
      return;
    }

    window.history.pushState({}, "", nextPath);
    setPathname(nextPath);
    scrollAppToTop();
  }

  return {
    navigate,
    pathname,
    route: getRoute(pathname),
  };
}
