import { useEffect, useState, type FormEvent } from "react";
import {
  ApiError,
  request,
  type Page,
  type ServiceState,
  type SpareMove,
  type SparePart,
  type SpareReturnable,
} from "./api";
import { wholeQuantity } from "./fields";
import { formatWhen, statusLabel, statusTone } from "./format";
import type { Copy, Lang } from "./i18n";
import { End, useQuery, useWindow } from "./list";
import { problemText } from "./problems";
import { moveKind, noteTooLong, rangeText, spareProblem } from "./spare";

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  serviceId: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function SpareService({ copy, lang, states, serviceId, onNavigate, onUnauthorized }: Props) {
  const [revision, setRevision] = useState(0);
  const [picker, setPicker] = useState(0);
  const [returnOpen, setReturnOpen] = useState(false);
  const [canReturn, setCanReturn] = useState(false);
  const events = useWindow(`${serviceId}:${revision}`, (after: SpareMove | null, signal) => {
    const params = new URLSearchParams();
    if (after) {
      params.set("after", after.at);
      params.set("afterId", String(after.id));
    }
    return request<Page<SpareMove>>(`/api/services/${serviceId}/spare-parts?${params}`, { signal });
  });

  useEffect(() => {
    if (events.error instanceof ApiError && events.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [events.error, onUnauthorized]);

  useEffect(() => {
    const controller = new AbortController();
    request<Page<SpareReturnable>>(`/api/services/${serviceId}/spare-returns?limit=1`, { signal: controller.signal })
      .then((page) => {
        if (!controller.signal.aborted) {
          setCanReturn(page.items.length > 0);
        }
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        setCanReturn(false);
        if (err instanceof ApiError && err.code === "unauthorized") {
          onUnauthorized();
        }
      });
    return () => controller.abort();
  }, [onUnauthorized, revision, serviceId]);

  function changed() {
    setRevision((value) => value + 1);
    setPicker(0);
    setReturnOpen(false);
  }

  const none = !events.loading && events.items.length === 0;

  return (
    <section className="panel">
      <div className="panel-head">
        <div className="panel-title">
          <h2>{copy.spareParts}</h2>
        </div>
        <div className="panel-actions">
          <button className="button secondary small" type="button" onClick={() => setPicker((value) => (value === 0 ? revision + 1 : 0))}>
            {copy.useSparePart}
          </button>
        </div>
      </div>
      {events.error ? <p className="alert">{problemText(copy, events.error)}</p> : null}
      {events.items.length > 0 ? (
        <ul className="moves">
          {events.items.map((move) => (
            <li className="move" key={move.id}>
              <div className="move-line">
                <button className="link" type="button" onClick={() => onNavigate(`/spare-parts/${move.partId}`)}>
                  {move.brand} {move.model}
                </button>
                <span className="qty">{move.quantity}</span>
                <span className="badge" data-tone="gray">
                  <span>{moveKind(copy, move.kind)}</span>
                </span>
                {move.status ? (
                  <span className="badge" data-tone={statusTone(move.status)}>
                    <span>{statusLabel(move.status, states, lang)}</span>
                  </span>
                ) : null}
              </div>
              <p className="meta">
                <span className="meta-item">{formatWhen(move.at, lang)}</span>
                <span className="meta-item">{move.actorUsername}</span>
              </p>
              {move.note ? <p className="notes">{move.note}</p> : null}
            </li>
          ))}
        </ul>
      ) : null}
      {none ? <p className="muted">{copy.noServiceSpares}</p> : null}
      {events.loading ? (
        <div className="loading">
          <span className="spinner" />
          {copy.loading}
        </div>
      ) : null}
      <End enabled={events.hasMore && !events.loading} onEnd={events.more} />
      {picker ? (
        <SparePicker
          key={picker}
          copy={copy}
          onUnauthorized={onUnauthorized}
          onUsed={changed}
          save={(partId, quantity) =>
            request(`/api/services/${serviceId}/spare-parts`, {
              method: "POST",
              body: JSON.stringify({ partId, quantity }),
            })
          }
        />
      ) : null}
      {canReturn ? (
        <button className="button secondary" type="button" onClick={() => setReturnOpen((value) => !value)}>
          {copy.returnToStock}
        </button>
      ) : null}
      {returnOpen && canReturn ? (
        <ReturnForm
          key={`${serviceId}:${revision}`}
          copy={copy}
          serviceId={serviceId}
          onUnauthorized={onUnauthorized}
          onSaved={changed}
        />
      ) : null}
    </section>
  );
}

function SparePicker({
  copy,
  onUnauthorized,
  onUsed,
  save,
}: {
  copy: Copy;
  onUnauthorized: () => void;
  onUsed: () => void;
  save: (partId: string, quantity: number) => Promise<unknown>;
}) {
  const search = useQuery();
  const [reload, setReload] = useState(0);
  const list = useWindow(`${search.query}:${reload}`, (after: SparePart | null, signal) => {
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
  const [selected, setSelected] = useState<SparePart | null>(null);
  const [quantity, setQuantity] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!selected || selected.onHand < 1) {
      return;
    }
    const amount = wholeQuantity(quantity, 1, selected.onHand);
    if (typeof amount !== "number") {
      setError(amount === "empty" ? copy.quantityRequired : rangeText(copy, 1, selected.onHand));
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await save(selected.id, amount);
      onUsed();
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setError(spareProblem(copy, err, 1, selected.onHand) ?? problemText(copy, err));
      setSelected(null);
      setReload((value) => value + 1);
    } finally {
      setSaving(false);
    }
  }

  const empty = !list.loading && list.items.length === 0;

  return (
    <form className="stack" onSubmit={(event) => void submit(event)}>
      <label className="field">
        {copy.search}
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
      {list.items.map((part) => (
        <button
          className="button secondary"
          type="button"
          key={part.id}
          disabled={part.onHand === 0}
          aria-pressed={selected?.id === part.id}
          onClick={() => {
            setSelected(part);
            setQuantity("");
            setError(null);
          }}
        >
          {part.brand} {part.model} · {copy.inStock} {part.onHand}
        </button>
      ))}
      {empty ? <p className="muted">{search.query ? copy.noSpareMatches : copy.noSpareParts}</p> : null}
      {list.loading ? <p className="muted">{copy.loading}</p> : null}
      <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      {selected ? (
        <label className="field">
          {copy.quantity}
          <input inputMode="numeric" value={quantity} onChange={(event) => setQuantity(event.target.value)} />
        </label>
      ) : null}
      {error ? <p className="alert">{error}</p> : null}
      <button className="button" type="submit" disabled={!selected || saving}>
        {copy.save}
      </button>
    </form>
  );
}

function ReturnForm({
  copy,
  serviceId,
  onUnauthorized,
  onSaved,
}: {
  copy: Copy;
  serviceId: string;
  onUnauthorized: () => void;
  onSaved: () => void;
}) {
  const [reload, setReload] = useState(0);
  const list = useWindow(String(reload), (after: SpareReturnable | null, signal) => {
    const params = new URLSearchParams();
    if (after) {
      params.set("after", after.brand);
      params.set("afterModel", after.model);
      params.set("afterId", after.id);
    }
    return request<Page<SpareReturnable>>(`/api/services/${serviceId}/spare-returns?${params}`, { signal });
  });
  const [selected, setSelected] = useState<SpareReturnable | null>(null);
  const [quantity, setQuantity] = useState("");
  const [note, setNote] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!selected) {
      return;
    }
    const amount = wholeQuantity(quantity, 1, selected.returnable);
    if (typeof amount !== "number") {
      setError(amount === "empty" ? copy.quantityRequired : rangeText(copy, 1, selected.returnable));
      return;
    }
    if (noteTooLong(note)) {
      setError(problemText(copy, new ApiError("validation", { tooLong: ["spareNote"] })));
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await request(`/api/services/${serviceId}/spare-returns`, {
        method: "POST",
        body: JSON.stringify({ partId: selected.id, quantity: amount, note: note.trim() }),
      });
      onSaved();
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setError(spareProblem(copy, err, 1, selected.returnable) ?? problemText(copy, err));
      setSelected(null);
      setReload((value) => value + 1);
    } finally {
      setSaving(false);
    }
  }

  return (
    <form className="stack" onSubmit={(event) => void submit(event)}>
      {list.error ? <p className="alert">{problemText(copy, list.error)}</p> : null}
      {list.items.map((part) => (
        <button
          className="button secondary"
          type="button"
          key={part.id}
          aria-pressed={selected?.id === part.id}
          onClick={() => {
            setSelected(part);
            setQuantity("");
            setError(null);
          }}
        >
          {part.brand} {part.model} · {copy.returnable} {part.returnable}
        </button>
      ))}
      {list.loading ? <p className="muted">{copy.loading}</p> : null}
      <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      {selected ? (
        <>
          <label className="field">
            {copy.quantity}
            <input inputMode="numeric" value={quantity} onChange={(event) => setQuantity(event.target.value)} />
          </label>
          <label className="field">
            {copy.optionalNote}
            <textarea value={note} onChange={(event) => setNote(event.target.value)} />
          </label>
        </>
      ) : null}
      {error ? <p className="alert">{error}</p> : null}
      <button className="button" type="submit" disabled={!selected || saving}>
        {copy.returnToStock}
      </button>
    </form>
  );
}
