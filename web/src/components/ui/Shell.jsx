export default function Shell({ children }) {
  return (
    <main className="min-h-screen text-ink">
      <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(circle_at_top_left,rgba(37,99,235,0.12),transparent_32%),radial-gradient(circle_at_right,rgba(14,165,233,0.08),transparent_24%),linear-gradient(180deg,rgba(255,255,255,0.45),transparent_42%)]" />
      <div className="pointer-events-none fixed inset-0 opacity-40 bg-[linear-gradient(rgba(148,163,184,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.08)_1px,transparent_1px)] bg-size-[72px_72px]" />
      <div className="relative mx-auto w-full max-w-[1600px]">{children}</div>
    </main>
  );
}
