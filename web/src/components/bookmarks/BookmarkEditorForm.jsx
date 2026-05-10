import { useEffect, useState } from "react";

import Button from "../ui/Button";
import Input from "../ui/Input";
import TextArea from "../ui/TextArea";

function optionsFromServices(services) {
  return [...services].sort((left, right) => left.name.localeCompare(right.name));
}

function SelectField({ children, label, onChange, value }) {
  return (
    <label className="grid gap-2 text-sm font-medium text-ink-soft">
      {label}
      <select
        className="w-full rounded-lg border border-line bg-panel-strong px-4 py-3 text-sm text-ink shadow-sm outline-hidden transition focus:border-accent focus-visible:ring-4 focus-visible:ring-accent/15"
        onChange={(event) => onChange(event.target.value)}
        value={value}
      >
        {children}
      </select>
    </label>
  );
}

function FieldGroup({ children, description, title }) {
  return (
    <section className="grid gap-4 rounded-lg border border-line bg-panel/60 p-4">
      <div>
        <h3 className="text-sm font-semibold text-ink">{title}</h3>
        {description ? (
          <p className="mt-1 text-sm leading-6 text-muted">{description}</p>
        ) : null}
      </div>
      {children}
    </section>
  );
}

function CheckboxCard({ checked, children, disabled = false, onChange }) {
  return (
    <label
      className={`flex items-start gap-3 rounded-lg border px-4 py-3 text-sm transition ${
        checked
          ? "border-accent bg-accent/10 text-ink"
          : "border-line bg-panel-strong text-ink-soft hover:border-line-strong"
      } ${disabled ? "cursor-not-allowed opacity-60" : "cursor-pointer"}`}
    >
      <input
        checked={checked}
        className="mt-0.5"
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
        type="checkbox"
      />
      <span className="leading-6">{children}</span>
    </label>
  );
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

  const submitLabel = bookmark ? "Save bookmark" : "Create bookmark";

  return (
    <form className="grid gap-0" onSubmit={handleSubmit}>
      <div className="grid gap-4 pb-20">
        <FieldGroup
          description="Name the launcher and decide where it should live in the bookmark tree."
          title="Bookmark details"
        >
          <div className="grid gap-4 sm:grid-cols-2">
            <Input
              label="Display name"
              onChange={(value) => setForm((current) => ({ ...current, name: value }))}
              placeholder="Home Assistant"
              value={form.name}
            />
            <SelectField
              label="Folder"
              onChange={(value) => setForm((current) => ({ ...current, folderId: value }))}
              value={form.folderId}
            >
              <option value="">Unfiled</option>
              {folders.map((folder) => (
                <option key={folder.id} value={folder.id}>
                  {folder.name}
                </option>
              ))}
            </SelectField>
          </div>

          <SelectField
            label="Link existing service"
            onChange={(value) =>
              setForm((current) => ({
                ...current,
                serviceId: value,
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
          </SelectField>
        </FieldGroup>

        {!form.serviceId ? (
          <FieldGroup
            description="Use a direct URL, optionally tied to a known device or promoted into monitoring."
            title="Launch target"
          >
            <Input
              autoComplete="url"
              label="Launch URL"
              onChange={(value) => setForm((current) => ({ ...current, url: value }))}
              placeholder="http://192.168.1.20:8123"
              value={form.url}
            />

            <div className="grid gap-4 sm:grid-cols-2">
              <SelectField
                label="Attach device"
                onChange={(value) => setForm((current) => ({ ...current, deviceId: value }))}
                value={form.deviceId}
              >
                <option value="">No device</option>
                {devices.map((device) => (
                  <option key={device.id} value={device.id}>
                    {device.displayName || device.hostname || device.identityKey}
                  </option>
                ))}
              </SelectField>

              <div className="flex items-end">
                <CheckboxCard
                  checked={form.useDevicePrimaryAddress}
                  disabled={!form.deviceId}
                  onChange={(checked) =>
                    setForm((current) => ({
                      ...current,
                      useDevicePrimaryAddress: checked,
                    }))
                  }
                >
                  Use the device's current primary IP
                </CheckboxCard>
              </div>
            </div>

            <div className="grid gap-3">
              <CheckboxCard
                checked={form.monitorEnabled}
                onChange={(checked) =>
                  setForm((current) => ({
                    ...current,
                    monitorEnabled: checked,
                  }))
                }
              >
                Create a monitored service behind this bookmark
              </CheckboxCard>

              {form.monitorEnabled ? (
                <CheckboxCard
                  checked={form.serviceVisible}
                  onChange={(checked) =>
                    setForm((current) => ({
                      ...current,
                      serviceVisible: checked,
                    }))
                  }
                >
                  Show the created service in the Services inventory
                </CheckboxCard>
              ) : null}
            </div>
          </FieldGroup>
        ) : null}

        <FieldGroup
          description="These details control how the bookmark appears in the workspace."
          title="Card appearance"
        >
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
            rows={3}
            value={form.description}
          />

          <div className="grid gap-4 sm:grid-cols-[180px_minmax(0,1fr)]">
            <SelectField
              label="Icon mode"
              onChange={(value) => setForm((current) => ({ ...current, iconMode: value }))}
              value={form.iconMode}
            >
              <option value="auto">Automatic</option>
              <option value="external">External URL</option>
              <option value="uploaded">Uploaded image</option>
            </SelectField>

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
              <label className="grid gap-2 text-sm font-medium text-ink-soft">
                Upload icon
                <input
                  className="w-full rounded-lg border border-dashed border-line bg-panel-strong px-4 py-3 text-sm text-ink file:mr-3 file:rounded-lg file:border-0 file:bg-accent file:px-3 file:py-2 file:text-sm file:font-medium file:text-white"
                  onChange={handleIconUpload}
                  type="file"
                />
              </label>
            ) : (
              <div className="rounded-lg border border-line bg-panel-strong px-4 py-3 text-sm leading-6 text-muted">
                HomelabWatch will use a known service badge when possible, then fall back to the target favicon.
              </div>
            )}
          </div>

          <CheckboxCard
            checked={form.isFavorite}
            onChange={(checked) =>
              setForm((current) => ({ ...current, isFavorite: checked }))
            }
          >
            Pin this bookmark to the favorites strip
          </CheckboxCard>
        </FieldGroup>
      </div>

      <div className="sticky bottom-0 -mx-5 -mb-5 flex justify-end border-t border-line bg-panel-strong px-5 py-4 sm:-mx-6 sm:px-6">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
