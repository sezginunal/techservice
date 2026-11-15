import { useEffect, useState, type FormEvent } from "react";
import { ApiError, request, type Page, type SparePart } from "./api";
import { wholeQuantity } from "./fields";
import type { Copy } from "./i18n";
import { End, useQuery, useWindow } from "./list";
import { problemText } from "./problems";
import { SpareField } from "./SpareField";
import { maxStock, nameFields, rangeText, spareProblem } from "./spare";

type Props = {
  copy: Copy;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function SpareParts({ copy, onNavigate, onUnauthorized }: Props) {
  const search = useQuery();
  const list = useWindow(search.query, (after: SparePart | null, signal) => {
    const params = new URLSearchParams();
    if (search.query) {
      params.set("q", search.query);
    }
    if (after) {
      params.set("after", after.brand);
      params.set("afterModel", after.model);
      params.set("afterId", after.id);
    }
    return request<Page<SparePart>>(`/api/spare-parts?${params}`, { signal });
  });
  const [brand, setBrand] = useState("");
  const [model, setModel] = useState("");
  const [quantity, setQuantity] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [existing, setExisting] = useState<SparePart | null>(null);

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  async function create(event: FormEvent) {
    event.preventDefault();
    const names = nameFields(brand, model);
    const amount = wholeQuantity(quantity, 0, maxStock);
    if (names.missing.length || names.tooLong.length || typeof amount !== "number") {
      const lines = [problemText(copy, new ApiError("validation", names))];
      if (amount === "empty") {
        lines.push(copy.quantityRequired);
      }
      if (amount === "range") {
        lines.push(rangeText(copy, 0, maxStock));
      }
      setError(lines.filter(Boolean).join(" "));
      setExisting(null);
      return;
    }
    setSaving(true);
    setError(null);
    setExisting(null);
    try {
      const created = await request<SparePart>("/api/spare-parts", {
        method: "POST",
        body: JSON.stringify({ brand, model, quantity: amount }),
      });
      if (created.existed && !created.added) {
        setExisting(created);
        return;
      }
      const notice = created.existed && created.added ? "?notice=added" : "";
      onNavigate(`/spare-parts/${created.id}${notice}`);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setError(spareProblem(copy, err, 0, maxStock) ?? problemText(copy, err));
    } finally {
      setSaving(false);
    }
  }

  const empty = !list.loading && list.items.length === 0;

  return (
    <>
      <header className="page-head">
        <h1>{copy.spareParts}</h1>
        <p className="muted">{copy.sparePartsLead}</p>
      </header>
      <form className="panel stack" onSubmit={(event) => void create(event)}>
        <h2>{copy.addSparePart}</h2>
        <SpareField id="spare-brand" label={copy.brand} value={brand} onChange={setBrand} onUnauthorized={onUnauthorized} />
        <SpareField
          id="spare-model"
          label={copy.model}
          value={model}
          brand={brand}
          onChange={setModel}
          onUnauthorized={onUnauthorized}
        />
        <label className="field">
          {copy.quantity}
          <input inputMode="numeric" value={quantity} onChange={(event) => setQuantity(event.target.value)} />
        </label>
        {error ? <p className="alert">{error}</p> : null}
        {existing ? (
          <p className="alert">
            {copy.spareExists}{" "}
            <button className="button linkish" type="button" onClick={() => onNavigate(`/spare-parts/${existing.id}`)}>
              {copy.openSparePart}
            </button>
          </p>
        ) : null}
        <button className="button" type="submit" disabled={saving}>
          {copy.addSparePart}
        </button>
      </form>
      <form className="search-row" onSubmit={search.commit}>
        <label className="field">
          {copy.search}
          <input
            value={search.value}
            placeholder={copy.spareSearch}
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
              <th>{copy.brand}</th>
              <th>{copy.model}</th>
              <th>{copy.inStock}</th>
            </tr>
          </thead>
          <tbody>
            {list.items.map((part) => (
              <tr key={part.id}>
                <td data-label={copy.brand}>
                  <button className="button linkish" type="button" onClick={() => onNavigate(`/spare-parts/${part.id}`)}>
                    {part.brand}
                  </button>
                </td>
                <td data-label={copy.model}>{part.model}</td>
                <td data-label={copy.inStock}>{part.onHand}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {empty ? <p className="muted empty">{search.query ? copy.noSpareMatches : copy.noSpareParts}</p> : null}
        {list.loading ? <p className="muted empty">{copy.loading}</p> : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </div>
    </>
  );
}
