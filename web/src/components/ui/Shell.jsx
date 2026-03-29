import ThemeToggle from "./ThemeToggle";

export default function Shell({ children, onToggleTheme, theme }) {
  return (
    <main className="app-shell">
      <div className="app-shell__backdrop pointer-events-none fixed inset-0" />
      <div className="app-shell__grid pointer-events-none fixed inset-0" />
      <ThemeToggle onToggle={onToggleTheme} theme={theme} />
      <div className="relative mx-auto w-full max-w-[1600px]">{children}</div>
    </main>
  );
}
