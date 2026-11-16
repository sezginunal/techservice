import { useEffect, useState, type FormEvent } from "react";
import { ApiError, request, type Page, type ServiceState, type SpareMove, type SparePart } from "./api";
import { wholeQuantity } from "./fields";
import { formatWhen, statusLabel, statusTone } from "./format";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
import { End, useWindow } from "./list";
import { problemText } from "./problems";
import { SpareField } from "./SpareField";
import { maxStock, moveKind, nameFields, noteTooLong, rangeText, spareProblem } from "./spare";

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  id: string;
  notice: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function SparePartPage({ copy, lang, states, id, notice, onNavigate, onUnauthorized }: Props) {
  const [part, setPart] = useState<SparePart | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [revision, setRevision] = useState(0);
  const [brand, setBrand] = useState("");
  const [model, setModel] = useState("");
  const [identityError, setIdentityError] = useState<string | null>(null);
  const [savingIdentity, setSavingIdentity] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    request<SparePart>(`/api/spare-parts/${id}`, { signal: controller.signal })
      .then((loaded) => {
        if (!controller.signal.aborted) {
          setPart(loaded);
          setBrand(loaded.brand);
          setModel(loaded.model);
        }
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

  async function saveIdentity(event: FormEvent) {
    event.preventDefault();
    if (!part) {
      return;
    }
    const names = nameFields(brand, model);
    if (names.missing.length || names.tooLong.length) {
      setIdentityError(problemText(copy, new ApiError("validation", names)));
      return;
    }
    if (brand.trim() === part.brand && model.trim() === part.model) {
      setIdentityError(null);
      return;
    }
    setSavingIdentity(true);
    setIdentityError(null);
    try {
      const saved = await request<SparePart>(`/api/spare-parts/${part.id}`, {
        method: "PATCH",
        body: JSON.stringify({ brand, model }),
      });
      setPart(saved);
      setBrand(saved.brand);
      setModel(saved.model);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setIdentityError(spareProblem(copy, err, 1, maxStock) ?? problemText(copy, err));
    } finally {
      setSavingIdentity(false);
    }
  }

  if (!part) {
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
            <button className="link" type="button" onClick={() => onNavigate("/spare-parts")}>
              {copy.spareParts}
            </button>
            <Icon name="chevronRight" size={14} />
          </li>
          <li aria-current="page">
            {part.brand} {part.model}
          </li>
        </ol>
        <div className="page-title">
          <div className="page-title-text">
            <h1>
              {part.brand} {part.model}
            </h1>
            <p className="meta">
              <span className="stock" data-tone={part.onHand === 0 ? "red" : part.onHand < 5 ? "amber" : "green"}>
                {part.onHand}
              </span>
              <span className="meta-item">{copy.inStock}</span>
            </p>
          </div>
        </div>
      </header>
      {notice === "added" ? <p className="alert" data-tone="success">{copy.spareAddedExisting}</p> : null}
      <form className="panel" onSubmit={(event) => void saveIdentity(event)}>
        <div className="panel-head">
          <div className="panel-title">
            <h2>{copy.correct}</h2>
          </div>
        </div>
        <div className="fields-2">
          <SpareField id="part-brand" label={copy.brand} value={brand} onChange={setBrand} onUnauthorized={onUnauthorized} />
          <SpareField
            id="part-model"
            label={copy.model}
            value={model}
            brand={brand}
            onChange={setModel}
            onUnauthorized={onUnauthorized}
          />
        </div>
        {identityError ? <p className="alert">{identityError}</p> : null}
        <div className="form-actions">
          <button className="button" type="submit" disabled={savingIdentity}>
            {copy.save}
          </button>
        </div>
      </form>
      <StockForm
        copy={copy}
        title={copy.addStock}
        submitLabel={copy.addStock}
        min={1}
        max={maxStock}
        stock={part.onHand}
        withNote
        onUnauthorized={onUnauthorized}
        onSaved={(saved) => {
          setPart(saved);
          setRevision((value) => value + 1);
        }}
        save={(quantity, note) =>
          request<SparePart>(`/api/spare-parts/${part.id}/receipts`, {
            method: "POST",
            body: JSON.stringify({ quantity, note }),
          })
        }
      />
      <StockForm
        copy={copy}
        title={copy.writeOff}
        submitLabel={copy.writeOff}
        min={1}
        max={part.onHand}
        disabled={part.onHand === 0}
        withNote
        onUnauthorized={onUnauthorized}
        onSaved={(saved) => {
          setPart(saved);
          setRevision((value) => value + 1);
        }}
        onShort={(onHand) => setPart((current) => (current ? { ...current, onHand } : current))}
        save={(quantity, note) =>
          request<SparePart>(`/api/spare-parts/${part.id}/write-offs`, {
            method: "POST",
            body: JSON.stringify({ quantity, note }),
          })
        }
      />
      <MoveList
        copy={copy}
        lang={lang}
        states={states}
        title={copy.usedOnServices}
        empty={copy.spareNotUsed}
        path={`/api/spare-parts/${part.id}/uses`}
        revision={`${part.id}:${revision}`}
        mode="uses"
        onNavigate={onNavigate}
        onUnauthorized={onUnauthorized}
      />
      <MoveList
        copy={copy}
        lang={lang}
        states={states}
        title={copy.stockChanges}
        empty={copy.noStockChanges}
        path={`/api/spare-parts/${part.id}/stock`}
        revision={`${part.id}:${revision}:stock`}
        mode="stock"
        onNavigate={onNavigate}
        onUnauthorized={onUnauthorized}
      />
    </div>
  );
}

function StockForm({
  copy,
  title,
  submitLabel,
  min,
  max,
  stock,
  disabled,
  withNote,
  onUnauthorized,
  onSaved,
  onShort,
  save,
}: {
  copy: Copy;
  title: string;
  submitLabel: string;
  min: number;
  max: number;
  stock?: number;
  disabled?: boolean;
  withNote: boolean;
  onUnauthorized: () => void;
  onSaved: (part: SparePart) => void;
  onShort?: (onHand: number) => void;
  save: (quantity: number, note: string) => Promise<SparePart>;
}) {
  const [quantity, setQuantity] = useState("");
  const [note, setNote] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (disabled) {
      return;
    }
    const amount = wholeQuantity(quantity, min, max);
    if (typeof amount !== "number") {
      setError(amount === "empty" ? copy.quantityRequired : rangeText(copy, min, max));
      return;
    }
    if (stock !== undefined && stock + amount > maxStock) {
      setError(copy.stockCap);
      return;
    }
    if (withNote && noteTooLong(note)) {
      setError(problemText(copy, new ApiError("validation", { tooLong: ["spareNote"] })));
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const saved = await save(amount, note.trim());
      setQuantity("");
      setNote("");
      onSaved(saved);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      if (err instanceof ApiError && err.code === "stock_short") {
        onShort?.(err.onHand);
      }
      setError(spareProblem(copy, err, min, max) ?? problemText(copy, err));
    } finally {
      setSaving(false);
    }
  }

  return (
    <form className="panel" onSubmit={(event) => void submit(event)}>
      <div className="panel-head">
        <div className="panel-title">
          <h2>{title}</h2>
        </div>
      </div>
      <div className="fields">
        <label className="field">
          {copy.quantity}
          <input inputMode="numeric" value={quantity} onChange={(event) => setQuantity(event.target.value)} />
        </label>
        {withNote ? (
          <label className="field">
            {copy.optionalNote}
            <textarea value={note} onChange={(event) => setNote(event.target.value)} />
          </label>
        ) : null}
      </div>
      {error ? <p className="alert">{error}</p> : null}
      <div className="form-actions">
        <button className="button" type="submit" disabled={disabled || saving}>
          {submitLabel}
        </button>
      </div>
    </form>
  );
}

function MoveList({
  copy,
  lang,
  states,
  title,
  empty,
  path,
  revision,
  mode,
  onNavigate,
  onUnauthorized,
}: {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  title: string;
  empty: string;
  path: string;
  revision: string;
  mode: "uses" | "stock";
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
}) {
  const list = useWindow(revision, (after: SpareMove | null, signal) => {
    const params = new URLSearchParams();
    if (after) {
      params.set("after", after.at);
      params.set("afterId", String(after.id));
    }
    return request<Page<SpareMove>>(`${path}?${params}`, { signal });
  });

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  const none = !list.loading && list.items.length === 0;

  return (
    <section className="panel flush">
      <div className="panel-head">
        <div className="panel-title">
          <h2>{title}</h2>
        </div>
      </div>
      {list.error ? <p className="alert">{problemText(copy, list.error)}</p> : null}
      <div className="table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>{copy.when}</th>
              <th>{copy.person}</th>
              <th>{copy.event}</th>
              <th>{copy.quantity}</th>
              {mode === "uses" ? (
                <>
                  <th>{copy.device}</th>
                  <th>{copy.status}</th>
                </>
              ) : null}
              <th>{copy.optionalNote}</th>
            </tr>
          </thead>
          <tbody>
            {list.items.map((move) => (
              <tr key={move.id}>
                <td data-label={copy.when}>{formatWhen(move.at, lang)}</td>
                <td data-label={copy.person}>{move.actorUsername}</td>
                <td data-label={copy.event}>{moveKind(copy, move.kind)}</td>
                <td data-label={copy.quantity}>{move.quantity}</td>
                {mode === "uses" ? (
                  <>
                    <td data-label={copy.device}>
                      <button
                        className="link"
                        type="button"
                        onClick={() => onNavigate(`/services/${move.serviceId}`)}
                      >
                        {move.deviceBrand} {move.deviceModel} · {move.serial}
                      </button>
                    </td>
                    <td data-label={copy.status}>
                      {move.status ? (
                        <span className="badge" data-tone={statusTone(move.status)}>
                          {statusLabel(move.status, states, lang)}
                        </span>
                      ) : null}
                    </td>
                  </>
                ) : null}
                <td data-label={copy.optionalNote}>{move.note}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {none ? <p className="muted empty">{empty}</p> : null}
        {list.loading ? <p className="muted empty">{copy.loading}</p> : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </div>
    </section>
  );
}
