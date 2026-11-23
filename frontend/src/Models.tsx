import { useEffect, useState, type FormEvent } from "react";
import { ApiError, request, type CatalogEntry, type Page } from "./api";
import { CatalogField } from "./CatalogField";
import type { Copy } from "./i18n";
import { Icon } from "./icons";
import { End, useQuery, useWindow } from "./list";
import { ModelIcon } from "./ModelIcon";
import { problemText } from "./problems";
import { nameFields } from "./spare";

type Props = {
  copy: Copy;
  onUnauthorized: () => void;
};

export function Models({ copy, onUnauthorized }: Props) {
  const search = useQuery();
  const [listKey, setListKey] = useState(0);
  const list = useWindow(`${search.query}\n${listKey}`, (after: CatalogEntry | null, signal) => {
    const params = new URLSearchParams();
    if (search.query) {
      params.set("q", search.query);
    }
    if (after) {
      params.set("after", after.brand);
      params.set("afterModel", after.model);
      params.set("afterId", after.id);
    }
    return request<Page<CatalogEntry>>(`/api/models?${params}`, { signal });
  });
  const [brand, setBrand] = useState("");
  const [model, setModel] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activePatch, setActivePatch] = useState<Record<string, boolean>>({});
  const [pending, setPending] = useState<string | null>(null);

  useEffect(() => {
    setActivePatch({});
  }, [search.query, listKey]);

  useEffect(() => {
    if (list.error instanceof ApiError && list.error.code === "unauthorized") {
      onUnauthorized();
    }
  }, [list.error, onUnauthorized]);

  async function create(event: FormEvent) {
    event.preventDefault();
    const names = nameFields(brand, model);
    if (names.missing.length || names.tooLong.length) {
      setError(problemText(copy, new ApiError("validation", names)));
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await request<CatalogEntry>("/api/models", {
        method: "POST",
        body: JSON.stringify({ brand, model }),
      });
      setBrand("");
      setModel("");
      setListKey((current) => current + 1);
    } catch (err: unknown) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setError(problemText(copy, err));
    } finally {
      setSaving(false);
    }
  }

  async function setActive(item: CatalogEntry, active: boolean) {
    setPending(item.id);
    setActivePatch((current) => ({ ...current, [item.id]: active }));
    setError(null);
    try {
      const saved = await request<CatalogEntry>(`/api/models/${item.id}`, {
        method: "PATCH",
        body: JSON.stringify({ active }),
      });
      setActivePatch((current) => ({ ...current, [item.id]: saved.active }));
    } catch (err: unknown) {
      setActivePatch((current) => {
        const next = { ...current };
        delete next[item.id];
        return next;
      });
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      setError(problemText(copy, err));
    } finally {
      setPending(null);
    }
  }

  const empty = !list.loading && list.items.length === 0;

  return (
    <div className="page">
      <header className="page-head">
        <div className="page-title">
          <div className="page-title-text">
            <h1>{copy.models}</h1>
            <p className="lead">{copy.modelsLead}</p>
            <p className="muted">{copy.modelsCaption}</p>
          </div>
        </div>
      </header>
      <form className="panel" onSubmit={(event) => void create(event)}>
        <div className="panel-head">
          <div className="panel-title">
            <h2>{copy.addModel}</h2>
          </div>
        </div>
        <div className="fields-2">
          <CatalogField id="model-brand" label={copy.brand} value={brand} onChange={setBrand} onUnauthorized={onUnauthorized} />
          <CatalogField
            id="model-name"
            label={copy.model}
            value={model}
            brand={brand}
            onChange={setModel}
            onUnauthorized={onUnauthorized}
          />
        </div>
        {error ? <p className="alert">{error}</p> : null}
        <div className="form-actions">
          <button className="button" type="submit" disabled={saving}>
            {copy.addModel}
          </button>
        </div>
      </form>
      <form className="toolbar" onSubmit={search.commit}>
        <label className="field search">
          <span className="sr-only">{copy.search}</span>
          <Icon name="search" size={16} />
          <input
            type="search"
            value={search.value}
            placeholder={copy.spareSearch}
            onChange={(event) => search.setValue(event.target.value)}
          />
        </label>
        <button className="button secondary" type="submit">
          {copy.search}
        </button>
      </form>
      {list.error ? <p className="alert">{problemText(copy, list.error)}</p> : null}
      <div className="panel flush">
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>{copy.brand}</th>
                <th>{copy.model}</th>
                <th>{copy.source}</th>
                <th>{copy.stateActive}</th>
              </tr>
            </thead>
            <tbody>
              {list.items.map((item) => {
                const active = activePatch[item.id] ?? item.active;
                return (
                  <tr key={item.id}>
                    <td data-label={copy.brand} className="cell-main cell-strong">
                      {item.brand}
                    </td>
                    <td data-label={copy.model}>
                      <span className="thing">
                        {item.image ? <ModelIcon brand={item.brand} model={item.model} /> : null}
                        <span className="thing-title">{item.model}</span>
                      </span>
                    </td>
                    <td data-label={copy.source}>{item.source === "catalog" ? copy.sourceCatalog : copy.sourceAdded}</td>
                    <td data-label={copy.stateActive}>
                      <label className="check">
                        <input
                          type="checkbox"
                          checked={active}
                          disabled={pending === item.id}
                          onChange={(event) => void setActive(item, event.target.checked)}
                        />
                        {active ? copy.stateActive : copy.stateInactive}
                      </label>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
        {empty ? (
          <div className="empty">
            <span className="empty-icon">
              <Icon name="device" size={18} />
            </span>
            <span>{search.query ? copy.noModelMatches : copy.noModels}</span>
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
