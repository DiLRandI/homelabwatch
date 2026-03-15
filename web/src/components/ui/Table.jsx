import { cn } from "../../lib/cn";

export function Table({ children, className }) {
  return (
    <div className="overflow-x-auto">
      <table className={cn("min-w-full border-separate border-spacing-0", className)}>
        {children}
      </table>
    </div>
  );
}

export function TableHead({ children }) {
  return <thead>{children}</thead>;
}

export function TableBody({ children }) {
  return <tbody className="divide-y divide-slate-200">{children}</tbody>;
}

export function TableRow({ children, className }) {
  return (
    <tr className={cn("transition hover:bg-slate-50/80", className)}>
      {children}
    </tr>
  );
}

export function TableHeader({ children, className }) {
  return (
    <th
      className={cn(
        "border-b border-slate-200 px-4 py-3 text-left text-xs font-semibold uppercase tracking-[0.16em] text-slate-500",
        className,
      )}
      scope="col"
    >
      {children}
    </th>
  );
}

export function TableCell({ children, className }) {
  return (
    <td className={cn("px-4 py-4 align-top text-sm text-slate-600", className)}>
      {children}
    </td>
  );
}
