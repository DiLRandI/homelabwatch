import { HalfMoonIcon, SunIcon } from "./Icons";

export default function ThemeToggle({ onToggle, theme }) {
  const isDark = theme === "dark";
  const Icon = isDark ? HalfMoonIcon : SunIcon;
  const label = isDark
    ? "Dark mode active. Switch to light mode."
    : "Light mode active. Switch to dark mode.";

  return (
    <button
      aria-label={label}
      aria-pressed={isDark}
      className="theme-toggle"
      onClick={onToggle}
      title={label}
      type="button"
    >
      <Icon className="h-4 w-4" />
      <span className="hidden text-xs font-semibold uppercase tracking-[0.18em] sm:inline">
        {isDark ? "Dark" : "Light"}
      </span>
    </button>
  );
}
