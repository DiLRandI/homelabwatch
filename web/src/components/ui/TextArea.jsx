export default function TextArea({ label, value, onChange, placeholder }) {
  return (
    <label className="rounded-3xl border border-white/10 bg-white/5 p-4">
      <span className="block text-xs uppercase tracking-[0.24em] text-muted">
        {label}
      </span>
      <textarea
        className="mt-2 min-h-24 w-full rounded-2xl border border-white/10 bg-base/80 px-4 py-3 text-sm text-ink outline-hidden placeholder:text-muted/60 focus:border-accent/60"
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        value={value}
      />
    </label>
  );
}
