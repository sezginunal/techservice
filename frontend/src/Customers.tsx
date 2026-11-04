import { useEffect } from "react";
import { ApiError, request, type Customer, type Page } from "./api";
import type { Copy } from "./i18n";
import { End, useQuery, useWindow } from "./list";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function Customers({ copy, onNavigate, onUnauthorized }: Props) {
  const search = useQuery();
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

  const empty = !list.loading && list.items.length === 0;

  return (
    <>
      <header className="page-head">
        <h1>{copy.customers}</h1>
        <p className="muted">{copy.customersLead}</p>
      </header>
      <form className="search-row" onSubmit={search.commit}>
        <label className="field">
          {copy.search}
          <input
            value={search.value}
            placeholder={copy.searchCustomers}
            onChange={(event) => search.setValue(event.target.value)}
          />
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
              <th>{copy.customerName}</th>
              <th>{copy.phone}</th>
              <th>{copy.email}</th>
              <th>{copy.address}</th>
            </tr>
          </thead>
          <tbody>
            {list.items.map((customer) => (
              <tr key={customer.id}>
                <td data-label={copy.customerName}>
                  <button className="button linkish" type="button" onClick={() => onNavigate(`/customers/${customer.id}`)}>
                    {customer.name}
                  </button>
                </td>
                <td data-label={copy.phone}>{customer.phone}</td>
                <td data-label={copy.email}>{customer.email}</td>
                <td data-label={copy.address}>{customer.address}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {empty ? <p className="muted empty">{search.query ? copy.noCustomerMatches : copy.noCustomers}</p> : null}
        {list.loading ? <p className="muted empty">{copy.loading}</p> : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </div>
    </>
  );
}
