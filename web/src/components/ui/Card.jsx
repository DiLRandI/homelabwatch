import { cn } from "../../lib/cn";

export function Card({ children, className }) {
  return (
    <section
      className={cn(
        "rounded-3xl border border-slate-200 bg-white shadow-card",
        className,
      )}
    >
      {children}
    </section>
  );
}

export function CardHeader({
  action,
  children,
  className,
  description,
  title,
}) {
  return (
    <div
      className={cn(
        "flex flex-col gap-4 border-b border-slate-200 px-5 py-5 sm:flex-row sm:items-start sm:justify-between sm:px-6",
        className,
      )}
    >
      <div className="min-w-0">
        {children || (
          <>
            <h2 className="text-lg font-semibold tracking-tight text-slate-950">
              {title}
            </h2>
            {description ? (
              <p className="mt-1 max-w-2xl text-sm leading-6 text-slate-500">
                {description}
              </p>
            ) : null}
          </>
        )}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}

export function CardContent({ children, className }) {
  return <div className={cn("px-5 py-5 sm:px-6", className)}>{children}</div>;
}

export function CardFooter({ children, className }) {
  return (
    <div
      className={cn(
        "border-t border-slate-200 px-5 py-4 sm:px-6",
        className,
      )}
    >
      {children}
    </div>
  );
}
