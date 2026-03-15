import { cn } from "../../lib/cn";

export default function Input({
  autoComplete,
  containerClassName,
  inputClassName,
  label,
  labelClassName,
  value,
  onChange,
  placeholder,
  type = "text",
  compact = false,
  ...props
}) {
  return (
    <label
      className={cn("grid gap-2", compact ? "gap-1.5" : "", containerClassName)}
    >
      <span
        className={cn(
          "block text-sm font-medium text-slate-700",
          compact ? "text-xs uppercase tracking-[0.18em] text-slate-500" : "",
          labelClassName,
        )}
      >
        {label}
      </span>
      <input
        autoComplete={autoComplete}
        className={cn(
          "w-full rounded-2xl border border-slate-200 bg-white px-4 text-sm text-slate-950 shadow-sm outline-hidden transition placeholder:text-slate-400 focus:border-accent focus-visible:ring-4 focus-visible:ring-accent/15",
          compact ? "py-2.5" : "py-3",
          inputClassName,
        )}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        type={type}
        value={value}
        {...props}
      />
    </label>
  );
}
