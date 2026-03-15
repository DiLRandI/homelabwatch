export default function Shell({ children }) {
  return (
    <main className="min-h-screen bg-base px-4 py-8 text-ink sm:px-6 lg:px-10">
      <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(242,196,90,0.14),_transparent_35%),radial-gradient(circle_at_bottom_right,_rgba(82,212,155,0.12),_transparent_32%),linear-gradient(180deg,_rgba(255,255,255,0.03),_transparent_45%)]" />
      <div className="pointer-events-none fixed inset-0 opacity-30 [background-image:linear-gradient(rgba(255,255,255,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.03)_1px,transparent_1px)] [background-size:72px_72px]" />
      <div className="relative mx-auto max-w-7xl">{children}</div>
    </main>
  );
}
