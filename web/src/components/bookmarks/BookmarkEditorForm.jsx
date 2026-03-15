import { useEffect, useState } from "react";

import Button from "../ui/Button";
import Input from "../ui/Input";
import TextArea from "../ui/TextArea";

function optionsFromServices(services) {
  return [...services].sort((left, right) => left.name.localeCompare(right.name));
}

export default function BookmarkEditorForm({
  bookmark = null,
  devices = [],
  folders = [],
  onSubmit,
  onUploadIcon,
  services = [],
}) {
  const [form, setForm] = useState({
    description: "",
    deviceId: "",
    folderId: "",
    iconMode: "auto",
    iconValue: "",
    isFavorite: false,
    monitorEnabled: false,
    name: "",
    serviceId: "",
    serviceVisible: false,
    tags: "",
    url: "",
    useDevicePrimaryAddress: false,
  });
  const serviceOptions = optionsFromServices(services);

  useEffect(() => {
    setForm({
      description: bookmark?.description || "",
      deviceId: bookmark?.deviceId || "",
      folderId: bookmark?.folderId || "",
      iconMode: bookmark?.iconMode || "auto",
      iconValue: bookmark?.iconValue || "",
      isFavorite: Boolean(bookmark?.isFavorite),
      monitorEnabled: false,
      name: bookmark?.manualName || bookmark?.name || "",
      serviceId: bookmark?.serviceId || "",
      serviceVisible: !bookmark?.serviceHidden,
      tags: (bookmark?.tags || []).join(", "),
      url: bookmark?.manualUrl || bookmark?.url || "",
      useDevicePrimaryAddress: Boolean(bookmark?.useDevicePrimaryAddress),
    });
  }, [bookmark]);

  async function handleIconUpload(event) {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    const asset = await onUploadIcon(file);
    if (!asset?.url) {
      return;
    }
    setForm((current) => ({
      ...current,
      iconMode: "uploaded",
      iconValue: asset.url,
    }));
  }

  async function handleSubmit(event) {
    event.preventDefault();
    const tags = form.tags
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);

    await onSubmit({
      id: bookmark?.id,
      description: form.description,
      deviceId: form.serviceId ? "" : form.deviceId,
      folderId: form.folderId,
      iconMode: form.iconMode,
      iconValue: form.iconMode === "auto" ? "" : form.iconValue,
      isFavorite: form.isFavorite,
      name: form.name,
      serviceId: form.serviceId,
      tags,
      url: form.serviceId ? "" : form.url,
      useDevicePrimaryAddress:
        !form.serviceId && form.useDevicePrimaryAddress && Boolean(form.deviceId),
      ...(form.monitorEnabled && !form.serviceId
        ? {
            monitor: {
              enabled: true,
              serviceName: form.name,
              serviceVisible: form.serviceVisible,
            },
          }
        : {}),
    });
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit}>
      <div className="grid gap-4 sm:grid-cols-2">
        <Input
          label="Display name"
          onChange={(value) => setForm((current) => ({ ...current, name: value }))}
          placeholder="Home Assistant"
          value={form.name}
        />
        <label className="grid gap-2 text-sm font-medium text-slate-700">
          Folder
          <select
            className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-accent focus:ring-4 focus:ring-accent/10"
            onChange={(event) =>
              setForm((current) => ({ ...current, folderId: event.target.value }))
            }
            value={form.folderId}
          >
            <option value="">Unfiled</option>
            {folders.map((folder) => (
              <option key={folder.id} value={folder.id}>
                {folder.name}
              </option>
            ))}
          </select>
        </label>
      </div>

      <label className="grid gap-2 text-sm font-medium text-slate-700">
        Link existing service
        <select
          className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-accent focus:ring-4 focus:ring-accent/10"
          onChange={(event) =>
            setForm((current) => ({
              ...current,
              serviceId: event.target.value,
            }))
          }
          value={form.serviceId}
        >
          <option value="">No linked service</option>
          {serviceOptions.map((service) => (
            <option key={service.id} value={service.id}>
              {service.name}
            </option>
          ))}
        </select>
      </label>

      {!form.serviceId ? (
        <>
          <Input
            autoComplete="url"
            label="Launch URL"
            onChange={(value) => setForm((current) => ({ ...current, url: value }))}
            placeholder="http://192.168.1.20:8123"
            value={form.url}
          />

          <div className="grid gap-4 sm:grid-cols-2">
            <label className="grid gap-2 text-sm font-medium text-slate-700">
              Attach device
              <select
                className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-accent focus:ring-4 focus:ring-accent/10"
                onChange={(event) =>
                  setForm((current) => ({ ...current, deviceId: event.target.value }))
                }
                value={form.deviceId}
              >
                <option value="">No device</option>
                {devices.map((device) => (
                  <option key={device.id} value={device.id}>
                    {device.displayName || device.hostname || device.identityKey}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
              <input
                checked={form.useDevicePrimaryAddress}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    useDevicePrimaryAddress: event.target.checked,
                  }))
                }
                type="checkbox"
              />
              Use the device's current primary IP
            </label>
          </div>

          <label className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
            <input
              checked={form.monitorEnabled}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  monitorEnabled: event.target.checked,
                }))
              }
              type="checkbox"
            />
            Create a monitored service behind this bookmark
          </label>

          {form.monitorEnabled ? (
            <label className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
              <input
                checked={form.serviceVisible}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    serviceVisible: event.target.checked,
                  }))
                }
                type="checkbox"
              />
              Show the created service in the Services inventory
            </label>
          ) : null}
        </>
      ) : null}

      <Input
        label="Tags"
        onChange={(value) => setForm((current) => ({ ...current, tags: value }))}
        placeholder="monitoring, media, infrastructure"
        value={form.tags}
      />

      <TextArea
        label="Description"
        onChange={(value) =>
          setForm((current) => ({ ...current, description: value }))
        }
        placeholder="Optional notes shown on the bookmark card."
        value={form.description}
      />

      <div className="grid gap-4 sm:grid-cols-[180px_minmax(0,1fr)]">
        <label className="grid gap-2 text-sm font-medium text-slate-700">
          Icon mode
          <select
            className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-accent focus:ring-4 focus:ring-accent/10"
            onChange={(event) =>
              setForm((current) => ({ ...current, iconMode: event.target.value }))
            }
            value={form.iconMode}
          >
            <option value="auto">Automatic</option>
            <option value="external">External URL</option>
            <option value="uploaded">Uploaded image</option>
          </select>
        </label>

        {form.iconMode === "external" ? (
          <Input
            label="Icon URL"
            onChange={(value) =>
              setForm((current) => ({ ...current, iconValue: value }))
            }
            placeholder="https://example.com/icon.png"
            value={form.iconValue}
          />
        ) : form.iconMode === "uploaded" ? (
          <label className="grid gap-2 text-sm font-medium text-slate-700">
            Upload icon
            <input
              className="w-full rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-3 text-sm text-slate-700"
              onChange={handleIconUpload}
              type="file"
            />
          </label>
        ) : (
          <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-500">
            HomelabWatch will use a known service badge when possible, then fall back to the target favicon.
          </div>
        )}
      </div>

      <label className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
        <input
          checked={form.isFavorite}
          onChange={(event) =>
            setForm((current) => ({ ...current, isFavorite: event.target.checked }))
          }
          type="checkbox"
        />
        Pin this bookmark to the favorites strip
      </label>

      <div className="flex justify-end">
        <Button type="submit">{bookmark ? "Save bookmark" : "Create bookmark"}</Button>
      </div>
    </form>
  );
}
