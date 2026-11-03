import { useEffect, useRef, useState, type FormEvent } from "react";
import { ApiError, photoSrc, request, type Customer, type Device, type Page } from "./api";
import type { Copy } from "./i18n";
import { usePaused } from "./list";
import { fileProblem, maxPhotos, problemText } from "./problems";

type LocalFile = { key: string; file: File; url: string };
type ExistingPhoto = { id: string };
type PartDraft = {
  key: string;
  id: string;
  name: string;
  note: string;
  existing: ExistingPhoto[];
  files: LocalFile[];
};

type Props = {
  copy: Copy;
  mode: "create" | "edit";
  deviceId?: string;
  customerId?: string | null;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function DeviceForm({ copy, mode, deviceId, customerId, onNavigate, onUnauthorized }: Props) {
  const editing = mode === "edit";
  const urls = useRef(new Set<string>());
  const [ready, setReady] = useState(!editing);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<Customer | null>(null);
  const [customerDraft, setCustomerDraft] = useState("");
  const [customerQuery] = usePaused(customerDraft);
  const [results, setResults] = useState<Customer[]>([]);
  const [resultsMore, setResultsMore] = useState(false);
  const [customerName, setCustomerName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [brand, setBrand] = useState("");
  const [model, setModel] = useState("");
  const [serial, setSerial] = useState("");
  const [note, setNote] = useState("");
  const [deviceExisting, setDeviceExisting] = useState<ExistingPhoto[]>([]);
  const [deviceFiles, setDeviceFiles] = useState<LocalFile[]>([]);
  const [parts, setParts] = useState<PartDraft[]>([]);

  useEffect(() => {
    const owned = urls.current;
    return () => owned.forEach((url) => URL.revokeObjectURL(url));
  }, []);

  useEffect(() => {
    if (!editing || !deviceId) {
      return;
    }
    const controller = new AbortController();
    request<Device>(`/api/devices/${deviceId}`, { signal: controller.signal })
      .then((device) => {
        setName(device.name);
        setBrand(device.brand);
        setModel(device.model);
        setSerial(device.serial);
        setNote(device.note);
        setCustomerName(device.customer.name);
        setPhone(device.customer.phone);
        setEmail(device.customer.email);
        setDeviceExisting(device.photos.map((photo) => ({ id: photo.id })));
        setParts(
          device.parts.map((part) => ({
            key: part.id,
            id: part.id,
            name: part.name,
            note: part.note,
            existing: part.photos.map((photo) => ({ id: photo.id })),
            files: [],
          })),
        );
        setReady(true);
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        if (err instanceof ApiError && err.code === "unauthorized") {
          onUnauthorized();
          return;
        }
        setError(problemText(copy, err));
      });
    return () => controller.abort();
  }, [copy, deviceId, editing, onUnauthorized]);

  useEffect(() => {
    if (editing || !customerId) {
      return;
    }
    const controller = new AbortController();
    request<Customer>(`/api/customers/${customerId}`, { signal: controller.signal })
      .then(setSelected)
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        if (err instanceof ApiError && err.code === "unauthorized") {
          onUnauthorized();
          return;
        }
        setError(problemText(copy, err));
      });
    return () => controller.abort();
  }, [copy, customerId, editing, onUnauthorized]);

  useEffect(() => {
    if (editing || selected || !customerQuery) {
      return;
    }
    const controller = new AbortController();
    const params = new URLSearchParams({ q: customerQuery, limit: "20" });
    request<Page<Customer>>(`/api/customers?${params}`, { signal: controller.signal })
      .then((page) => {
        setResults(page.items);
        setResultsMore(page.hasMore);
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        if (err instanceof ApiError && err.code === "unauthorized") {
          onUnauthorized();
        }
      });
    return () => controller.abort();
  }, [customerQuery, editing, onUnauthorized, selected]);

  function takeFiles(list: FileList | null) {
    return Array.from(list ?? []).map((file) => {
      const url = URL.createObjectURL(file);
      urls.current.add(url);
      return { key: crypto.randomUUID(), file, url };
    });
  }

  function dropFile(file: LocalFile) {
    URL.revokeObjectURL(file.url);
    urls.current.delete(file.url);
  }

  function photoCount() {
    return (
      deviceExisting.length +
      deviceFiles.length +
      parts.reduce((sum, part) => sum + part.existing.length + part.files.length, 0)
    );
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    const missing: string[] = [];
    const invalid: string[] = [];
    if (!editing && !selected && !customerName.trim()) {
      missing.push("customerName");
    }
    if (editing && !customerName.trim()) {
      missing.push("customerName");
    }
    if (!name.trim()) {
      missing.push("name");
    }
    if (!brand.trim()) {
      missing.push("brand");
    }
    if (!model.trim()) {
      missing.push("model");
    }
    if (!note.trim()) {
      missing.push("note");
    }
    if (email.trim() && !emailOk(email)) {
      invalid.push("email");
    }
    parts.forEach((part, index) => {
      if (!part.name.trim()) {
        missing.push(`parts.${index}.name`);
      }
      if (!part.note.trim()) {
        missing.push(`parts.${index}.note`);
      }
    });
    const locals = [...deviceFiles, ...parts.flatMap((part) => part.files)];
    let photo = "";
    for (const item of locals) {
      const problem = fileProblem(item.file);
      if (problem) {
        photo = problem;
        break;
      }
    }
    if (photoCount() > maxPhotos) {
      photo = "limit";
    }
    if (missing.length || invalid.length || photo) {
      setError(problemText(copy, new ApiError("validation", { missing, invalid, photo })));
      return;
    }

    const files: File[] = [];
    const indexes = (list: LocalFile[]) =>
      list.map((item) => {
        const index = files.length;
        files.push(item.file);
        return index;
      });
    const payload = editing
      ? {
          name,
          brand,
          model,
          serial,
          note,
          customerName,
          customerPhone: phone,
          customerEmail: email,
          deviceKeep: deviceExisting.map((photo) => photo.id),
          devicePhotos: indexes(deviceFiles),
          parts: parts.map((part) => ({
            id: part.id,
            name: part.name,
            note: part.note,
            keep: part.existing.map((photo) => photo.id),
            photos: indexes(part.files),
          })),
        }
      : {
          ...(selected ? { customerId: selected.id } : { customerName, customerPhone: phone, customerEmail: email }),
          name,
          brand,
          model,
          serial,
          note,
          devicePhotos: indexes(deviceFiles),
          parts: parts.map((part) => ({
            name: part.name,
            note: part.note,
            photos: indexes(part.files),
          })),
        };
    const body = new FormData();
    body.set("payload", JSON.stringify(payload));
    files.forEach((file, index) => body.append(`file${index}`, file));

    setSaving(true);
    setError(null);
    try {
      const saved = await request<Device>(editing ? `/api/devices/${deviceId}` : "/api/devices", {
        method: editing ? "PUT" : "POST",
        body,
      });
      onNavigate(`/devices/${saved.id}`);
    } catch (err) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setError(problemText(copy, err));
      setSaving(false);
    }
  }

  if (!ready) {
    return <p className="muted">{error ?? copy.loading}</p>;
  }

  return (
    <form className="stack" onSubmit={(event) => void save(event)}>
      <header className="page-head row">
        <div>
          <h1>{editing ? copy.correct : copy.newDevice}</h1>
        </div>
        <button
          className="button secondary"
          type="button"
          onClick={() => onNavigate(editing && deviceId ? `/devices/${deviceId}` : "/devices")}
        >
          {copy.back}
        </button>
      </header>

      <section className="panel stack">
        <h2>{copy.customer}</h2>
        {editing || selected ? (
          <>
            {selected ? (
              <div className="stack">
                <p className="muted">{copy.selectedCustomer}</p>
                <strong>{selected.name}</strong>
                <span>{selected.phone}</span>
                <span>{selected.email}</span>
                <button className="button secondary" type="button" onClick={() => setSelected(null)}>
                  {copy.clearCustomer}
                </button>
              </div>
            ) : (
              <div className="split">
                <label className="field">
                  {copy.customerName}
                  <input value={customerName} onChange={(event) => setCustomerName(event.target.value)} />
                </label>
                <label className="field">
                  {copy.phone}
                  <input value={phone} onChange={(event) => setPhone(event.target.value)} />
                </label>
                <label className="field">
                  {copy.email}
                  <input value={email} onChange={(event) => setEmail(event.target.value)} />
                </label>
              </div>
            )}
          </>
        ) : (
          <>
            <label className="field">
              {copy.search}
              <input
                value={customerDraft}
                placeholder={copy.searchCustomers}
                onChange={(event) => setCustomerDraft(event.target.value)}
              />
            </label>
            {results.length > 0 ? (
              <div className="choices">
                {results.map((customer) => (
                  <button
                    className="button secondary"
                    type="button"
                    key={customer.id}
                    onClick={() => {
                      setSelected(customer);
                      setResults([]);
                      setCustomerDraft("");
                    }}
                  >
                    <strong>{customer.name}</strong>
                    <span className="muted">
                      {[customer.phone, customer.email].filter(Boolean).join(" · ")}
                    </span>
                  </button>
                ))}
              </div>
            ) : null}
            {resultsMore ? <p className="muted">{copy.keepTyping}</p> : null}
            <div className="split">
              <label className="field">
                {copy.customerName}
                <input value={customerName} onChange={(event) => setCustomerName(event.target.value)} />
              </label>
              <label className="field">
                {copy.phone}
                <input value={phone} onChange={(event) => setPhone(event.target.value)} />
              </label>
              <label className="field">
                {copy.email}
                <input value={email} onChange={(event) => setEmail(event.target.value)} />
              </label>
            </div>
          </>
        )}
      </section>

      <section className="panel stack">
        <h2>{copy.devices}</h2>
        <label className="field">
          {copy.deviceName}
          <input value={name} onChange={(event) => setName(event.target.value)} />
        </label>
        <div className="split">
          <label className="field">
            {copy.brand}
            <input value={brand} onChange={(event) => setBrand(event.target.value)} />
          </label>
          <label className="field">
            {copy.model}
            <input value={model} onChange={(event) => setModel(event.target.value)} />
          </label>
        </div>
        <label className="field">
          {copy.serial}
          <input value={serial} onChange={(event) => setSerial(event.target.value)} />
        </label>
        <label className="field">
          {copy.conditionNote}
          <textarea value={note} onChange={(event) => setNote(event.target.value)} />
        </label>
        <PhotoFields
          copy={copy}
          label={copy.photos}
          existing={deviceExisting}
          files={deviceFiles}
          deviceId={deviceId}
          onAdd={(list) => setDeviceFiles((current) => [...current, ...takeFiles(list)])}
          onDropExisting={(id) => setDeviceExisting((current) => current.filter((photo) => photo.id !== id))}
          onDropFile={(file) => {
            dropFile(file);
            setDeviceFiles((current) => current.filter((item) => item.key !== file.key));
          }}
        />
      </section>

      <section className="panel stack">
        <div className="row">
          <h2>{copy.parts}</h2>
          <button
            className="button secondary"
            type="button"
            onClick={() =>
              setParts((current) => [
                ...current,
                { key: crypto.randomUUID(), id: "", name: "", note: "", existing: [], files: [] },
              ])
            }
          >
            {copy.addPart}
          </button>
        </div>
        {parts.length === 0 ? <p className="muted">{copy.noParts}</p> : null}
        {parts.map((part, index) => (
          <div className="stack part" key={part.key}>
            <div className="row">
              <h2>
                {copy.parts} {index + 1}
              </h2>
              <button
                className="button secondary"
                type="button"
                onClick={() => {
                  part.files.forEach(dropFile);
                  setParts((current) => current.filter((item) => item.key !== part.key));
                }}
              >
                {copy.removePart}
              </button>
            </div>
            <label className="field">
              {copy.partName}
              <input
                value={part.name}
                onChange={(event) =>
                  setParts((current) =>
                    current.map((item) => (item.key === part.key ? { ...item, name: event.target.value } : item)),
                  )
                }
              />
            </label>
            <label className="field">
              {copy.partNote}
              <textarea
                value={part.note}
                onChange={(event) =>
                  setParts((current) =>
                    current.map((item) => (item.key === part.key ? { ...item, note: event.target.value } : item)),
                  )
                }
              />
            </label>
            <PhotoFields
              copy={copy}
              label={copy.photos}
              existing={part.existing}
              files={part.files}
              deviceId={deviceId}
              onAdd={(list) =>
                setParts((current) =>
                  current.map((item) =>
                    item.key === part.key ? { ...item, files: [...item.files, ...takeFiles(list)] } : item,
                  ),
                )
              }
              onDropExisting={(id) =>
                setParts((current) =>
                  current.map((item) =>
                    item.key === part.key
                      ? { ...item, existing: item.existing.filter((photo) => photo.id !== id) }
                      : item,
                  ),
                )
              }
              onDropFile={(file) => {
                dropFile(file);
                setParts((current) =>
                  current.map((item) =>
                    item.key === part.key ? { ...item, files: item.files.filter((entry) => entry.key !== file.key) } : item,
                  ),
                );
              }}
            />
          </div>
        ))}
      </section>

      {error ? <p className="alert">{error}</p> : null}
      <button className="button" type="submit" disabled={saving}>
        {copy.save}
      </button>
    </form>
  );
}

function PhotoFields({
  copy,
  label,
  existing,
  files,
  deviceId,
  onAdd,
  onDropExisting,
  onDropFile,
}: {
  copy: Copy;
  label: string;
  existing: ExistingPhoto[];
  files: LocalFile[];
  deviceId?: string;
  onAdd: (list: FileList | null) => void;
  onDropExisting: (id: string) => void;
  onDropFile: (file: LocalFile) => void;
}) {
  return (
    <div className="stack">
      <label className="field">
        {label}
        <input
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          onChange={(event) => {
            onAdd(event.target.files);
            event.target.value = "";
          }}
        />
      </label>
      <div className="photos">
        {existing.map((photo) => (
          <figure className="photo" key={photo.id}>
            {deviceId ? <img src={photoSrc(deviceId, photo.id)} alt="" /> : null}
            <button className="button secondary" type="button" onClick={() => onDropExisting(photo.id)}>
              {copy.removePhoto}
            </button>
          </figure>
        ))}
        {files.map((file) => (
          <figure className="photo" key={file.key}>
            <img src={file.url} alt="" />
            <button className="button secondary" type="button" onClick={() => onDropFile(file)}>
              {copy.removePhoto}
            </button>
          </figure>
        ))}
      </div>
    </div>
  );
}

function emailOk(value: string) {
  const email = value.trim();
  const at = email.lastIndexOf("@");
  return at > 0 && !email.includes(" ") && email.slice(at + 1).includes(".");
}
