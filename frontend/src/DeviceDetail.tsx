import { useEffect, useState, type FormEvent } from "react";
import {
  ApiError,
  photoSrc,
  request,
  statusPhotoSrc,
  type Device,
  type ServiceComment,
  type ServiceState,
  type User,
} from "./api";
import { ModelIcon } from "./ModelIcon";
import { choosableStates, formatWhen, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { PhotoButton } from "./PhotoView";
import { fileProblem, problemText } from "./problems";
import { SpareService } from "./SpareService";

const maxCommentLen = 4000;

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  user: User;
  id: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function DeviceDetail({ copy, lang, states, user, id, onNavigate, onUnauthorized }: Props) {
  const [device, setDevice] = useState<Device | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [next, setNext] = useState("");
  const [note, setNote] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [fileKey, setFileKey] = useState(0);
  const [statusError, setStatusError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [draft, setDraft] = useState("");
  const [commentError, setCommentError] = useState<string | null>(null);
  const [commentSaving, setCommentSaving] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editText, setEditText] = useState("");
  const [editError, setEditError] = useState<string | null>(null);
  const [editSaving, setEditSaving] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    request<Device>(`/api/services/${id}`, { signal: controller.signal })
      .then(setDevice)
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

  async function saveStatus(event: FormEvent) {
    event.preventDefault();
    if (!device) {
      return;
    }
    const missing = next ? [] : ["status"];
    const photo = file ? fileProblem(file) : null;
    if (missing.length || photo) {
      setStatusError(problemText(copy, new ApiError("validation", { missing, photo: photo ?? "" })));
      return;
    }
    const body = new FormData();
    body.set("payload", JSON.stringify({ status: next, note }));
    if (file) {
      body.set("file0", file);
    }
    setSaving(true);
    setStatusError(null);
    try {
      const saved = await request<Device>(`/api/services/${device.id}/status`, { method: "POST", body });
      setDevice(saved);
      setNext("");
      setNote("");
      setFile(null);
      setFileKey((value) => value + 1);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setStatusError(problemText(copy, err));
    } finally {
      setSaving(false);
    }
  }

  async function saveComment(event: FormEvent) {
    event.preventDefault();
    if (!device) {
      return;
    }
    const text = draft.trim();
    if (!text) {
      setCommentError(problemText(copy, new ApiError("validation", { missing: ["comment"] })));
      return;
    }
    if ([...text].length > maxCommentLen) {
      setCommentError(problemText(copy, new ApiError("validation", { tooLong: ["comment"] })));
      return;
    }
    setCommentSaving(true);
    setCommentError(null);
    try {
      const saved = await request<ServiceComment>(`/api/services/${device.id}/comments`, {
        method: "POST",
        body: JSON.stringify({ text: draft }),
      });
      setDevice({ ...device, comments: [...device.comments, saved] });
      setDraft("");
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setCommentError(problemText(copy, err));
    } finally {
      setCommentSaving(false);
    }
  }

  function startEdit(comment: ServiceComment) {
    setEditingId(comment.id);
    setEditText(comment.text);
    setEditError(null);
  }

  function cancelEdit() {
    setEditingId(null);
    setEditText("");
    setEditError(null);
  }

  async function saveEdit(event: FormEvent) {
    event.preventDefault();
    if (!device || editingId == null) {
      return;
    }
    const text = editText.trim();
    if (!text) {
      setEditError(problemText(copy, new ApiError("validation", { missing: ["comment"] })));
      return;
    }
    if ([...text].length > maxCommentLen) {
      setEditError(problemText(copy, new ApiError("validation", { tooLong: ["comment"] })));
      return;
    }
    setEditSaving(true);
    setEditError(null);
    try {
      const saved = await request<ServiceComment>(`/api/services/${device.id}/comments/${editingId}`, {
        method: "PATCH",
        body: JSON.stringify({ text: editText }),
      });
      setDevice({
        ...device,
        comments: device.comments.map((item) => (item.id === saved.id ? saved : item)),
      });
      cancelEdit();
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setEditError(problemText(copy, err));
    } finally {
      setEditSaving(false);
    }
  }

  if (!device) {
    return <p className={error ? "alert" : "muted"}>{error ?? copy.loading}</p>;
  }

  return (
    <>
      <header className="page-head row">
        <div className="stack">
          <h1 className="with-icon">
            {device.image ? <ModelIcon brand={device.brand} model={device.model} /> : null}
            <button className="button linkish" type="button" onClick={() => onNavigate(`/devices/${device.deviceId}`)}>
              {device.brand} {device.model} · {device.serial}
            </button>
          </h1>
          <p className="uuid">{device.id}</p>
          <p className="muted">
            {copy.registeredOn}: {formatWhen(device.createdAt, lang)}
          </p>
          <span className="badge">{statusLabel(device.status, states, lang)}</span>
        </div>
        <button className="button" type="button" onClick={() => onNavigate(`/services/${device.id}/edit`)}>
          {copy.correct}
        </button>
      </header>
      <section className="panel stack">
        <h2>{copy.customer}</h2>
        <p>{device.customer.name}</p>
        <p>{device.customer.phone}</p>
        <p>{device.customer.email}</p>
        <p>{device.customer.address}</p>
      </section>
      <section className="panel stack">
        <h2>{copy.device}</h2>
        <p>{device.note}</p>
        <PhotoRow copy={copy} deviceId={device.id} photos={device.photos} onUnauthorized={onUnauthorized} />
      </section>
      <section className="panel stack">
        <h2>{copy.parts}</h2>
        {device.parts.length === 0 ? <p className="muted">{copy.noParts}</p> : null}
        {device.parts.map((part) => (
          <div className="stack part" key={part.id}>
            <h2>{part.name}</h2>
            <p>{part.note}</p>
            <PhotoRow copy={copy} deviceId={device.id} photos={part.photos} onUnauthorized={onUnauthorized} />
          </div>
        ))}
      </section>
      <SpareService
        copy={copy}
        lang={lang}
        states={states}
        serviceId={device.id}
        onNavigate={onNavigate}
        onUnauthorized={onUnauthorized}
      />
      <section className="panel stack">
        <h2>{copy.comments}</h2>
        {device.comments.length === 0 ? <p className="muted">{copy.noComments}</p> : null}
        {device.comments.map((comment) => (
          <div className="stack part" key={comment.id}>
            {editingId === comment.id ? (
              <form className="stack" onSubmit={(event) => void saveEdit(event)}>
                <label className="field">
                  {copy.comment}
                  <textarea value={editText} onChange={(event) => setEditText(event.target.value)} />
                </label>
                {editError ? <p className="alert">{editError}</p> : null}
                <div className="row">
                  <button className="button" type="submit" disabled={editSaving}>
                    {copy.save}
                  </button>
                  <button className="button" type="button" onClick={cancelEdit} disabled={editSaving}>
                    {copy.cancel}
                  </button>
                </div>
              </form>
            ) : (
              <>
                <p>
                  {formatWhen(comment.createdAt, lang)} · {comment.authorUsername}
                  {comment.edited ? (
                    <>
                      {" "}
                      · {copy.edited} {formatWhen(comment.updatedAt, lang)}
                    </>
                  ) : null}
                </p>
                <p>{comment.text}</p>
                {comment.authorId === user.id ? (
                  <button className="button linkish" type="button" onClick={() => startEdit(comment)}>
                    {copy.edit}
                  </button>
                ) : null}
              </>
            )}
          </div>
        ))}
        <form className="stack" onSubmit={(event) => void saveComment(event)}>
          <label className="field">
            {copy.comment}
            <textarea value={draft} onChange={(event) => setDraft(event.target.value)} />
          </label>
          {commentError ? <p className="alert">{commentError}</p> : null}
          <button className="button" type="submit" disabled={commentSaving}>
            {copy.save}
          </button>
        </form>
      </section>
      <form className="panel stack" onSubmit={(event) => void saveStatus(event)}>
        <h2>{copy.changeStatus}</h2>
        <label className="field">
          {copy.status}
          <select value={next} onChange={(event) => setNext(event.target.value)}>
            <option value="">{copy.statusRequired}</option>
            {choosableStates(states).map((item) => (
              <option key={item.code} value={item.code}>
                {statusLabel(item.code, states, lang)}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          {copy.statusNote}
          <textarea value={note} onChange={(event) => setNote(event.target.value)} />
        </label>
        <label className="field">
          {copy.statusPhoto}
          <input
            key={fileKey}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            onChange={(event) => setFile(event.target.files?.[0] ?? null)}
          />
        </label>
        {statusError ? <p className="alert">{statusError}</p> : null}
        <button className="button" type="submit" disabled={saving}>
          {copy.save}
        </button>
      </form>
      <section className="panel stack">
        <h2>{copy.statusHistory}</h2>
        {device.history.length === 0 ? <p className="muted">{copy.noStatusHistory}</p> : null}
        {device.history.map((change) => (
          <div className="stack part" key={change.id}>
            <p>
              {formatWhen(change.createdAt, lang)} · {change.actorUsername}
            </p>
            <p>
              {statusLabel(change.from, states, lang)} → {statusLabel(change.to, states, lang)}
            </p>
            {change.note ? <p>{change.note}</p> : null}
            {change.hasPhoto ? (
              <PhotoButton
                copy={copy}
                className="thumb"
                src={statusPhotoSrc(device.id, change.id)}
                name={String(change.id)}
                onUnauthorized={onUnauthorized}
              />
            ) : null}
          </div>
        ))}
      </section>
    </>
  );
}

function PhotoRow({
  copy,
  deviceId,
  photos,
  onUnauthorized,
}: {
  copy: Copy;
  deviceId: string;
  photos: { id: string }[];
  onUnauthorized: () => void;
}) {
  if (photos.length === 0) {
    return null;
  }
  return (
    <div className="photos">
      {photos.map((photo) => (
        <PhotoButton
          key={photo.id}
          copy={copy}
          className="thumb"
          src={photoSrc(deviceId, photo.id)}
          name={photo.id}
          onUnauthorized={onUnauthorized}
        />
      ))}
    </div>
  );
}
