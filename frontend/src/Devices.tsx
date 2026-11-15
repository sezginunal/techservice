import { useEffect } from "react";
import { ApiError, request, type DeviceItem, type Page, type ServiceState } from "./api";
import { ModelIcon } from "./ModelIcon";
import { formatWhen, listedStates, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { End, useQuery, useWindow } from "./list";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  status: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function Devices({ copy, lang, states, status, onNavigate, onUnauthorized }: Props) {
  const search = useQuery();
  const choices = listedStates(states);
  const selected = choices.some((item) => item.code === status) ? status : "";
  const list = useWindow(`${search.query}|${selected}`, (after: DeviceItem | null, signal) => {
    const params = new URLSearchParams();
    if (search.query) {
      params.set("q", search.query);
    }
    if (selected) {
      params.set("status", selected);
    }
    if (after) {
      params.set("after", after.createdAt);
      params.set("afterId", after.id);
    }
    return request<Page<DeviceItem>>(`/api/services?${params}`, { signal });
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
        <button className="button" type="button" onClick={() => onNavigate("/services/new")}>
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
          <select
            value={selected}
            onChange={(event) => {
              const next = event.target.value;
              onNavigate(next ? `/services?status=${encodeURIComponent(next)}` : "/services");
            }}
          >
            <option value="">{copy.allStatuses}</option>
            {choices.map((item) => (
              <option key={item.code} value={item.code}>
                {statusLabel(item.code, states, lang)}
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
              <th>{copy.brand}</th>
              <th>{copy.model}</th>
              <th>{copy.serial}</th>
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
                <td data-label={copy.brand}>{item.brand}</td>
                <td data-label={copy.model}>
                  <span className="with-icon">
                    {item.image ? <ModelIcon brand={item.brand} model={item.model} /> : null}
                    {item.model}
                  </span>
                </td>
                <td data-label={copy.serial}>{item.serial}</td>
                <td data-label={copy.customer}>{item.customerName}</td>
                <td data-label={copy.status}>
                  <span className="badge">{statusLabel(item.status, states, lang)}</span>
                </td>
                <td data-label={copy.registeredOn}>{formatWhen(item.createdAt, lang)}</td>
                <td className="row-action">
                  <button className="button secondary" type="button" onClick={() => onNavigate(`/services/${item.id}`)}>
                    {copy.open}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {empty ? (
          <p className="muted empty">{search.query || selected ? copy.noDeviceMatches : copy.noDevices}</p>
        ) : null}
        {list.loading ? <p className="muted empty">{copy.loading}</p> : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </div>
    </>
  );
}
