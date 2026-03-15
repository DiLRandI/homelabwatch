import { cn } from "../../lib/cn";

const variants = {
  primary:
    "border border-accent bg-accent text-white shadow-card hover:bg-accent-strong hover:border-accent-strong focus-visible:ring-accent/25",
  secondary:
    "border border-slate-200 bg-white text-slate-700 hover:bg-slate-50 focus-visible:ring-slate-300",
  ghost:
    "border border-transparent bg-transparent text-slate-600 hover:bg-slate-100 hover:text-slate-900 focus-visible:ring-slate-300",
  subtle:
    "border border-slate-200 bg-slate-50 text-slate-700 hover:bg-slate-100 focus-visible:ring-slate-300",
};

const sizes = {
  sm: "h-9 px-3 text-sm",
  md: "h-10 px-4 text-sm",
  lg: "h-11 px-5 text-sm",
  icon: "h-10 w-10",
};

export default function Button({
  children,
  className,
  disabled = false,
  leadingIcon: LeadingIcon,
  size = "md",
  trailingIcon: TrailingIcon,
  type = "button",
  variant = "primary",
  ...props
}) {
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-xl font-medium transition focus-visible:outline-hidden focus-visible:ring-4 disabled:cursor-not-allowed disabled:opacity-60",
        variants[variant] || variants.primary,
        sizes[size] || sizes.md,
        className,
      )}
      disabled={disabled}
      type={type}
      {...props}
    >
      {LeadingIcon ? <LeadingIcon className="h-4 w-4" /> : null}
      {children}
      {TrailingIcon ? <TrailingIcon className="h-4 w-4" /> : null}
    </button>
  );
}
