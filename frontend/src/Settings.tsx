import { useEffect, useRef, useState, type DragEvent, type FormEvent } from "react";
import { ApiError, request, type ServiceState } from "./api";
import type { Copy, Lang } from "./i18n";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  onStates: (states: ServiceState[]) => void;
  onUnauthorized: () => void;
};

export function Settings({ copy, lang, states, onStates, onUnauthorized }: Props) {
  const [error, setError] = useState<string | null>(null);
  const [nameEn, setNameEn] = useState("");
  const [nameTr, setNameTr] = useState("");
  const [showOnHome, setShowOnHome] = useState(false);
  const [dragCode, setDragCode] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const order = useRef(states);
  const drag = useRef<{ code: string; start: number } | null>(null);
  if (!drag.current) {
    order.current = states;
  }

  function fail(err: unknown) {
    if (err instanceof ApiError && err.code === "unauthorized") {
      onUnauthorized();
      return;
    }
    setError(problemText(copy, err));
  }

  async function reload() {
    const next = await request<ServiceState[]>("/api/states");
    onStates(next);
  }

  async function create(event: FormEvent) {
    event.preventDefault();
    const english = nameEn.trim();
    const turkish = nameTr.trim();
    if (!english || !turkish) {
      setError(!english ? copy.englishNameRequired : copy.turkishNameRequired);
      return;
    }
    if ([...english].length > 80 || [...turkish].length > 80) {
      setError(copy.tooLong.replace("{field}", [...english].length > 80 ? copy.englishName : copy.turkishName));
      return;
    }
    setError(null);
    try {
      await request("/api/states", {
        method: "POST",
        body: JSON.stringify({ nameEn: english, nameTr: turkish, showOnHome }),
      });
      setNameEn("");
      setNameTr("");
      setShowOnHome(false);
      await reload();
    } catch (err) {
      fail(err);
    }
  }

  async function patch(code: string, body: Record<string, unknown>) {
    setError(null);
    try {
      await request(`/api/states/${encodeURIComponent(code)}`, {
        method: "PATCH",
        body: JSON.stringify(body),
      });
      await reload();
      return true;
    } catch (err) {
      fail(err);
      return false;
    }
  }

  function grab(code: string) {
    if (busy) {
      return;
    }
    const start = order.current.findIndex((item) => item.code === code);
    if (start < 0) {
      return;
    }
    drag.current = { code, start };
    setDragCode(code);
  }

  function moveToIndex(to: number) {
    const current = drag.current;
    if (!current) {
      return;
    }
    const from = order.current.findIndex((item) => item.code === current.code);
    if (from < 0 || to < 0 || to >= order.current.length || from === to) {
      return;
    }
    const next = [...order.current];
    const [item] = next.splice(from, 1);
    next.splice(to, 0, item);
    order.current = next;
    onStates(next);
  }

  function moveTo(targetCode: string) {
    moveToIndex(order.current.findIndex((item) => item.code === targetCode));
  }

  function slide(clientY: number) {
    const current = drag.current;
    if (!current) {
      return;
    }
    const cards = document.querySelectorAll<HTMLElement>("[data-state]");
    let to = 0;
    cards.forEach((card) => {
      if (card.dataset.state === current.code) {
        return;
      }
      const rect = card.getBoundingClientRect();
      if (clientY > rect.top + rect.height / 2) {
        to += 1;
      }
    });
    if (to > order.current.length - 1) {
      to = order.current.length - 1;
    }
    moveToIndex(to);
  }

  function release() {
    const current = drag.current;
    drag.current = null;
    setDragCode(null);
    if (!current) {
      return;
    }
    const end = order.current.findIndex((item) => item.code === current.code);
    if (end < 0 || end === current.start) {
      return;
    }
    setBusy(true);
    setError(null);
    void request(`/api/states/${encodeURIComponent(current.code)}/move`, {
      method: "POST",
      body: JSON.stringify({ position: end + 1 }),
    })
      .then(() => reload())
      .catch((err: unknown) => {
        fail(err);
        return reload();
      })
      .finally(() => setBusy(false));
  }

  return (
    <>
      <header className="page-head">
        <h1>{copy.settings}</h1>
        <p className="muted">{copy.settingsLead}</p>
      </header>
      <section className="stack">
        <header>
          <h2>{copy.serviceStates}</h2>
          <p className="muted">{copy.serviceStatesLead}</p>
        </header>
        {error ? <p className="alert">{error}</p> : null}
        <div className="state-list">
          {states.map((state) => (
            <StateCard
              key={state.code}
              copy={copy}
              lang={lang}
              state={state}
              dragging={dragCode === state.code}
              onPatch={patch}
              onGrab={() => grab(state.code)}
              onSlide={slide}
              onMoveTo={moveTo}
              onRelease={release}
            />
          ))}
        </div>
      </section>
      <form className="panel stack" onSubmit={(event) => void create(event)}>
        <h2>{copy.addState}</h2>
        <div className="state-add">
          <label className="field">
            {copy.englishName}
            <input value={nameEn} onChange={(event) => setNameEn(event.target.value)} />
          </label>
          <label className="field">
            {copy.turkishName}
            <input value={nameTr} onChange={(event) => setNameTr(event.target.value)} />
          </label>
          <label className="check">
            <input type="checkbox" checked={showOnHome} onChange={(event) => setShowOnHome(event.target.checked)} />
            {copy.showOnHome}
          </label>
          <button className="button" type="submit">
            {copy.addState}
          </button>
        </div>
      </form>
    </>
  );
}

function StateCard({
  copy,
  lang,
  state,
  dragging,
  onPatch,
  onGrab,
  onSlide,
  onMoveTo,
  onRelease,
}: {
  copy: Copy;
  lang: Lang;
  state: ServiceState;
  dragging: boolean;
  onPatch: (code: string, body: Record<string, unknown>) => Promise<boolean>;
  onGrab: () => void;
  onSlide: (clientY: number) => void;
  onMoveTo: (code: string) => void;
  onRelease: () => void;
}) {
  const [nameEn, setNameEn] = useState(state.nameEn);
  const [nameTr, setNameTr] = useState(state.nameTr);

  useEffect(() => {
    setNameEn(state.nameEn);
    setNameTr(state.nameTr);
  }, [state.nameEn, state.nameTr]);
  const primary = lang === "tr" ? state.nameTr : state.nameEn;
  const secondary = lang === "tr" ? state.nameEn : state.nameTr;
  const note =
    state.code === "registered"
      ? copy.registrationUsesState
      : state.code === "delivered"
        ? copy.deliveryUsesState
        : state.code === "cancelled"
          ? copy.cancelUsesState
          : "";

  function dragStart(event: DragEvent<HTMLButtonElement>) {
    event.dataTransfer.setData("text/plain", state.code);
    event.dataTransfer.effectAllowed = "move";
    const card = event.currentTarget.closest("article");
    if (card) {
      event.dataTransfer.setDragImage(card, 24, 24);
    }
    onGrab();
  }

  function dragOver(event: DragEvent<HTMLElement>) {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    onSlide(event.clientY);
  }

  function drop(event: DragEvent<HTMLElement>) {
    event.preventDefault();
    onMoveTo(state.code);
  }

  return (
    <article
      className={dragging ? "state-card dragging" : state.active ? "state-card" : "state-card off"}
      data-state={state.code}
      onDragOver={dragOver}
      onDrop={drop}
    >
      <button className="drag-handle" type="button" draggable aria-label={copy.dragState} onDragStart={dragStart} onDragEnd={onRelease}>
        <Grip />
      </button>
      <div className="state-body">
        {state.fixed ? (
          <>
            <p className="state-title">{primary}</p>
            <p className="muted">{secondary}</p>
          </>
        ) : (
          <form
            className="state-names"
            onSubmit={(event) => {
              event.preventDefault();
              void onPatch(state.code, { nameEn: nameEn.trim(), nameTr: nameTr.trim() });
            }}
          >
            <label className="field">
              {copy.englishName}
              <input value={nameEn} onChange={(event) => setNameEn(event.target.value)} />
            </label>
            <label className="field">
              {copy.turkishName}
              <input value={nameTr} onChange={(event) => setNameTr(event.target.value)} />
            </label>
            <button className="button secondary" type="submit">
              {copy.save}
            </button>
          </form>
        )}
      </div>
      <div className="state-flags">
        {state.locked ? (
          <p className="muted state-note">{note}</p>
        ) : (
          <>
            <label className="check">
              <input
                type="checkbox"
                checked={state.active}
                onChange={(event) => void onPatch(state.code, { active: event.target.checked })}
              />
              {state.active ? copy.stateActive : copy.stateInactive}
            </label>
            <label className="check">
              <input
                type="checkbox"
                checked={state.showOnHome}
                onChange={(event) => void onPatch(state.code, { showOnHome: event.target.checked })}
              />
              {copy.showOnHome}
            </label>
          </>
        )}
      </div>
    </article>
  );
}

function Grip() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="9" cy="6" r="1.4" fill="currentColor" />
      <circle cx="15" cy="6" r="1.4" fill="currentColor" />
      <circle cx="9" cy="12" r="1.4" fill="currentColor" />
      <circle cx="15" cy="12" r="1.4" fill="currentColor" />
      <circle cx="9" cy="18" r="1.4" fill="currentColor" />
      <circle cx="15" cy="18" r="1.4" fill="currentColor" />
    </svg>
  );
}
