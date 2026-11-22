import { useEffect, useState, type FormEvent } from "react";
import { createPortal } from "react-dom";
import {
  ApiError,
  photoSrc,
  request,
  statusPhotoSrc,
  type Brand,
  type Device,
  type ServiceComment,
  type ServiceState,
  type SpareReturnable,
  type User,
} from "./api";
import { ModelIcon } from "./ModelIcon";
import { choosableStates, formatWhen, statusLabel, statusTone } from "./format";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
import { PhotoButton } from "./PhotoView";
import { fileProblem, problemText } from "./problems";
import { loadReportSpares, type ReportLayout } from "./report";
import { ServiceReport } from "./ServiceReport";
import { SpareService } from "./SpareService";
import { initials } from "./Shell";

const maxCommentLen = 4000;

type Props = {
  copy: Copy;
  lang: Lang;
  brand: Brand;
  states: ServiceState[];
  user: User;
  id: string;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function DeviceDetail({ copy, lang, brand, states, user, id, onNavigate, onUnauthorized }: Props) {
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
  const [reportLayout, setReportLayout] = useState<ReportLayout | null>(null);
  const [reportSpares, setReportSpares] = useState<SpareReturnable[] | null>(null);
  const [reportError, setReportError] = useState<string | null>(null);
  const [reportBusy, setReportBusy] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    request<Device>(`/api/services/${id}`, { signal: controller.signal })
      .then((loaded) => {
        if (!controller.signal.aborted) {
          setDevice(loaded);
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

  useEffect(() => {
    if (!reportLayout || !reportSpares) {
      return;
    }
    const timer = window.setTimeout(() => window.print(), 50);
    function clearReport() {
      setReportLayout(null);
      setReportSpares(null);
    }
    window.addEventListener("afterprint", clearReport);
    return () => {
      window.clearTimeout(timer);
      window.removeEventListener("afterprint", clearReport);
    };
  }, [reportLayout, reportSpares]);

  async function printReport(layout: ReportLayout) {
    if (!device || reportBusy) {
      return;
    }
    setReportBusy(true);
    setReportError(null);
    try {
      const spares = await loadReportSpares(device.id);
      setReportSpares(spares);
      setReportLayout(layout);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setReportError(problemText(copy, err));
    } finally {
      setReportBusy(false);
    }
  }

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
            <button className="link" type="button" onClick={() => onNavigate("/services")}>
              {copy.devices}
            </button>
            <Icon name="chevronRight" size={14} />
          </li>
          <li aria-current="page">
            {device.brand} {device.model}
          </li>
        </ol>
        <div className="page-title">
          <div className="page-title-main">
            <span className="page-media">
              {device.image ? <ModelIcon brand={device.brand} model={device.model} /> : <Icon name="device" size={20} />}
            </span>
            <div className="page-title-text">
              <h1>
                <button className="title-link" type="button" onClick={() => onNavigate(`/devices/${device.deviceId}`)}>
                  {device.brand} {device.model}
                </button>{" "}
                <span className="title-serial">{device.serial}</span>
              </h1>
              <div className="meta">
                <span className="badge lg" data-tone={statusTone(device.status)}>
                  <span>{statusLabel(device.status, states, lang)}</span>
                </span>
                <span className="meta-item">
                  <Icon name="calendar" size={14} />
                  {copy.registeredOn} {formatWhen(device.createdAt, lang)}
                </span>
                <span className="meta-item uuid" title={device.id}>
                  {device.id}
                </span>
              </div>
            </div>
          </div>
          <div className="page-actions">
            <button className="button secondary" type="button" disabled={reportBusy} onClick={() => void printReport("summary")}>
              <Icon name="printer" size={16} />
              {copy.printSummary}
            </button>
            <button className="button secondary" type="button" disabled={reportBusy} onClick={() => void printReport("detailed")}>
              <Icon name="printer" size={16} />
              {copy.printDetailed}
            </button>
            <button className="button secondary" type="button" onClick={() => onNavigate(`/services/${device.id}/edit`)}>
              <Icon name="pencil" size={16} />
              {copy.correct}
            </button>
          </div>
        </div>
        {reportError ? <p className="alert">{reportError}</p> : null}
      </header>

      <div className="detail">
        <div className="detail-main">
          <form className="panel" onSubmit={(event) => void saveStatus(event)}>
            <div className="panel-head">
              <div className="panel-title">
                <h2>
                  <Icon name="status" size={16} />
                  {copy.changeStatus}
                </h2>
              </div>
            </div>
            <div className="fields">
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
            </div>
            {statusError ? <p className="alert">{statusError}</p> : null}
            <div className="form-actions">
              <button className="button" type="submit" disabled={saving}>
                {copy.save}
              </button>
            </div>
          </form>

          <section className="panel">
            <div className="panel-head">
              <div className="panel-title">
                <h2>
                  <Icon name="message" size={16} />
                  {copy.comments}
                </h2>
              </div>
            </div>
            {device.comments.length === 0 ? (
              <div className="empty compact">
                <span className="empty-icon">
                  <Icon name="message" size={16} />
                </span>
                <span>{copy.noComments}</span>
              </div>
            ) : (
              <ul className="comments">
                {device.comments.map((comment) => (
                  <li className="comment" key={comment.id}>
                    <span className="avatar sm" data-tone="teal" aria-hidden="true">
                      {initials(comment.authorUsername)}
                    </span>
                    <div className="comment-body">
                      {editingId === comment.id ? (
                        <form className="fields" onSubmit={(event) => void saveEdit(event)}>
                          <label className="field">
                            {copy.comment}
                            <textarea value={editText} onChange={(event) => setEditText(event.target.value)} />
                          </label>
                          {editError ? <p className="alert">{editError}</p> : null}
                          <div className="form-actions">
                            <button className="button secondary" type="button" onClick={cancelEdit} disabled={editSaving}>
                              {copy.cancel}
                            </button>
                            <button className="button" type="submit" disabled={editSaving}>
                              {copy.save}
                            </button>
                          </div>
                        </form>
                      ) : (
                        <>
                          <div className="comment-meta">
                            <strong>{comment.authorUsername}</strong>
                            <span className="muted">{formatWhen(comment.createdAt, lang)}</span>
                            {comment.edited ? (
                              <span className="muted">
                                · {copy.edited} {formatWhen(comment.updatedAt, lang)}
                              </span>
                            ) : null}
                            {comment.authorId === user.id ? (
                              <button className="button ghost small" type="button" onClick={() => startEdit(comment)}>
                                {copy.edit}
                              </button>
                            ) : null}
                          </div>
                          <p className="comment-text">{comment.text}</p>
                        </>
                      )}
                    </div>
                  </li>
                ))}
              </ul>
            )}
            <form className="composer" onSubmit={(event) => void saveComment(event)}>
              <textarea
                value={draft}
                placeholder={copy.comment}
                aria-label={copy.comment}
                onChange={(event) => setDraft(event.target.value)}
              />
              {commentError ? <p className="alert">{commentError}</p> : null}
              <div className="composer-actions">
                <button className="button" type="submit" disabled={commentSaving}>
                  {copy.save}
                </button>
              </div>
            </form>
          </section>

          <section className="panel">
            <div className="panel-head">
              <div className="panel-title">
                <h2>
                  <Icon name="layers" size={16} />
                  {copy.parts}
                </h2>
              </div>
            </div>
            {device.parts.length === 0 ? (
              <div className="empty compact">
                <span className="empty-icon">
                  <Icon name="layers" size={16} />
                </span>
                <span>{copy.noParts}</span>
              </div>
            ) : (
              <ul className="part-list">
                {device.parts.map((part, index) => (
                  <li className="part-item" key={part.id}>
                    <span className="part-index">{index + 1}</span>
                    <div className="part-body">
                      <h3>{part.name}</h3>
                      {part.note ? <p className="notes">{part.note}</p> : null}
                      <PhotoRow copy={copy} deviceId={device.id} photos={part.photos} onUnauthorized={onUnauthorized} />
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <SpareService
            copy={copy}
            serviceId={device.id}
            onNavigate={onNavigate}
            onUnauthorized={onUnauthorized}
          />

          <section className="panel">
            <div className="panel-head">
              <div className="panel-title">
                <h2>
                  <Icon name="history" size={16} />
                  {copy.statusHistory}
                </h2>
              </div>
            </div>
            {device.history.length === 0 ? (
              <div className="empty compact">
                <span className="empty-icon">
                  <Icon name="history" size={16} />
                </span>
                <span>{copy.noStatusHistory}</span>
              </div>
            ) : (
              <ol className="timeline">
                {device.history.map((change) => (
                  <li className="timeline-item" key={change.id}>
                    <span className="timeline-dot" data-tone={statusTone(change.to)} />
                    <div className="timeline-body">
                      <div className="timeline-change">
                        <span className="badge" data-tone={statusTone(change.from)}>
                          <span>{statusLabel(change.from, states, lang)}</span>
                        </span>
                        <Icon name="arrowRight" size={14} />
                        <span className="badge" data-tone={statusTone(change.to)}>
                          <span>{statusLabel(change.to, states, lang)}</span>
                        </span>
                      </div>
                      <p className="meta">
                        <span className="meta-item">{formatWhen(change.createdAt, lang)}</span>
                        <span className="meta-item">{change.actorUsername}</span>
                      </p>
                      {change.note ? <p className="notes">{change.note}</p> : null}
                      {change.hasPhoto ? (
                        <div className="photo-grid">
                          <div className="photo-tile sm">
                            <PhotoButton
                              copy={copy}
                              src={statusPhotoSrc(device.id, change.id)}
                              name={String(change.id)}
                              onUnauthorized={onUnauthorized}
                            />
                          </div>
                        </div>
                      ) : null}
                    </div>
                  </li>
                ))}
              </ol>
            )}
          </section>
        </div>

        <aside className="detail-side">
          <section className="panel">
            <div className="panel-head">
              <div className="panel-title">
                <h2>{copy.customer}</h2>
              </div>
            </div>
            <button
              className="person"
              type="button"
              onClick={() => onNavigate(`/customers/${device.customer.id}`)}
            >
              <span className="avatar lg" data-tone="cyan" aria-hidden="true">
                {initials(device.customer.name)}
              </span>
              <span className="person-name">{device.customer.name}</span>
            </button>
            <dl className="facts">
              <div className="fact">
                <Icon name="phone" size={16} />
                <div>
                  <dt>{copy.phone}</dt>
                  <dd>{device.customer.phone || <span className="empty-value">—</span>}</dd>
                </div>
              </div>
              <div className="fact">
                <Icon name="mail" size={16} />
                <div>
                  <dt>{copy.email}</dt>
                  <dd>{device.customer.email || <span className="empty-value">—</span>}</dd>
                </div>
              </div>
              <div className="fact">
                <Icon name="pin" size={16} />
                <div>
                  <dt>{copy.address}</dt>
                  <dd>{device.customer.address || <span className="empty-value">—</span>}</dd>
                </div>
              </div>
            </dl>
          </section>
          <section className="panel">
            <div className="panel-head">
              <div className="panel-title">
                <h2>{copy.device}</h2>
              </div>
            </div>
            {device.note ? <p className="notes">{device.note}</p> : <p className="muted">{copy.device}</p>}
            <PhotoRow copy={copy} deviceId={device.id} photos={device.photos} onUnauthorized={onUnauthorized} />
          </section>
        </aside>
      </div>
      {reportLayout && reportSpares
        ? createPortal(
            <ServiceReport
              copy={copy}
              lang={lang}
              brand={brand}
              states={states}
              device={device}
              spares={reportSpares}
              layout={reportLayout}
            />,
            document.body,
          )
        : null}
    </div>
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
    <div className="photo-grid">
      {photos.map((photo) => (
        <div className="photo-tile" key={photo.id}>
          <PhotoButton
            copy={copy}
            src={photoSrc(deviceId, photo.id)}
            name={photo.id}
            onUnauthorized={onUnauthorized}
          />
        </div>
      ))}
    </div>
  );
}
