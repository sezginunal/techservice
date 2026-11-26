import { useEffect, useRef, useState, type FormEvent } from "react";
import { ApiError, request, type Customer, type DeviceItem, type Page, type ServiceState } from "./api";
import { ModelIcon } from "./ModelIcon";
import { formatWhen, statusLabel, statusTone } from "./format";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
import { End, useWindow } from "./list";
import { emailOk, phoneOk } from "./fields";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  id: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function CustomerDetail({ copy, lang, states, id, onNavigate, onUnauthorized }: Props) {
  const [customer, setCustomer] = useState<Customer | null>(null);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  const [address, setAddress] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const copyRef = useRef(copy);
  copyRef.current = copy;
  const unauthorizedRef = useRef(onUnauthorized);
  unauthorizedRef.current = onUnauthorized;
  const devices = useWindow(id, (after: DeviceItem | null, signal) => {
    const params = new URLSearchParams({ customerId: id });
    if (after) {
      params.set("after", after.createdAt);
      params.set("afterId", after.id);
    }
    return request<Page<DeviceItem>>(`/api/services?${params}`, { signal });
  });

  useEffect(() => {
    const controller = new AbortController();
    request<Customer>(`/api/customers/${id}`, { signal: controller.signal })
      .then((loaded) => {
        setCustomer(loaded);
        setName(loaded.name);
        setPhone(loaded.phone);
        setEmail(loaded.email);
        setAddress(loaded.address);
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        if (err instanceof ApiError && err.code === "unauthorized") {
          unauthorizedRef.current();
          return;
        }
        setError(problemText(copyRef.current, err));
      });
    return () => controller.abort();
  }, [id]);

  useEffect(() => {
    if (devices.error instanceof ApiError && devices.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [devices.error, onUnauthorized]);

  async function save(event: FormEvent) {
    event.preventDefault();
    const missing: string[] = [];
    const invalid: string[] = [];
    if (!name.trim()) {
      missing.push("customerName");
    }
    if (!phone.trim()) {
      missing.push("phone");
    } else if (!phoneOk(phone)) {
      invalid.push("phone");
    }
    if (!emailOk(email)) {
      invalid.push("email");
    }
    if (missing.length || invalid.length) {
      setError(problemText(copy, new ApiError("validation", { missing, invalid })));
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const saved = await request<Customer>(`/api/customers/${id}`, {
        method: "PATCH",
        body: JSON.stringify({ name, phone, email, address }),
      });
      setCustomer(saved);
      setName(saved.name);
      setPhone(saved.phone);
      setEmail(saved.email);
      setAddress(saved.address);
    } catch (err) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setError(problemText(copy, err));
    } finally {
      setSaving(false);
    }
  }

  if (!customer) {
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
        <ol className="crumbs">
          <li>
            <button className="link" type="button" onClick={() => onNavigate("/customers")}>
              {copy.customers}
            </button>
            <Icon name="chevronRight" size={14} />
          </li>
          <li aria-current="page">{customer.name}</li>
        </ol>
        <div className="page-title">
          <div className="page-title-main">
            <span className="page-media" aria-hidden="true">
              <Icon name="user" size={20} />
            </span>
            <div className="page-title-text">
              <h1>{customer.name}</h1>
            </div>
          </div>
          <div className="page-actions">
            <button className="button" type="button" onClick={() => onNavigate(`/services/new?customer=${customer.id}`)}>
              <Icon name="plus" size={16} />
              {copy.newDevice}
            </button>
          </div>
        </div>
      </header>
      <div className="detail">
        <div className="detail-main">
          <form className="panel" onSubmit={(event) => void save(event)}>
            <div className="panel-head">
              <div className="panel-title">
                <h2>{copy.customer}</h2>
              </div>
            </div>
            <div className="fields-2">
              <label className="field">
                {copy.customerName}
                <input value={name} onChange={(event) => setName(event.target.value)} />
              </label>
              <label className="field">
                {copy.phone}
                <input value={phone} inputMode="tel" autoComplete="tel" onChange={(event) => setPhone(event.target.value)} />
              </label>
              <label className="field">
                {copy.email}
                <input value={email} inputMode="email" autoComplete="email" onChange={(event) => setEmail(event.target.value)} />
              </label>
              <label className="field">
                {copy.address}
                <input value={address} autoComplete="street-address" onChange={(event) => setAddress(event.target.value)} />
              </label>
            </div>
            {error ? <p className="alert">{error}</p> : null}
            <div className="form-actions">
              <button className="button" type="submit" disabled={saving}>
                {copy.save}
              </button>
            </div>
          </form>
        </div>
        <div className="detail-side">
          <section className="panel flush">
            <div className="panel-head">
              <div className="panel-title">
                <h2>{copy.customerDevices}</h2>
              </div>
            </div>
            {devices.items.length > 0 ? (
              <ul className="rows">
                {devices.items.map((item) => (
                  <li key={item.id}>
                    <button className="row" type="button" onClick={() => onNavigate(`/services/${item.id}`)}>
                      <span className="row-main">
                        {item.image ? <ModelIcon brand={item.brand} model={item.model} /> : (
                          <span className="thing-media">
                            <Icon name="device" size={16} />
                          </span>
                        )}
                        <span className="row-text">
                          <span className="row-title">
                            {item.brand} {item.model} · <span className="mono">{item.serial}</span>
                          </span>
                        </span>
                      </span>
                      <span className="row-end">
                        <span className="badge" data-tone={statusTone(item.status)}>
                          <span>{statusLabel(item.status, states, lang)}</span>
                        </span>
                        <span className="row-sub">{formatWhen(item.createdAt, lang)}</span>
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            ) : null}
            {!devices.loading && devices.items.length === 0 ? (
              <div className="empty compact">
                <span className="empty-icon">
                  <Icon name="device" size={16} />
                </span>
                <span>{copy.noCustomerDevices}</span>
              </div>
            ) : null}
            {devices.loading ? (
              <div className="loading">
                <span className="spinner" />
                {copy.loading}
              </div>
            ) : null}
            <End enabled={devices.hasMore && !devices.loading} onEnd={devices.more} />
          </section>
        </div>
      </div>
    </div>
  );
}
