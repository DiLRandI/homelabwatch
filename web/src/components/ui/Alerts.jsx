export default function Alerts({ error, notice }) {
  if (!error && !notice) {
    return null;
  }

  return (
    <div className="mt-4 grid gap-2">
      {notice ? (
        <div className="rounded-2xl border border-ok/30 bg-ok/10 px-4 py-3 text-sm text-ok">
          {notice}
        </div>
      ) : null}
      {error ? (
        <div className="rounded-2xl border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger">
          {error}
        </div>
      ) : null}
    </div>
  );
}
