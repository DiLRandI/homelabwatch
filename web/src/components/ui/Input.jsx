export default function Input({
  label,
  value,
  onChange,
  placeholder,
  type = "text",
  compact = false,
}) {
  return (
    <label
      className={`block ${compact ? "" : "rounded-3xl border border-white/10 bg-white/5 p-4"}`}
    >
      <span className="block text-xs uppercase tracking-[0.24em] text-muted">
        {label}
      </span>
      <input
        className="mt-2 w-full rounded-2xl border border-white/10 bg-base/80 px-4 py-3 text-sm text-ink outline-none ring-0 transition placeholder:text-muted/60 focus:border-accent/60"
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        type={type}
        value={value}
      />
    </label>
  );
}
