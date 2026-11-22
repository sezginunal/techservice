import { useEffect, useState, type FormEvent } from "react";
import { ApiError, request, type AuditItem, type Page, type Role, type ServiceState } from "./api";
import { DateField } from "./DateField";
import { dayStart, formatWhen, nextDay, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
import { End, usePaused, useWindow } from "./list";
import { problemText } from "./problems";

const eventIds = [
  "customer.created",
  "customer.updated",
  "service.registered",
  "service.updated",
  "device.updated",
  "service.status_changed",
  "user.created",
  "user.role_changed",
  "user.password_changed",
  "user.activated",
  "user.deactivated",
  "state.added",
  "state.updated",
  "states.reordered",
  "spare_part.added",
  "spare_part.updated",
  "stock.added",
  "stock.written_off",
  "spare_part.used",
  "spare_part.returned",
] as const;

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  onUnauthorized: () => void;
};

export function Audit({ copy, lang, states, onUnauthorized }: Props) {
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
    <div className="page">
      <header className="page-head">
        <div className="page-title">
          <div className="page-title-text">
            <h1>{copy.audit}</h1>
            <p className="lead">{copy.auditLead}</p>
          </div>
        </div>
      </header>
      <form className="filters toolbar-filters" onSubmit={submit}>
        <DateField copy={copy} lang={lang} label={copy.fromDate} value={from} onChange={setFrom} />
        <DateField copy={copy} lang={lang} label={copy.toDate} value={to} onChange={setTo} />
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
      <div className="panel flush">
        <div className="table-wrap">
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
                  <td data-label={copy.when} className="cell-nowrap">
                    {formatWhen(item.at, lang)}
                  </td>
                  <td data-label={copy.person} className="cell-strong">
                    {item.actorUsername}
                  </td>
                  <td data-label={copy.role}>{copy[item.actorRole]}</td>
                  <td data-label={copy.event}>{eventLabel(copy, item.event)}</td>
                  <td data-label={copy.subject} className="wrap">
                    {subjectOf(copy, lang, states, item)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {empty ? (
          <div className="empty">
            <span className="empty-icon">
              <Icon name="audit" size={18} />
            </span>
            <span>{filtered ? copy.noAuditMatches : copy.emptyAudit}</span>
          </div>
        ) : null}
        {list.loading ? (
          <div className="loading">
            <span className="spinner" />
            {copy.loading}
          </div>
        ) : null}
        <End enabled={list.hasMore && !list.loading} onEnd={list.more} />
      </div>
    </div>
  );
}

function eventLabel(copy: Copy, event: string) {
  const labels: Record<string, string> = {
    "customer.created": copy.eventCustomerCreated,
    "customer.updated": copy.eventCustomerUpdated,
    "service.registered": copy.eventDeviceRegistered,
    "service.updated": copy.eventServiceUpdated,
    "device.updated": copy.eventDeviceUpdated,
    "service.status_changed": copy.eventDeviceStatusChanged,
    "user.created": copy.eventUserCreated,
    "user.role_changed": copy.eventUserRoleChanged,
    "user.password_changed": copy.eventUserPasswordChanged,
    "user.activated": copy.eventUserActivated,
    "user.deactivated": copy.eventUserDeactivated,
    "state.added": copy.eventStateAdded,
    "state.updated": copy.eventStateUpdated,
    "states.reordered": copy.eventStatesReordered,
    "spare_part.added": copy.eventSpareAdded,
    "spare_part.updated": copy.eventSpareUpdated,
    "stock.added": copy.eventStockAdded,
    "stock.written_off": copy.eventStockWrittenOff,
    "spare_part.used": copy.eventSpareUsed,
    "spare_part.returned": copy.eventSpareReturned,
  };
  return labels[event] ?? event;
}

function subjectOf(copy: Copy, lang: Lang, states: ServiceState[], item: AuditItem) {
  const detail = item.detail;
  if (item.event === "customer.updated" && detail.previousName) {
    return `${detail.previousName} → ${detail.name ?? ""}`;
  }
  if (item.event === "customer.created" || item.event === "customer.updated") {
    return detail.name ?? "";
  }
  if (item.event === "service.registered" || item.event === "service.updated") {
    const customer =
      detail.previousCustomer && detail.customer ? `${detail.previousCustomer} → ${detail.customer}` : detail.customer;
    return [detail.brand, detail.model, detail.serial, detail.uuid, customer].filter(Boolean).join(" · ");
  }
  if (item.event === "device.updated") {
    const brand = detail.previousBrand ? `${detail.previousBrand} → ${detail.brand ?? ""}` : detail.brand;
    const model = detail.previousModel ? `${detail.previousModel} → ${detail.model ?? ""}` : detail.model;
    const serial = detail.previousSerial ? `${detail.previousSerial} → ${detail.serial ?? ""}` : detail.serial;
    return [brand, model, serial, detail.uuid].filter(Boolean).join(" · ");
  }
  if (item.event === "service.status_changed") {
    const from = statusLabel(detail.from ?? "", states, lang);
    const to = statusLabel(detail.to ?? "", states, lang);
    return [detail.brand, detail.model, detail.serial, detail.uuid, `${from} → ${to}`].filter(Boolean).join(" · ");
  }
  if (item.event === "user.created") {
    return [detail.username, roleName(copy, detail.role)].filter(Boolean).join(" · ");
  }
  if (item.event === "user.role_changed") {
    return `${detail.username ?? ""} · ${roleName(copy, detail.from)} → ${roleName(copy, detail.to)}`;
  }
  if (item.event === "spare_part.updated") {
    const from = [detail.previousBrand, detail.previousModel].filter(Boolean).join(" ");
    const to = [detail.brand, detail.model].filter(Boolean).join(" ");
    return `${from} → ${to}`;
  }
  if (
    item.event === "spare_part.added" ||
    item.event === "stock.added" ||
    item.event === "stock.written_off" ||
    item.event === "spare_part.used" ||
    item.event === "spare_part.returned"
  ) {
    const parts: Array<string | number | undefined> = [detail.brand, detail.model, detail.quantity];
    if (item.event === "spare_part.used" || item.event === "spare_part.returned") {
      const device = [detail.deviceBrand, detail.deviceModel].filter(Boolean).join(" ");
      parts.push(device, detail.serial, detail.uuid, detail.status ? statusLabel(detail.status, states, lang) : "");
    }
    return parts.filter((value) => value !== undefined && value !== "").join(" · ");
  }
  if (item.event === "state.added" || item.event === "state.updated" || item.event === "states.reordered") {
    return stateSubject(copy, lang, detail, item.event);
  }
  return detail.username ?? "";
}

function stateName(lang: Lang, detail: AuditItem["detail"]) {
  return lang === "tr" ? detail.nameTr ?? detail.nameEn ?? "" : detail.nameEn ?? detail.nameTr ?? "";
}

function stateSubject(copy: Copy, lang: Lang, detail: AuditItem["detail"], event: string) {
  const name = stateName(lang, detail);
  if (event === "states.reordered") {
    return name;
  }
  const parts = [name];
  if (detail.previousNameEn || detail.previousNameTr) {
    const previous = lang === "tr" ? detail.previousNameTr ?? "" : detail.previousNameEn ?? "";
    parts[0] = `${previous} → ${name}`;
  }
  if (detail.previousActive && detail.active) {
    parts.push(`${activeWord(copy, detail.previousActive)} → ${activeWord(copy, detail.active)}`);
  }
  if (detail.previousHome && detail.home) {
    parts.push(`${homeWord(copy, detail.previousHome)} → ${homeWord(copy, detail.home)}`);
  } else if (event === "state.added" && detail.showOnHome === "shown") {
    parts.push(copy.shownOnHome);
  }
  return parts.filter(Boolean).join(" · ");
}

function activeWord(copy: Copy, value: string) {
  return value === "active" ? copy.stateActive : copy.stateInactive;
}

function homeWord(copy: Copy, value: string) {
  return value === "shown" ? copy.shownOnHome : copy.hiddenOnHome;
}

function roleName(copy: Copy, role?: string) {
  if (role === "admin" || role === "owner" || role === "technician") {
    return copy[role as Role];
  }
  return role ?? "";
}
