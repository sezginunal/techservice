import { useEffect, useState, type FormEvent } from "react";
import { ApiError, photoSrc, request, statusPhotoSrc, type Device } from "./api";
import { ModelIcon } from "./ModelIcon";
import { deviceStatuses, formatWhen, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { PhotoButton } from "./PhotoView";
import { fileProblem, problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  id: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function DeviceDetail({ copy, lang, id, onNavigate, onUnauthorized }: Props) {
  const [device, setDevice] = useState<Device | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [next, setNext] = useState("");
  const [note, setNote] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [fileKey, setFileKey] = useState(0);
  const [statusError, setStatusError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    request<Device>(`/api/devices/${id}`, { signal: controller.signal })
      .then(setDevice)
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
  }, [copy, id, onUnauthorized]);

  async function saveStatus(event: FormEvent) {
    event.preventDefault();
    if (!device) {
      return;
    }
    const missing = next ? [] : ["status"];
    const photo = file ? fileProblem(file) : null;
    if (missing.length || photo) {
      setStatusError(problemText(copy, new ApiError("validation", { missing, photo: photo ?? "" })));
      return;
    }
    const body = new FormData();
    body.set("payload", JSON.stringify({ status: next, note }));
    if (file) {
      body.set("file0", file);
    }
    setSaving(true);
    setStatusError(null);
    try {
      const saved = await request<Device>(`/api/devices/${device.id}/status`, { method: "POST", body });
      setDevice(saved);
      setNext("");
      setNote("");
      setFile(null);
      setFileKey((value) => value + 1);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setStatusError(problemText(copy, err));
    } finally {
      setSaving(false);
    }
  }

  if (!device) {
    return <p className={error ? "alert" : "muted"}>{error ?? copy.loading}</p>;
  }

  return (
    <>
      <header className="page-head row">
        <div className="stack">
          <h1 className="with-icon">
            {device.image ? <ModelIcon brand={device.brand} model={device.model} /> : null}
            {device.brand} {device.model}
          </h1>
          <p className="uuid">{device.id}</p>
          <p className="muted">
            {copy.registeredOn}: {formatWhen(device.createdAt, lang)}
          </p>
          <span className="badge">{statusLabel(device.status, copy)}</span>
        </div>
        <button className="button" type="button" onClick={() => onNavigate(`/devices/${device.id}/edit`)}>
          {copy.correct}
        </button>
      </header>
      <section className="panel stack">
        <h2>{copy.customer}</h2>
        <p>{device.customer.name}</p>
        <p>{device.customer.phone}</p>
        <p>{device.customer.email}</p>
      </section>
      <section className="panel stack">
        <h2>{copy.devices}</h2>
        {device.serial ? (
          <p>
            {copy.serial}: {device.serial}
          </p>
        ) : null}
        <p>{device.note}</p>
        <PhotoRow copy={copy} deviceId={device.id} photos={device.photos} onUnauthorized={onUnauthorized} />
      </section>
      <section className="panel stack">
        <h2>{copy.parts}</h2>
        {device.parts.length === 0 ? <p className="muted">{copy.noParts}</p> : null}
        {device.parts.map((part) => (
          <div className="stack part" key={part.id}>
            <h2>{part.name}</h2>
            <p>{part.note}</p>
            <PhotoRow copy={copy} deviceId={device.id} photos={part.photos} onUnauthorized={onUnauthorized} />
          </div>
        ))}
      </section>
      <form className="panel stack" onSubmit={(event) => void saveStatus(event)}>
        <h2>{copy.changeStatus}</h2>
        <label className="field">
          {copy.status}
          <select value={next} onChange={(event) => setNext(event.target.value)}>
            <option value="">{copy.statusRequired}</option>
            {deviceStatuses.map((id) => (
              <option key={id} value={id}>
                {statusLabel(id, copy)}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          {copy.statusNote}
          <textarea value={note} onChange={(event) => setNote(event.target.value)} />
        </label>
        <label className="field">
          {copy.statusPhoto}
          <input
            key={fileKey}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            onChange={(event) => setFile(event.target.files?.[0] ?? null)}
          />
        </label>
        {statusError ? <p className="alert">{statusError}</p> : null}
        <button className="button" type="submit" disabled={saving}>
          {copy.save}
        </button>
      </form>
      <section className="panel stack">
        <h2>{copy.statusHistory}</h2>
        {device.history.length === 0 ? <p className="muted">{copy.noStatusHistory}</p> : null}
        {device.history.map((change) => (
          <div className="stack part" key={change.id}>
            <p>
              {formatWhen(change.createdAt, lang)} · {change.actorUsername}
            </p>
            <p>
              {statusLabel(change.from, copy)} → {statusLabel(change.to, copy)}
            </p>
            {change.note ? <p>{change.note}</p> : null}
            {change.hasPhoto ? (
              <PhotoButton
                copy={copy}
                className="thumb"
                src={statusPhotoSrc(device.id, change.id)}
                name={String(change.id)}
                onUnauthorized={onUnauthorized}
              />
            ) : null}
          </div>
        ))}
      </section>
    </>
  );
}

function PhotoRow({
  copy,
  deviceId,
  photos,
  onUnauthorized,
}: {
  copy: Copy;
  deviceId: string;
  photos: { id: string }[];
  onUnauthorized: () => void;
}) {
  if (photos.length === 0) {
    return null;
  }
  return (
    <div className="photos">
      {photos.map((photo) => (
        <PhotoButton
          key={photo.id}
          copy={copy}
          className="thumb"
          src={photoSrc(deviceId, photo.id)}
          name={photo.id}
          onUnauthorized={onUnauthorized}
        />
      ))}
    </div>
  );
}
