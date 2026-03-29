import { useEffect, useState } from "react";

const STORAGE_KEY = "homelabwatch-theme";
const DEFAULT_THEME = "dark";

function resolveInitialTheme() {
  if (typeof document !== "undefined") {
    const currentTheme = document.documentElement.dataset.theme;
    if (currentTheme === "dark" || currentTheme === "light") {
      return currentTheme;
    }
  }

  if (typeof window !== "undefined") {
    try {
      const storedTheme = window.localStorage.getItem(STORAGE_KEY);
      if (storedTheme === "dark" || storedTheme === "light") {
        return storedTheme;
      }
    } catch {
      // Ignore storage failures and fall back to the default theme.
    }
  }

  return DEFAULT_THEME;
}

export function useThemePreference() {
  const [theme, setTheme] = useState(resolveInitialTheme);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;

    try {
      window.localStorage.setItem(STORAGE_KEY, theme);
    } catch {
      // Ignore storage failures and keep the current in-memory selection.
    }
  }, [theme]);

  function toggleTheme() {
    setTheme((currentTheme) => (currentTheme === "dark" ? "light" : "dark"));
  }

  return { theme, toggleTheme };
}
