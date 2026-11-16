import { useEffect, useState } from "react";
import { ApiError, request, type HomeSummary, type ServiceState } from "./api";
import { formatWhen, statusLabel, statusTone } from "./format";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
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

  if (error) {
    return (
      <div className="page">
        <header className="page-head">
          <div className="page-title">
            <div className="page-title-text">
              <h1>{copy.home}</h1>
            </div>
          </div>
        </header>
        <p className="alert">{problemText(copy, error)}</p>
      </div>
    );
  }

  if (!summary) {
    return (
      <div className="page page-loading">
        <span className="skeleton" />
        <span className="skeleton" />
        <span className="skeleton" />
      </div>
    );
  }

  return <Shop copy={copy} lang={lang} states={states} summary={summary} onNavigate={onNavigate} />;
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
    <div className="page">
      <header className="page-head">
        <div className="page-title">
          <div className="page-title-text">
            <h1>{copy.home}</h1>
          </div>
        </div>
      </header>

      <div className="dash">
        <div className="kpi-grid">
          <section className="panel">
            <div className="kpi" data-tone="teal">
              <span className="kpi-icon">
                <Icon name="chart" size={18} />
              </span>
              <div className="kpi-text">
                <span className="kpi-label">{copy.last7Days}</span>
                <strong className="kpi-value">{summary.registered7}</strong>
              </div>
            </div>
          </section>
          <section className="panel">
            <div className="kpi" data-tone="cyan">
              <span className="kpi-icon">
                <Icon name="calendar" size={18} />
              </span>
              <div className="kpi-text">
                <span className="kpi-label">{copy.last30Days}</span>
                <strong className="kpi-value">{summary.registered30}</strong>
              </div>
            </div>
          </section>
          <section className="panel">
            <div className="kpi" data-tone="blue">
              <span className="kpi-icon">
                <Icon name="inbox" size={18} />
              </span>
              <div className="kpi-text">
                <span className="kpi-label">{copy.inShop}</span>
                <strong className="kpi-value">{summary.inShop}</strong>
                <span className="kpi-sub">{copy.deliveredCount.replace("{n}", String(summary.delivered))}</span>
              </div>
            </div>
          </section>
          <section className="panel">
            <div className="kpi" data-tone="amber">
              <span className="kpi-icon">
                <Icon name="hourglass" size={18} />
              </span>
              <div className="kpi-text">
                <span className="kpi-label">{copy.idle7}</span>
                <strong className="kpi-value">{summary.idle7}</strong>
              </div>
            </div>
          </section>
        </div>

        <section className="panel chart-card">
          <div className="panel-head">
            <div className="panel-title">
              <h2>
                <Icon name="chart" size={16} />
                {copy.registrations}
              </h2>
              <p>{copy.registrationsCaption}</p>
            </div>
          </div>
          <div className="chart-body">
            <div className="stats">
              <div className="stat">
                <span className="stat-label">{copy.last7Days}</span>
                <strong className="stat-value">{summary.registered7}</strong>
              </div>
              <div className="stat">
                <span className="stat-label">{copy.last30Days}</span>
                <strong className="stat-value">{summary.registered30}</strong>
              </div>
            </div>
            <div className="chart">
              <div className="bars">
                {summary.days.map((day, index) => {
                  const label = chartLabel(day.date, day.count, lang);
                  return (
                    <div
                      key={day.date}
                      className={index >= recentFrom ? "bar recent" : "bar"}
                      data-tip={label}
                    >
                      <div
                        className="bar-fill"
                        style={{ height: day.count === 0 ? "2px" : `${(day.count / peak) * 100}%` }}
                        role="img"
                        aria-label={label}
                      />
                    </div>
                  );
                })}
              </div>
              <div className="bar-labels" aria-hidden="true">
                {summary.days.map((day, index) => (
                  <span key={day.date}>{index % 5 === 4 ? String(Number(day.date.slice(8))) : ""}</span>
                ))}
              </div>
              <div className="chart-legend">
                <span>
                  <span className="legend-swatch soft" />
                  {copy.last30Days}
                </span>
                <span>
                  <span className="legend-swatch" />
                  {copy.last7Days}
                </span>
              </div>
            </div>
          </div>
        </section>

        <div className="queues">
          {summary.queues.map((queue) => (
            <section className="panel flush" key={queue.status}>
              <div className="panel-head">
                <div className="panel-title">
                  <h2>
                    <button className="link" type="button" onClick={() => onNavigate(`/services?status=${encodeURIComponent(queue.status)}`)}>
                      {statusLabel(queue.status, states, lang)}
                    </button>
                  </h2>
                  {queue.more ? <p>{copy.showingLongest}</p> : null}
                </div>
              </div>
              {queue.items.length === 0 ? (
                <div className="empty compact">
                  <span className="empty-icon">
                    <Icon name="inbox" size={16} />
                  </span>
                  <span>{copy.noQueue}</span>
                </div>
              ) : (
                <ul className="rows">
                  {queue.items.map((item) => (
                    <li key={item.id}>
                      <button
                        className="row"
                        type="button"
                        data-wait={waitTone(item.days)}
                        onClick={() => onNavigate(`/services/${item.id}`)}
                      >
                        <span className="row-main">
                          <span className="row-text">
                            <span className="row-title">
                              {item.brand} {item.model} · <span className="mono">{item.serial}</span>
                            </span>
                            <span className="row-sub">{item.customerName}</span>
                          </span>
                        </span>
                        <span className="row-end">
                          <span className="wait">{queueDayCount(copy, item.days)}</span>
                          <span className="row-sub">
                            {copy.sameStatusSince} {formatWhen(item.since, lang)}
                          </span>
                        </span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          ))}
        </div>

        <section className="panel flush half span-5">
          <div className="panel-head">
            <div className="panel-title">
              <h2>{copy.byStatus}</h2>
            </div>
            <div className="panel-actions">
              <button className="button ghost small" type="button" onClick={() => onNavigate("/services")}>
                {copy.allDevices}
              </button>
            </div>
          </div>
          <ul className="rows status-rows">
            {summary.byStatus.map((item) => (
              <li key={item.status}>
                <button
                  className="row status-row"
                  type="button"
                  onClick={() => onNavigate(`/services?status=${encodeURIComponent(item.status)}`)}
                >
                  <span className="badge" data-tone={statusTone(item.status)}>
                    <span>{statusLabel(item.status, states, lang)}</span>
                  </span>
                  <span className="row-count">{item.count}</span>
                  <span className="row-chevron" aria-hidden="true">
                    <Icon name="chevronRight" size={16} />
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </section>

        <section className="panel half span-7">
          <div className="panel-head">
            <div className="panel-title">
              <h2>{copy.byModel}</h2>
              <p>{copy.byModelCaption}</p>
            </div>
          </div>
          {summary.models.length === 0 ? (
            <div className="empty">
              <span className="empty-icon">
                <Icon name="device" size={18} />
              </span>
              <span>{copy.noRegistrations30}</span>
            </div>
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

        <section className="panel flush">
          <div className="panel-head">
            <div className="panel-title">
              <h2>{copy.unchanged}</h2>
              <p>{copy.unchangedCaption}</p>
            </div>
          </div>
          {summary.unchangedMore ? (
            <p className="note">{copy.showingLongest}</p>
          ) : null}
          {summary.unchanged.length === 0 ? (
            <div className="empty compact">
              <span className="empty-icon">
                <Icon name="check" size={16} />
              </span>
              <span>{copy.noFullDay}</span>
            </div>
          ) : (
            <ul className="rows">
              {summary.unchanged.map((item) => (
                <li key={item.id}>
                  <button
                    className="row"
                    type="button"
                    data-wait={waitTone(item.days)}
                    onClick={() => onNavigate(`/services/${item.id}`)}
                  >
                    <span className="row-main">
                      <span className="row-text">
                        <span className="row-title">
                          {item.brand} {item.model} · <span className="mono">{item.serial}</span>
                        </span>
                        <span className="row-sub">{item.customerName}</span>
                      </span>
                    </span>
                    <span className="row-end">
                      <span className="badge" data-tone={statusTone(item.status)}>
                        <span>{statusLabel(item.status, states, lang)}</span>
                      </span>
                      <span className="wait">{dayCount(copy, item.days)}</span>
                      <span className="row-sub">
                        {copy.sameStatusSince} {formatWhen(item.since, lang)}
                      </span>
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="panel flush">
          <div className="panel-head">
            <div className="panel-title">
              <h2>{copy.latestRegistrations}</h2>
            </div>
          </div>
          {summary.latest.length === 0 ? (
            <div className="empty compact">
              <span className="empty-icon">
                <Icon name="device" size={16} />
              </span>
              <span>{copy.noDevices}</span>
            </div>
          ) : (
            <ul className="rows">
              {summary.latest.map((item) => (
                <li key={item.id}>
                  <button className="row" type="button" onClick={() => onNavigate(`/services/${item.id}`)}>
                    <span className="row-main">
                      {item.image ? <ModelIcon brand={item.brand} model={item.model} /> : (
                        <span className="thing-media">
                          <Icon name="device" size={16} />
                        </span>
                      )}
                      <span className="row-text">
                        <span className="row-title">
                          {item.brand} {item.model} · <span className="mono">{item.serial}</span>
                        </span>
                      </span>
                    </span>
                    <span className="row-end">
                      <span className="badge" data-tone={statusTone(item.status)}>
                        <span>{statusLabel(item.status, states, lang)}</span>
                      </span>
                      <span className="row-sub">{formatWhen(item.createdAt, lang)}</span>
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
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
      {image ? (
        <ModelIcon brand={brand} model={model} />
      ) : (
        <span className="thing-media">
          <Icon name="device" size={16} />
        </span>
      )}
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
  return undefined;
}
