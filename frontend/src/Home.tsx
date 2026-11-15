import { useEffect, useState } from "react";
import { ApiError, request, type HomeSummary, type ServiceState } from "./api";
import { formatWhen, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { ModelIcon } from "./ModelIcon";
import { problemText } from "./problems";

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function Home({ copy, lang, states, onNavigate, onUnauthorized }: Props) {
  const [summary, setSummary] = useState<HomeSummary | null>(null);
  const [error, setError] = useState<unknown>(null);

  useEffect(() => {
    const controller = new AbortController();
    request<HomeSummary>("/api/home", { signal: controller.signal })
      .then((body) => {
        if (!controller.signal.aborted) {
          setSummary(body);
        }
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") {
          return;
        }
        if (!controller.signal.aborted) {
          setError(err);
        }
      });
    return () => controller.abort();
  }, []);

  useEffect(() => {
    if (error instanceof ApiError && error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [error, onUnauthorized]);

  return (
    <>
      <header className="page-head">
        <h1>{copy.home}</h1>
      </header>
      {error ? <p className="alert">{problemText(copy, error)}</p> : null}
      {!summary && !error ? <p className="muted">{copy.loading}</p> : null}
      {summary ? <Shop copy={copy} lang={lang} states={states} summary={summary} onNavigate={onNavigate} /> : null}
    </>
  );
}

function Shop({
  copy,
  lang,
  states,
  summary,
  onNavigate,
}: {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  summary: HomeSummary;
  onNavigate: (path: string) => void;
}) {
  const peak = Math.max(1, ...summary.days.map((day) => day.count));
  const recentFrom = summary.days.length - 7;
  const modelPeak = Math.max(summary.other, ...summary.models.map((item) => item.count), 1);

  return (
    <>
      <section className="panel stack">
        <h2>{copy.registrations}</h2>
        <div className="figures">
          <div className="figure">
            <h3>{copy.last7Days}</h3>
            <p className="figure-value">{summary.registered7}</p>
          </div>
          <div className="figure">
            <h3>{copy.last30Days}</h3>
            <p className="figure-value">{summary.registered30}</p>
          </div>
        </div>
        <p className="muted">{copy.registrationsCaption}</p>
        <div className="bars-scroll">
          <div className="bars">
            {summary.days.map((day, index) => {
              const label = chartLabel(day.date, day.count, lang);
              return (
                <div key={day.date} className={index >= recentFrom ? "bar recent" : "bar"}>
                  <div className="bar-plot">
                    <div
                      className="bar-fill"
                      style={{ height: day.count === 0 ? "2px" : `${(day.count / peak) * 100}%` }}
                      role="img"
                      aria-label={label}
                      title={label}
                    />
                  </div>
                  <span className="bar-label">{index % 5 === 4 ? String(Number(day.date.slice(8))) : ""}</span>
                </div>
              );
            })}
          </div>
        </div>
      </section>
      <div className="figures">
        <section className="panel figure">
          <h2>{copy.inShop}</h2>
          <p className="figure-value">{summary.inShop}</p>
          <p className="muted">{copy.deliveredCount.replace("{n}", String(summary.delivered))}</p>
        </section>
        <section className="panel figure">
          <h2>{copy.idle7}</h2>
          <p className="figure-value">{summary.idle7}</p>
        </section>
      </div>
      {summary.queues.map((queue) => (
        <section className="panel stack" key={queue.status}>
          <header className="section-head">
            <h2>
              <button
                className="button linkish"
                type="button"
                onClick={() => onNavigate(`/services?status=${encodeURIComponent(queue.status)}`)}
              >
                {statusLabel(queue.status, states, lang)}
              </button>
            </h2>
          </header>
          {queue.more ? <p className="muted">{copy.showingLongest}</p> : null}
          {queue.items.length === 0 ? (
            <p className="muted empty">{copy.noQueue}</p>
          ) : (
            <div className="home-list">
              {queue.items.map((item) => (
                <button
                  key={item.id}
                  className="home-link"
                  type="button"
                  data-wait={waitTone(item.days)}
                  onClick={() => onNavigate(`/services/${item.id}`)}
                >
                  <span className="home-id">
                    <strong>
                      {item.brand} {item.model} · {item.serial}
                    </strong>
                    <span className="muted">{item.customerName}</span>
                  </span>
                  <span className="home-meta">
                    <strong>{queueDayCount(copy, item.days)}</strong>
                    <span className="muted">
                      {copy.sameStatusSince} {formatWhen(item.since, lang)}
                    </span>
                  </span>
                </button>
              ))}
            </div>
          )}
        </section>
      ))}
      <div className="home-split">
        <section className="panel stack">
          <header className="section-head">
            <h2>{copy.byStatus}</h2>
            <button className="button linkish" type="button" onClick={() => onNavigate("/services")}>
              {copy.allDevices}
            </button>
          </header>
          <div className="home-list">
            {summary.byStatus.map((item) => (
              <button
                key={item.status}
                className="home-link"
                type="button"
                onClick={() => onNavigate(`/services?status=${encodeURIComponent(item.status)}`)}
              >
                <span>{statusLabel(item.status, states, lang)}</span>
                <strong>{item.count}</strong>
              </button>
            ))}
          </div>
        </section>
        <section className="panel stack">
          <header>
            <h2>{copy.byModel}</h2>
            <p className="muted">{copy.byModelCaption}</p>
          </header>
          {summary.models.length === 0 ? (
            <p className="muted empty">{copy.noRegistrations30}</p>
          ) : (
            <div className="model-list">
              {summary.models.map((item) => (
                <ModelRow
                  key={`${item.brand}/${item.model}`}
                  brand={item.brand}
                  model={item.model}
                  count={item.count}
                  image={item.image}
                  peak={modelPeak}
                />
              ))}
              {summary.other > 0 ? (
                <ModelRow brand="" model={copy.otherModels} count={summary.other} image={false} peak={modelPeak} />
              ) : null}
            </div>
          )}
        </section>
      </div>
      <section className="panel stack">
        <header>
          <h2>{copy.unchanged}</h2>
          <p className="muted">{copy.unchangedCaption}</p>
        </header>
        {summary.unchangedMore ? <p className="muted">{copy.showingLongest}</p> : null}
        {summary.unchanged.length === 0 ? (
          <p className="muted empty">{copy.noFullDay}</p>
        ) : (
          <div className="home-list">
            {summary.unchanged.map((item) => (
              <button
                key={item.id}
                className="home-link"
                type="button"
                data-wait={waitTone(item.days)}
                onClick={() => onNavigate(`/services/${item.id}`)}
              >
                <span className="home-id">
                  <strong>
                    {item.brand} {item.model} · {item.serial}
                  </strong>
                  <span className="muted">{item.customerName}</span>
                </span>
                <span className="home-meta">
                  <span className="badge">{statusLabel(item.status, states, lang)}</span>
                  <strong>{dayCount(copy, item.days)}</strong>
                  <span className="muted">
                    {copy.sameStatusSince} {formatWhen(item.since, lang)}
                  </span>
                </span>
              </button>
            ))}
          </div>
        )}
      </section>
      <section className="panel stack">
        <h2>{copy.latestRegistrations}</h2>
        {summary.latest.length === 0 ? (
          <p className="muted empty">{copy.noDevices}</p>
        ) : (
          <div className="home-list">
            {summary.latest.map((item) => (
              <button
                key={item.id}
                className="home-link"
                type="button"
                onClick={() => onNavigate(`/services/${item.id}`)}
              >
                <span className="with-icon">
                  {item.image ? <ModelIcon brand={item.brand} model={item.model} /> : null}
                  <strong>
                    {item.brand} {item.model} · {item.serial}
                  </strong>
                </span>
                <span className="home-meta">
                  <span className="badge">{statusLabel(item.status, states, lang)}</span>
                  <span className="muted">{formatWhen(item.createdAt, lang)}</span>
                </span>
              </button>
            ))}
          </div>
        )}
      </section>
    </>
  );
}

function ModelRow({
  brand,
  model,
  count,
  image,
  peak,
}: {
  brand: string;
  model: string;
  count: number;
  image: boolean;
  peak: number;
}) {
  const name = brand ? `${brand} ${model}` : model;
  return (
    <div className="model-row">
      {image ? <ModelIcon brand={brand} model={model} /> : <span className="model-gap" />}
      <span className="model-name">{name}</span>
      <strong className="model-count">{count}</strong>
      <span className="model-meter" aria-hidden="true">
        <span style={{ width: `${(count / peak) * 100}%` }} />
      </span>
    </div>
  );
}

function chartLabel(date: string, count: number, lang: Lang) {
  const [year, month, day] = date.split("-").map(Number);
  const formatted = new Intl.DateTimeFormat(lang === "tr" ? "tr-TR" : "en-GB", {
    day: "numeric",
    month: "short",
  }).format(new Date(year, month - 1, day));
  return `${formatted}: ${count}`;
}

function queueDayCount(copy: Copy, days: number) {
  if (days < 1) {
    return copy.lessThanDay;
  }
  return dayCount(copy, days);
}

function dayCount(copy: Copy, days: number) {
  if (days === 1) {
    return copy.oneDay;
  }
  return copy.manyDays.replace("{n}", String(days));
}

function waitTone(days: number) {
  if (days >= 30) {
    return "dark";
  }
  if (days >= 14) {
    return "red";
  }
  if (days >= 7) {
    return "orange";
  }
  if (days >= 3) {
    return "amber";
  }
  return "neutral";
}
