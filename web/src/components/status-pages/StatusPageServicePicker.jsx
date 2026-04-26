import { useEffect, useState } from "react";

import Button from "../ui/Button";
import { Card } from "../ui/Card";
import Input from "../ui/Input";

export default function StatusPageServicePicker({ canManage, onSave, page, services = [] }) {
  const [selected, setSelected] = useState([]);

  useEffect(() => {
    setSelected(page?.services || []);
  }, [page?.id, page?.services?.length]);

  if (!page) {
    return null;
  }

  const selectedIDs = new Set(selected.map((item) => item.serviceId));
  const available = services.filter((service) => !selectedIDs.has(service.id));

  function move(index, delta) {
    const next = [...selected];
    const target = index + delta;
    if (target < 0 || target >= next.length) return;
    [next[index], next[target]] = [next[target], next[index]];
    setSelected(next);
  }

  function save() {
    void onSave(page.id, selected.map((item, index) => ({
      serviceId: item.serviceId,
      sortOrder: index,
      displayName: item.displayName || "",
    })));
  }

  return (
    <Card className="p-5">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-ink">Services</h2>
          <p className="text-sm text-muted">Public names only expose curated display data.</p>
        </div>
        <Button disabled={!canManage} onClick={save} size="sm">Save order</Button>
      </div>
      <div className="mt-5 grid gap-3">
        {selected.map((item, index) => (
          <div className="grid gap-3 rounded-2xl border border-line bg-panel-strong p-3 md:grid-cols-[minmax(0,1fr)_auto]" key={item.serviceId}>
            <Input
              compact
              disabled={!canManage}
              label={item.serviceName || "Service"}
              onChange={(displayName) => setSelected(selected.map((current) => current.serviceId === item.serviceId ? { ...current, displayName } : current))}
              placeholder={item.serviceName}
              value={item.displayName || ""}
            />
            <div className="flex items-end gap-2">
              <Button disabled={!canManage || index === 0} onClick={() => move(index, -1)} size="sm" variant="secondary">Up</Button>
              <Button disabled={!canManage || index === selected.length - 1} onClick={() => move(index, 1)} size="sm" variant="secondary">Down</Button>
              <Button disabled={!canManage} onClick={() => setSelected(selected.filter((current) => current.serviceId !== item.serviceId))} size="sm" variant="ghost">Remove</Button>
            </div>
          </div>
        ))}
        <select
          className="rounded-2xl border border-line bg-panel-strong px-4 py-3 text-sm text-ink"
          disabled={!canManage}
          onChange={(event) => {
            const service = services.find((item) => item.id === event.target.value);
            if (service) {
              setSelected([...selected, { serviceId: service.id, serviceName: service.name, status: service.status, displayName: "" }]);
            }
            event.target.value = "";
          }}
          value=""
        >
          <option value="">Add service</option>
          {available.map((service) => <option key={service.id} value={service.id}>{service.name}</option>)}
        </select>
      </div>
    </Card>
  );
}
