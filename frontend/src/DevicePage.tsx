import { useEffect, useState } from "react";
import { ApiError, request, type DeviceService, type Identity, type Page } from "./api";
import { ModelIcon } from "./ModelIcon";
import { formatWhen, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { End, useWindow } from "./list";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  id: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function DevicePage({ copy, lang, id, onNavigate, onUnauthorized }: Props) {
  const [device, setDevice] = useState<Identity | null>(null);
  const [error, setError] = useState<string | null>(null);
  const list = useWindow(id, (after: DeviceService | null, signal) => {
    const params = new URLSearchParams();
    if (after) {
      params.set("after", after.createdAt);
      params.set("afterId", after.id);
    }
    return request<Page<DeviceService>>(`/api/devices/${id}/services?${params}`, { signal });
  });

  useEffect(() => {
    const controller = new AbortController();
    request<Identity>(`/api/devices/${id}`, { signal: controller.signal })
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

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  if (!device) {
    return <p className={error ? "alert" : "muted"}>{error ?? copy.loading}</p>;
  }

  return (
    <>
      <header className="page-head">
        <h1 className="with-icon">
          {device.image ? <ModelIcon brand={device.brand} model={device.model} /> : null}
          {device.brand} {device.model}
        </h1>
        <p className="muted">
          {copy.serial}: {device.serial}
        </p>
      </header>
      <section className="panel stack">
        <h2>{copy.customerDevices}</h2>
        {list.items.map((item) => (
          <button className="button secondary" type="button" key={item.id} onClick={() => onNavigate(`/services/${item.id}`)}>
            <strong>{formatWhen(item.createdAt, lang)}</strong>
            <span className="muted">
              {item.customerName} · {statusLabel(item.status, copy)}
            </span>
          </button>
        ))}
        {list.error ? <p className="alert">{problemText(copy, list.error)}</p> : null}
        {list.loading ? <p className="muted">{copy.loading}</p> : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </section>
    </>
  );
}
