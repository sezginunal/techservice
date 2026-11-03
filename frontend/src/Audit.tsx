import { useEffect, useState, type FormEvent } from "react";
import { ApiError, request, type AuditItem, type Page, type Role } from "./api";
import { dayStart, formatWhen, nextDay, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { End, usePaused, useWindow } from "./list";
import { problemText } from "./problems";

const eventIds = [
  "customer.created",
  "customer.updated",
  "device.registered",
  "device.updated",
  "device.status_changed",
  "user.created",
  "user.role_changed",
  "user.password_changed",
  "user.activated",
  "user.deactivated",
] as const;

type Props = {
  copy: Copy;
  lang: Lang;
  onUnauthorized: () => void;
};

export function Audit({ copy, lang, onUnauthorized }: Props) {
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [person, setPerson] = useState("");
  const [personQuery, setPersonQuery] = usePaused(person);
  const [event, setEvent] = useState("");
  const [subject, setSubject] = useState("");
  const [subjectQuery, setSubjectQuery] = usePaused(subject);
  const [sort, setSort] = useState("desc");
  const key = [from, to, personQuery, event, subjectQuery, sort].join("|");
  const list = useWindow(key, (after: AuditItem | null, signal) => {
    const params = new URLSearchParams();
    const start = dayStart(from);
    const end = nextDay(to);
    if (start) {
      params.set("from", start);
    }
    if (end) {
      params.set("to", end);
    }
    if (personQuery) {
      params.set("person", personQuery);
    }
    if (event) {
      params.set("event", event);
    }
    if (subjectQuery) {
      params.set("subject", subjectQuery);
    }
    if (sort === "asc") {
      params.set("sort", "asc");
    }
    if (after) {
      params.set("after", after.at);
      params.set("afterId", String(after.id));
    }
    return request<Page<AuditItem>>(`/api/audit?${params}`, { signal });
  });

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  function submit(formEvent: FormEvent) {
    formEvent.preventDefault();
    setPersonQuery(person.trim());
    setSubjectQuery(subject.trim());
  }

  const filtered = Boolean(from || to || personQuery || event || subjectQuery);
  const empty = !list.loading && list.items.length === 0;

  return (
    <>
      <header className="page-head">
        <h1>{copy.audit}</h1>
        <p className="muted">{copy.auditLead}</p>
      </header>
      <form className="filters" onSubmit={submit}>
        <label className="field">
          {copy.fromDate}
          <input type="date" value={from} onChange={(input) => setFrom(input.target.value)} />
        </label>
        <label className="field">
          {copy.toDate}
          <input type="date" value={to} onChange={(input) => setTo(input.target.value)} />
        </label>
        <label className="field">
          {copy.person}
          <input value={person} onChange={(input) => setPerson(input.target.value)} />
        </label>
        <label className="field">
          {copy.event}
          <select value={event} onChange={(input) => setEvent(input.target.value)}>
            <option value="">{copy.allEvents}</option>
            {eventIds.map((id) => (
              <option key={id} value={id}>
                {eventLabel(copy, id)}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          {copy.subject}
          <input value={subject} onChange={(input) => setSubject(input.target.value)} />
        </label>
        <label className="field">
          {copy.sortBy}
          <select value={sort} onChange={(input) => setSort(input.target.value)}>
            <option value="desc">{copy.newest}</option>
            <option value="asc">{copy.oldest}</option>
          </select>
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
              <th>{copy.when}</th>
              <th>{copy.person}</th>
              <th>{copy.role}</th>
              <th>{copy.event}</th>
              <th>{copy.subject}</th>
            </tr>
          </thead>
          <tbody>
            {list.items.map((item) => (
              <tr key={item.id}>
                <td data-label={copy.when}>{formatWhen(item.at, lang)}</td>
                <td data-label={copy.person}>{item.actorUsername}</td>
                <td data-label={copy.role}>{copy[item.actorRole]}</td>
                <td data-label={copy.event}>{eventLabel(copy, item.event)}</td>
                <td data-label={copy.subject}>{subjectOf(copy, item)}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {empty ? <p className="muted empty">{filtered ? copy.noAuditMatches : copy.emptyAudit}</p> : null}
        {list.loading ? <p className="muted empty">{copy.loading}</p> : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </div>
    </>
  );
}

function eventLabel(copy: Copy, event: string) {
  const labels: Record<string, string> = {
    "customer.created": copy.eventCustomerCreated,
    "customer.updated": copy.eventCustomerUpdated,
    "device.registered": copy.eventDeviceRegistered,
    "device.updated": copy.eventDeviceUpdated,
    "device.status_changed": copy.eventDeviceStatusChanged,
    "user.created": copy.eventUserCreated,
    "user.role_changed": copy.eventUserRoleChanged,
    "user.password_changed": copy.eventUserPasswordChanged,
    "user.activated": copy.eventUserActivated,
    "user.deactivated": copy.eventUserDeactivated,
  };
  return labels[event] ?? event;
}

function subjectOf(copy: Copy, item: AuditItem) {
  const detail = item.detail;
  if (item.event === "customer.updated" && detail.previousName) {
    return `${detail.previousName} → ${detail.name ?? ""}`;
  }
  if (item.event === "customer.created" || item.event === "customer.updated") {
    return detail.name ?? "";
  }
  if (item.event === "device.registered" || item.event === "device.updated") {
    return [detail.name, detail.uuid, detail.customer].filter(Boolean).join(" · ");
  }
  if (item.event === "device.status_changed") {
    const from = statusLabel(detail.from ?? "", copy);
    const to = statusLabel(detail.to ?? "", copy);
    return [detail.name, detail.uuid, `${from} → ${to}`].filter(Boolean).join(" · ");
  }
  if (item.event === "user.created") {
    return [detail.username, roleName(copy, detail.role)].filter(Boolean).join(" · ");
  }
  if (item.event === "user.role_changed") {
    return `${detail.username ?? ""} · ${roleName(copy, detail.from)} → ${roleName(copy, detail.to)}`;
  }
  return detail.username ?? "";
}

function roleName(copy: Copy, role?: string) {
  if (role === "admin" || role === "owner" || role === "technician") {
    return copy[role as Role];
  }
  return role ?? "";
}
