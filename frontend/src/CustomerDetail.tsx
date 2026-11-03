import { useEffect, useState, type FormEvent } from "react";
import { ApiError, request, type Customer, type DeviceItem, type Page } from "./api";
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

export function CustomerDetail({ copy, lang, id, onNavigate, onUnauthorized }: Props) {
  const [customer, setCustomer] = useState<Customer | null>(null);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const devices = useWindow(id, (after: DeviceItem | null, signal) => {
    const params = new URLSearchParams({ customerId: id });
    if (after) {
      params.set("after", after.createdAt);
      params.set("afterId", after.id);
    }
    return request<Page<DeviceItem>>(`/api/devices?${params}`, { signal });
  });

  useEffect(() => {
    const controller = new AbortController();
    request<Customer>(`/api/customers/${id}`, { signal: controller.signal })
      .then((loaded) => {
        setCustomer(loaded);
        setName(loaded.name);
        setPhone(loaded.phone);
        setEmail(loaded.email);
      })
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
    if (devices.error instanceof ApiError && devices.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [devices.error, onUnauthorized]);

  async function save(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const saved = await request<Customer>(`/api/customers/${id}`, {
        method: "PATCH",
        body: JSON.stringify({ name, phone, email }),
      });
      setCustomer(saved);
      setName(saved.name);
      setPhone(saved.phone);
      setEmail(saved.email);
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
    return <p className={error ? "alert" : "muted"}>{error ?? copy.loading}</p>;
  }

  return (
    <>
      <header className="page-head row">
        <h1>{customer.name}</h1>
        <button className="button" type="button" onClick={() => onNavigate(`/devices/new?customer=${customer.id}`)}>
          {copy.newDevice}
        </button>
      </header>
      <form className="panel stack" onSubmit={(event) => void save(event)}>
        <div className="split">
          <label className="field">
            {copy.customerName}
            <input value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label className="field">
            {copy.phone}
            <input value={phone} onChange={(event) => setPhone(event.target.value)} />
          </label>
          <label className="field">
            {copy.email}
            <input value={email} onChange={(event) => setEmail(event.target.value)} />
          </label>
        </div>
        {error ? <p className="alert">{error}</p> : null}
        <button className="button" type="submit" disabled={saving}>
          {copy.save}
        </button>
      </form>
      <section className="panel stack">
        <h2>{copy.customerDevices}</h2>
        {devices.items.map((item) => (
          <button className="button secondary" type="button" key={item.id} onClick={() => onNavigate(`/devices/${item.id}`)}>
            <strong className="with-icon">
              {item.image ? <ModelIcon brand={item.brand} model={item.model} /> : null}
              {item.brand} {item.model}
            </strong>
            <span className="muted">
              {statusLabel(item.status, copy)} · {formatWhen(item.createdAt, lang)}
            </span>
          </button>
        ))}
        {!devices.loading && devices.items.length === 0 ? <p className="muted">{copy.noCustomerDevices}</p> : null}
        {devices.loading ? <p className="muted">{copy.loading}</p> : null}
        <End enabled={devices.hasMore && !devices.loading} onEnd={devices.more} />
      </section>
    </>
  );
}
