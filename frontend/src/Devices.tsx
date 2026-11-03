import { useEffect, useState } from "react";
import { ApiError, request, type DeviceItem, type Page } from "./api";
import { deviceStatuses, formatWhen, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { End, useQuery, useWindow } from "./list";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function Devices({ copy, lang, onNavigate, onUnauthorized }: Props) {
  const search = useQuery();
  const [status, setStatus] = useState("");
  const list = useWindow(`${search.query}|${status}`, (after: DeviceItem | null, signal) => {
    const params = new URLSearchParams();
    if (search.query) {
      params.set("q", search.query);
    }
    if (status) {
      params.set("status", status);
    }
    if (after) {
      params.set("after", after.createdAt);
      params.set("afterId", after.id);
    }
    return request<Page<DeviceItem>>(`/api/devices?${params}`, { signal });
  });

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  const empty = !list.loading && list.items.length === 0;

  return (
    <>
      <header className="page-head row">
        <div>
          <h1>{copy.devices}</h1>
          <p className="muted">{copy.devicesLead}</p>
        </div>
        <button className="button" type="button" onClick={() => onNavigate("/devices/new")}>
          {copy.newDevice}
        </button>
      </header>
      <form className="device-filters" onSubmit={search.commit}>
        <label className="field">
          {copy.search}
          <input
            value={search.value}
            placeholder={copy.searchDevices}
            onChange={(event) => search.setValue(event.target.value)}
          />
        </label>
        <label className="field">
          {copy.status}
          <select value={status} onChange={(event) => setStatus(event.target.value)}>
            <option value="">{copy.allStatuses}</option>
            {deviceStatuses.map((id) => (
              <option key={id} value={id}>
                {statusLabel(id, copy)}
              </option>
            ))}
          </select>
        </label>
        <button className="button" type="submit">
          {copy.search}
        </button>
      </form>
      {list.error ? <p className="alert">{problemText(copy, list.error)}</p> : null}
      <div className="panel table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>{copy.recordId}</th>
              <th>{copy.deviceName}</th>
              <th>{copy.brand}</th>
              <th>{copy.model}</th>
              <th>{copy.customer}</th>
              <th>{copy.status}</th>
              <th>{copy.registeredOn}</th>
              <th>{copy.open}</th>
            </tr>
          </thead>
          <tbody>
            {list.items.map((item) => (
              <tr key={item.id}>
                <td data-label={copy.recordId} className="uuid">
                  {item.id}
                </td>
                <td data-label={copy.deviceName}>{item.name}</td>
                <td data-label={copy.brand}>{item.brand}</td>
                <td data-label={copy.model}>{item.model}</td>
                <td data-label={copy.customer}>{item.customerName}</td>
                <td data-label={copy.status}>
                  <span className="badge">{statusLabel(item.status, copy)}</span>
                </td>
                <td data-label={copy.registeredOn}>{formatWhen(item.createdAt, lang)}</td>
                <td className="row-action">
                  <button className="button secondary" type="button" onClick={() => onNavigate(`/devices/${item.id}`)}>
                    {copy.open}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {empty ? (
          <p className="muted empty">{search.query || status ? copy.noDeviceMatches : copy.noDevices}</p>
        ) : null}
        {list.loading ? <p className="muted empty">{copy.loading}</p> : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </div>
    </>
  );
}
