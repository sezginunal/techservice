import { useEffect, useState } from "react";
import { ApiError, request, type DeviceItem, type Page, type Role, type ServiceState } from "./api";
import { canExport, downloadExcel } from "./export";
import { ModelIcon } from "./ModelIcon";
import { formatWhen, listedStates, statusLabel, statusTone } from "./format";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
import { End, useQuery, useWindow } from "./list";
import { messageFor } from "./Login";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  role: Role;
  states: ServiceState[];
  status: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function Devices({ copy, lang, role, states, status, onNavigate, onUnauthorized }: Props) {
  const search = useQuery();
  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);
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

  async function exportList() {
    const params = new URLSearchParams();
    if (search.query) {
      params.set("q", search.query);
    }
    if (selected) {
      params.set("status", selected);
    }
    const query = params.toString();
    setExporting(true);
    setExportError(null);
    try {
      await downloadExcel(`/api/services/export${query ? `?${query}` : ""}`, lang);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setExportError(messageFor(copy, err));
    } finally {
      setExporting(false);
    }
  }

  const empty = !list.loading && list.items.length === 0;
  const showExport = canExport(role);

  return (
    <div className="page">
      <header className="page-head">
        <div className="page-title">
          <div className="page-title-text">
            <h1>{copy.devices}</h1>
            <p className="lead">{copy.devicesLead}</p>
          </div>
          <div className="page-actions">
            {showExport ? (
              <button className="button secondary" type="button" disabled={exporting} onClick={() => void exportList()}>
                <Icon name="download" size={16} />
                {copy.exportExcel}
              </button>
            ) : null}
            <button className="button" type="button" onClick={() => onNavigate("/services/new")}>
              <Icon name="plus" size={16} />
              {copy.newDevice}
            </button>
          </div>
        </div>
      </header>
      {exportError ? <p className="alert">{exportError}</p> : null}
      <form className="toolbar panel" onSubmit={search.commit}>
        <label className="field search">
          <span className="sr-only">{copy.search}</span>
          <Icon name="search" size={16} />
          <input
            type="search"
            value={search.value}
            placeholder={copy.searchDevices}
            onChange={(event) => search.setValue(event.target.value)}
          />
        </label>
        <label className="field">
          <span className="sr-only">{copy.status}</span>
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
        <button className="button secondary" type="submit">
          {copy.search}
        </button>
      </form>
      {list.error ? <p className="alert">{problemText(copy, list.error)}</p> : null}
      <div className="panel flush">
        <div className="table-wrap">
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
                <th />
              </tr>
            </thead>
            <tbody>
              {list.items.map((item) => (
                <tr
                  key={item.id}
                  className="row-link"
                  onClick={() => onNavigate(`/services/${item.id}`)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      onNavigate(`/services/${item.id}`);
                    }
                  }}
                  tabIndex={0}
                >
                  <td data-label={copy.recordId} className="uuid cell-main" title={item.id}>
                    <span>{item.id}</span>
                  </td>
                  <td data-label={copy.brand}>{item.brand}</td>
                  <td data-label={copy.model}>
                    <span className="thing">
                      {item.image ? <ModelIcon brand={item.brand} model={item.model} /> : null}
                      <span className="thing-title">{item.model}</span>
                    </span>
                  </td>
                  <td data-label={copy.serial} className="mono">
                    {item.serial}
                  </td>
                  <td data-label={copy.customer}>{item.customerName}</td>
                  <td data-label={copy.status} className="cell-side">
                    <span className="badge" data-tone={statusTone(item.status)}>
                      <span>{statusLabel(item.status, states, lang)}</span>
                    </span>
                  </td>
                  <td data-label={copy.registeredOn}>{formatWhen(item.createdAt, lang)}</td>
                  <td className="cell-end">
                    <Icon name="chevronRight" size={16} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {empty ? (
          <div className="empty">
            <span className="empty-icon">
              <Icon name="device" size={18} />
            </span>
            <span>{search.query || selected ? copy.noDeviceMatches : copy.noDevices}</span>
          </div>
        ) : null}
        {list.loading ? (
          <div className="loading">
            <span className="spinner" />
            {copy.loading}
          </div>
        ) : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </div>
    </div>
  );
}
