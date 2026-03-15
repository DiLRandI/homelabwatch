export default function Section({ title, subtitle, children }) {
  return (
    <section className="animate-floatIn rounded-[2rem] border border-white/10 bg-panel/80 p-5 shadow-halo backdrop-blur">
      <div className="mb-5">
        <h2 className="font-display text-2xl font-semibold text-ink">{title}</h2>
        <p className="mt-1 text-sm text-muted">{subtitle}</p>
      </div>
      {children}
    </section>
  );
}
