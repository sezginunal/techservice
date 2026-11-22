import { useEffect, useId, useRef, useState, type FormEvent, type RefObject } from "react";
import { createPortal } from "react-dom";
import { ApiError, request, type Page, type SparePart, type SpareReturnable } from "./api";
import { wholeQuantity } from "./fields";
import type { Copy } from "./i18n";
import { End, useQuery, useWindow } from "./list";
import { problemText } from "./problems";
import { rangeText, spareProblem } from "./spare";

type Props = {
  copy: Copy;
  serviceId: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

type Draft = {
  part: SparePart;
  quantity: string;
};

export function SpareService({ copy, serviceId, onNavigate, onUnauthorized }: Props) {
  const [revision, setRevision] = useState(0);
  const [useOpen, setUseOpen] = useState(false);
  const [returning, setReturning] = useState<SpareReturnable | null>(null);
  const onHand = useWindow(`${serviceId}:${revision}`, (after: SpareReturnable | null, signal) => {
    const params = new URLSearchParams();
    if (after) {
      params.set("after", after.brand);
      params.set("afterModel", after.model);
      params.set("afterId", after.id);
    }
    return request<Page<SpareReturnable>>(`/api/services/${serviceId}/spare-returns?${params}`, { signal });
  });

  useEffect(() => {
    setUseOpen(false);
    setReturning(null);
  }, [serviceId]);

  useEffect(() => {
    if (onHand.error instanceof ApiError && onHand.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [onHand.error, onUnauthorized]);

  function saved() {
    setUseOpen(false);
    setReturning(null);
    setRevision((value) => value + 1);
  }

  const none = !onHand.loading && onHand.items.length === 0;

  return (
    <section className="panel">
      <div className="panel-head">
        <div className="panel-title">
          <h2>{copy.spareParts}</h2>
        </div>
        <div className="panel-actions">
          <button
            className="button secondary small"
            type="button"
            aria-expanded={useOpen}
            onClick={() => {
              setReturning(null);
              setUseOpen(true);
            }}
          >
            {copy.useSparePart}
          </button>
        </div>
      </div>
      {onHand.error ? <p className="alert">{problemText(copy, onHand.error)}</p> : null}
      {onHand.items.length > 0 ? (
        <ul className="moves">
          {onHand.items.map((part) => (
            <li className="move" key={part.id}>
              <div className="move-line">
                <button className="link" type="button" onClick={() => onNavigate(`/spare-parts/${part.id}`)}>
                  {part.brand} {part.model}
                </button>
                <span className="choice-tag">{copy.quantity}</span>
                <span className="qty">{part.returnable}</span>
                <button
                  className="button secondary small move-aside"
                  type="button"
                  onClick={() => {
                    setUseOpen(false);
                    setReturning(part);
                  }}
                >
                  {copy.returnToStock}
                </button>
              </div>
            </li>
          ))}
        </ul>
      ) : null}
      {none ? <p className="muted">{copy.noServiceSpares}</p> : null}
      {onHand.loading ? (
        <div className="loading">
          <span className="spinner" />
          {copy.loading}
        </div>
      ) : null}
      <End enabled={onHand.hasMore && !onHand.loading} onEnd={onHand.more} />
      {useOpen ? (
        <UseDialog copy={copy} serviceId={serviceId} onUnauthorized={onUnauthorized} onClose={() => setUseOpen(false)} onSaved={saved} />
      ) : null}
      {returning ? (
        <ReturnDialog
          copy={copy}
          serviceId={serviceId}
          part={returning}
          onUnauthorized={onUnauthorized}
          onClose={() => setReturning(null)}
          onSaved={saved}
        />
      ) : null}
    </section>
  );
}

function UseDialog({
  copy,
  serviceId,
  onUnauthorized,
  onClose,
  onSaved,
}: {
  copy: Copy;
  serviceId: string;
  onUnauthorized: () => void;
  onClose: () => void;
  onSaved: () => void;
}) {
  const titleId = useId();
  const search = useQuery();
  const choices = useRef<HTMLDivElement>(null);
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
  const [lines, setLines] = useState<Draft[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const chosen = new Set(lines.map((line) => line.part.id));

  useEffect(() => {
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  function add(part: SparePart) {
    setLines((current) => (current.some((line) => line.part.id === part.id) ? current : [...current, { part, quantity: "1" }]));
    setError(null);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    const ready: { partId: string; quantity: number }[] = [];
    for (const line of lines) {
      const amount = wholeQuantity(line.quantity, 1, line.part.onHand);
      if (typeof amount !== "number") {
        const text = amount === "empty" ? copy.quantityRequired : rangeText(copy, 1, line.part.onHand);
        setError(`${line.part.brand} ${line.part.model}: ${text}`);
        return;
      }
      ready.push({ partId: line.part.id, quantity: amount });
    }
    setSaving(true);
    setError(null);
    try {
      await request(`/api/services/${serviceId}/spare-parts`, {
        method: "POST",
        body: JSON.stringify({ lines: ready }),
      });
      onSaved();
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      if (err instanceof ApiError && err.partId) {
        setLines((current) =>
          current.map((line) =>
            line.part.id === err.partId && err.code === "stock_short" ? { ...line, part: { ...line.part, onHand: err.onHand } } : line,
          ),
        );
      }
      const named = err instanceof ApiError && err.brand ? `${err.brand} ${err.model}: ` : "";
      const line = lines.find((item) => err instanceof ApiError && item.part.id === err.partId);
      const max = err instanceof ApiError && err.code === "stock_short" ? err.onHand : (line?.part.onHand ?? 1);
      setError(named + (spareProblem(copy, err, 1, Math.max(1, max)) ?? problemText(copy, err)));
    } finally {
      setSaving(false);
    }
  }

  const visible = list.items.filter((part) => !chosen.has(part.id));
  const empty = !list.loading && list.items.length === 0;

  return createPortal(
    <div
      className="dialog-scrim"
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <form className="panel dialog stack" onSubmit={(event) => void submit(event)}>
        <div className="panel-head">
          <div className="panel-title">
            <h2 id={titleId}>{copy.useSparePart}</h2>
          </div>
        </div>
        {lines.map((line) => (
          <div className="move-line" key={line.part.id}>
            <span className="choice-name">
              {line.part.brand} {line.part.model}
            </span>
            <span className="choice-tag">
              {copy.inStock} {line.part.onHand}
            </span>
            <label className="field qty-field">
              {copy.quantity}
              <input
                inputMode="numeric"
                value={line.quantity}
                onChange={(event) => {
                  const quantity = event.target.value;
                  setLines((current) => current.map((item) => (item.part.id === line.part.id ? { ...item, quantity } : item)));
                }}
              />
            </label>
            <button
              className="button ghost small"
              type="button"
              onClick={() => setLines((current) => current.filter((item) => item.part.id !== line.part.id))}
            >
              {copy.removeLine}
            </button>
          </div>
        ))}
        <label className="field">
          <span className="sr-only">{copy.search}</span>
          <input
            value={search.value}
            placeholder={copy.spareSearch}
            onChange={(event) => search.setValue(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                search.commit();
              }
            }}
          />
        </label>
        {list.error ? <p className="alert">{problemText(copy, list.error)}</p> : null}
        <div className="choice-list dialog-choices" ref={choices}>
          {visible.map((part) => (
            <button className="choice" type="button" key={part.id} disabled={part.onHand === 0} onClick={() => add(part)}>
              <span className="choice-name">
                {part.brand} {part.model}
              </span>
              <span className="choice-tag">
                {copy.inStock} {part.onHand}
              </span>
            </button>
          ))}
          <DialogMore root={choices} enabled={list.hasMore && !list.loading} onEnd={list.more} />
        </div>
        {empty ? <p className="muted">{search.query ? copy.noSpareMatches : copy.noSpareParts}</p> : null}
        {list.loading ? <p className="muted">{copy.loading}</p> : null}
        {error ? <p className="alert">{error}</p> : null}
        <div className="dialog-actions">
          <button className="button secondary" type="button" onClick={onClose}>
            {copy.cancel}
          </button>
          <button className="button" type="submit" disabled={lines.length === 0 || saving}>
            {copy.save}
          </button>
        </div>
      </form>
    </div>,
    document.body,
  );
}

function ReturnDialog({
  copy,
  serviceId,
  part,
  onUnauthorized,
  onClose,
  onSaved,
}: {
  copy: Copy;
  serviceId: string;
  part: SpareReturnable;
  onUnauthorized: () => void;
  onClose: () => void;
  onSaved: () => void;
}) {
  const titleId = useId();
  const [quantity, setQuantity] = useState("1");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const amount = wholeQuantity(quantity, 1, part.returnable);

  useEffect(() => {
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (typeof amount !== "number") {
      setError(amount === "empty" ? copy.quantityRequired : rangeText(copy, 1, part.returnable));
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await request(`/api/services/${serviceId}/spare-returns`, {
        method: "POST",
        body: JSON.stringify({ partId: part.id, quantity: amount, note: "" }),
      });
      onSaved();
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setError(spareProblem(copy, err, 1, part.returnable) ?? problemText(copy, err));
    } finally {
      setSaving(false);
    }
  }

  return createPortal(
    <div
      className="dialog-scrim"
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <form className="panel dialog stack" onSubmit={(event) => void submit(event)}>
        <div className="panel-head">
          <div className="panel-title">
            <h2 id={titleId}>{copy.returnToStock}</h2>
          </div>
        </div>
        <p>
          {part.brand} {part.model}
        </p>
        <label className="field">
          {copy.quantity}
          <input inputMode="numeric" value={quantity} onChange={(event) => setQuantity(event.target.value)} />
        </label>
        <p>{typeof amount === "number" ? copy.returnSure.replace("{n}", String(amount)) : copy.returnPrompt}</p>
        {error ? <p className="alert">{error}</p> : null}
        <div className="dialog-actions">
          <button className="button secondary" type="button" onClick={onClose}>
            {copy.cancel}
          </button>
          <button className="button" type="submit" disabled={saving}>
            {copy.returnToStock}
          </button>
        </div>
      </form>
    </div>,
    document.body,
  );
}

function DialogMore({ root, enabled, onEnd }: { root: RefObject<HTMLDivElement | null>; enabled: boolean; onEnd: () => void }) {
  const ref = useRef<HTMLDivElement>(null);
  const onEndRef = useRef(onEnd);
  onEndRef.current = onEnd;

  useEffect(() => {
    const node = ref.current;
    const box = root.current;
    if (!enabled || !node || !box) {
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          onEndRef.current();
        }
      },
      { root: box, rootMargin: "80px" },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [enabled, root]);

  return <div ref={ref} />;
}
