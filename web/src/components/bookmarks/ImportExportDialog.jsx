import { useEffect, useState } from "react";

import Button from "../ui/Button";
import TextArea from "../ui/TextArea";

export default function ImportExportDialog({
  exportData = null,
  onExportRefresh,
  onImport,
}) {
  const [importError, setImportError] = useState("");
  const [importValue, setImportValue] = useState("");
  const [exportValue, setExportValue] = useState("");

  useEffect(() => {
    setExportValue(exportData ? JSON.stringify(exportData, null, 2) : "");
  }, [exportData]);

  async function handleImport() {
    if (!importValue.trim()) {
      return;
    }
    try {
      setImportError("");
      const payload = JSON.parse(importValue);
      const successful = await onImport(payload);
      if (successful !== false) {
        setImportValue("");
      }
    } catch (error) {
      setImportError(error.message || "Import payload is invalid.");
    }
  }

  async function handleCopy() {
    if (!exportValue) {
      return;
    }
    await navigator.clipboard.writeText(exportValue);
  }

  return (
    <div className="grid gap-5">
      <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h4 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-500">
              Export
            </h4>
            <p className="mt-1 text-sm text-slate-600">
              Capture folders, tags, bookmarks, usage history, and uploaded icons.
            </p>
          </div>
          <div className="flex gap-2">
            <Button onClick={onExportRefresh} size="sm" variant="secondary">
              Refresh export
            </Button>
            <Button onClick={() => void handleCopy()} size="sm">
              Copy JSON
            </Button>
          </div>
        </div>
        <TextArea
          containerClassName="mt-4"
          label="Export payload"
          onChange={setExportValue}
          rows={10}
          value={exportValue}
        />
      </div>

      <div className="rounded-3xl border border-slate-200 bg-white p-4">
        <h4 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-500">
          Import
        </h4>
        <p className="mt-1 text-sm text-slate-600">
          Paste a previously exported bookmark backup and merge it into this dashboard.
        </p>
        <TextArea
          containerClassName="mt-4"
          label="Import payload"
          onChange={setImportValue}
          placeholder='{"folders":[],"items":[]}'
          rows={10}
          value={importValue}
        />
        {importError ? (
          <p className="mt-3 rounded-2xl border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
            {importError}
          </p>
        ) : null}
        <div className="mt-4 flex justify-end">
          <Button onClick={() => void handleImport()}>Import bookmarks</Button>
        </div>
      </div>
    </div>
  );
}
