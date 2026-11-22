import { useEffect, useState } from "react";
import { ApiError, request, type Customer, type Page, type Role } from "./api";
import { canExport, downloadExcel } from "./export";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
import { End, useQuery, useWindow } from "./list";
import { messageFor } from "./Login";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  role: Role;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function Customers({ copy, lang, role, onNavigate, onUnauthorized }: Props) {
  const search = useQuery();
  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);
  const list = useWindow(search.query, (after: Customer | null, signal) => {
    const params = new URLSearchParams();
    if (search.query) {
      params.set("q", search.query);
    }
    if (after) {
      params.set("after", after.name);
      params.set("afterId", String(after.id));
    }
    return request<Page<Customer>>(`/api/customers?${params}`, { signal });
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
    const query = params.toString();
    setExporting(true);
    setExportError(null);
    try {
      await downloadExcel(`/api/customers/export${query ? `?${query}` : ""}`, lang);
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
            <h1>{copy.customers}</h1>
            <p className="lead">{copy.customersLead}</p>
          </div>
          {showExport ? (
            <div className="page-actions">
              <button className="button secondary" type="button" disabled={exporting} onClick={() => void exportList()}>
                <Icon name="download" size={16} />
                {copy.exportExcel}
              </button>
            </div>
          ) : null}
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
            placeholder={copy.searchCustomers}
            onChange={(event) => search.setValue(event.target.value)}
          />
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
                <th>{copy.customerName}</th>
                <th>{copy.phone}</th>
                <th>{copy.email}</th>
                <th>{copy.address}</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {list.items.map((customer) => (
                <tr
                  key={customer.id}
                  className="row-link"
                  onClick={() => onNavigate(`/customers/${customer.id}`)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      onNavigate(`/customers/${customer.id}`);
                    }
                  }}
                  tabIndex={0}
                >
                  <td data-label={copy.customerName} className="cell-main cell-strong">
                    {customer.name}
                  </td>
                  <td data-label={copy.phone}>{customer.phone}</td>
                  <td data-label={copy.email}>{customer.email}</td>
                  <td data-label={copy.address} className="wrap">
                    {customer.address}
                  </td>
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
              <Icon name="customers" size={18} />
            </span>
            <span>{search.query ? copy.noCustomerMatches : copy.noCustomers}</span>
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
