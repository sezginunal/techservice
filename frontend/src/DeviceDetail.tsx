import { useEffect, useState } from "react";
import { ApiError, photoSrc, request, type Device } from "./api";
import { formatWhen, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { problemText } from "./problems";

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

  if (!device) {
    return <p className={error ? "alert" : "muted"}>{error ?? copy.loading}</p>;
  }

  return (
    <>
      <header className="page-head row">
        <div className="stack">
          <h1>{device.name}</h1>
          <p className="uuid">{device.id}</p>
          <p className="muted">
            {copy.registeredOn}: {formatWhen(device.createdAt, lang)}
          </p>
          <span className="badge">{statusLabel(device.status, copy.statusRegistered)}</span>
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
        <p>
          {device.brand} {device.model}
        </p>
        {device.serial ? (
          <p>
            {copy.serial}: {device.serial}
          </p>
        ) : null}
        <p>{device.note}</p>
        <PhotoRow deviceId={device.id} photos={device.photos} />
      </section>
      <section className="panel stack">
        <h2>{copy.parts}</h2>
        {device.parts.length === 0 ? <p className="muted">{copy.noParts}</p> : null}
        {device.parts.map((part) => (
          <div className="stack part" key={part.id}>
            <h2>{part.name}</h2>
            <p>{part.note}</p>
            <PhotoRow deviceId={device.id} photos={part.photos} />
          </div>
        ))}
      </section>
    </>
  );
}

function PhotoRow({ deviceId, photos }: { deviceId: string; photos: { id: string }[] }) {
  if (photos.length === 0) {
    return null;
  }
  return (
    <div className="photos">
      {photos.map((photo) => (
        <img key={photo.id} className="thumb" src={photoSrc(deviceId, photo.id)} alt="" />
      ))}
    </div>
  );
}
