import { cn } from "../../lib/cn";

const variants = {
  neutral: "border-line bg-base text-muted",
  accent: "border-accent/15 bg-accent/10 text-accent-strong",
  success: "border-ok/15 bg-ok/10 text-ok-strong",
  warning: "border-warn/20 bg-warn/10 text-warn-strong",
  danger: "border-danger/15 bg-danger/10 text-danger-strong",
  info: "border-sky-500/25 bg-sky-500/12 text-sky-300",
};

export default function Badge({
  children,
  className,
  tone = "neutral",
  withDot = false,
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]",
        variants[tone] || variants.neutral,
        className,
      )}
    >
      {withDot ? <span className="h-1.5 w-1.5 rounded-full bg-current" /> : null}
      {children}
    </span>
  );
}
