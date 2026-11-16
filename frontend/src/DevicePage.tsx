import { useEffect, useState } from "react";
import { ApiError, request, type DeviceService, type Identity, type Page, type ServiceState } from "./api";
import { ModelIcon } from "./ModelIcon";
import { formatWhen, statusLabel, statusTone } from "./format";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
import { End, useWindow } from "./list";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  id: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function DevicePage({ copy, lang, states, id, onNavigate, onUnauthorized }: Props) {
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
    return (
      <div className="page">
        {error ? <p className="alert">{error}</p> : (
          <div className="loading">
            <span className="spinner" />
            {copy.loading}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="page">
      <header className="page-head">
        <div className="page-title">
          <div className="page-title-main">
            <span className="page-media">
              {device.image ? <ModelIcon brand={device.brand} model={device.model} /> : <Icon name="device" size={20} />}
            </span>
            <div className="page-title-text">
              <h1>
                {device.brand} {device.model}
              </h1>
              <p className="meta">
                <span className="meta-item">
                  <Icon name="hash" size={14} />
                  <span className="mono">{device.serial}</span>
                </span>
              </p>
            </div>
          </div>
        </div>
      </header>
      <section className="panel flush">
        <div className="panel-head">
          <div className="panel-title">
            <h2>{copy.customerDevices}</h2>
          </div>
        </div>
        {list.items.length > 0 ? (
          <ul className="rows">
            {list.items.map((item) => (
              <li key={item.id}>
                <button className="row" type="button" onClick={() => onNavigate(`/services/${item.id}`)}>
                  <span className="row-main">
                    <span className="row-text">
                      <span className="row-title">{formatWhen(item.createdAt, lang)}</span>
                      <span className="row-sub">{item.customerName}</span>
                    </span>
                  </span>
                  <span className="row-end">
                    <span className="badge" data-tone={statusTone(item.status)}>
                      <span>{statusLabel(item.status, states, lang)}</span>
                    </span>
                  </span>
                </button>
              </li>
            ))}
          </ul>
        ) : null}
        {list.error ? <p className="alert">{problemText(copy, list.error)}</p> : null}
        {list.loading ? (
          <div className="loading">
            <span className="spinner" />
            {copy.loading}
          </div>
        ) : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </section>
    </div>
  );
}
