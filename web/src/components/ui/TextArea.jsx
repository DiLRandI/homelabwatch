import { cn } from "../../lib/cn";

export default function TextArea({
  containerClassName,
  label,
  onChange,
  placeholder,
  rows = 4,
  value,
}) {
  return (
    <label className={cn("grid gap-2", containerClassName)}>
      <span className="block text-sm font-medium text-slate-700">
        {label}
      </span>
      <textarea
        className="min-h-28 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-950 shadow-sm outline-hidden transition placeholder:text-slate-400 focus:border-accent focus-visible:ring-4 focus-visible:ring-accent/15"
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        rows={rows}
        value={value}
      />
    </label>
  );
}
