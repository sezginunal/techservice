import { useEffect, useId, useLayoutEffect, useRef, useState, type FormEvent } from "react";
import { createPortal } from "react-dom";
import { ApiError, photoSrc, request, type CatalogBrand, type CatalogModel, type Customer, type Device, type EarlierService, type Page, type ServiceState } from "./api";
import { ModelIcon } from "./ModelIcon";
import { formatWhen, statusLabel } from "./format";
import type { Copy, Lang } from "./i18n";
import { Icon } from "./icons";
import { End, usePaused, useWindow } from "./list";
import { PhotoButton } from "./PhotoView";
import { emailOk, phoneOk } from "./fields";
import { fileProblem, maxPhotos, problemText } from "./problems";

type LocalFile = { key: string; file: File; url: string };
type ExistingPhoto = { id: string };
type PartDraft = {
  key: string;
  id: string;
  name: string;
  note: string;
  existing: ExistingPhoto[];
  files: LocalFile[];
};

type Props = {
  copy: Copy;
  lang: Lang;
  states: ServiceState[];
  mode: "create" | "edit";
  deviceId?: string;
  customerId?: string | null;
  onNavigate: (path: string) => void;
  onUnauthorized: () => void;
};

export function DeviceForm({ copy, lang, states, mode, deviceId, customerId, onNavigate, onUnauthorized }: Props) {
  const editing = mode === "edit";
  const urls = useRef(new Set<string>());
  const [ready, setReady] = useState(!editing);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<Customer | null>(null);
  const [customerLink, setCustomerLink] = useState<number | null>(null);
  const [shopServiceId, setShopServiceId] = useState("");
  const [customerDraft, setCustomerDraft] = useState("");
  const [customerQuery] = usePaused(customerDraft);
  const [results, setResults] = useState<Customer[]>([]);
  const [resultsMore, setResultsMore] = useState(false);
  const [customerName, setCustomerName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  const [address, setAddress] = useState("");
  const [brand, setBrand] = useState("");
  const [model, setModel] = useState("");
  const [catalog, setCatalog] = useState<CatalogBrand[]>([]);
  const [serial, setSerial] = useState("");
  const [note, setNote] = useState("");
  const [deviceExisting, setDeviceExisting] = useState<ExistingPhoto[]>([]);
  const [deviceFiles, setDeviceFiles] = useState<LocalFile[]>([]);
  const [parts, setParts] = useState<PartDraft[]>([]);

  useEffect(() => {
    const owned = urls.current;
    return () => owned.forEach((url) => URL.revokeObjectURL(url));
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    request<CatalogBrand[]>("/api/catalog", { signal: controller.signal })
      .then(setCatalog)
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        if (err instanceof ApiError && err.code === "unauthorized") {
          onUnauthorized();
        }
      });
    return () => controller.abort();
  }, [onUnauthorized]);

  useEffect(() => {
    if (!editing || !deviceId) {
      return;
    }
    const controller = new AbortController();
    request<Device>(`/api/services/${deviceId}`, { signal: controller.signal })
      .then((device) => {
        setBrand(device.brand);
        setModel(device.model);
        setSerial(device.serial);
        setNote(device.note);
        setCustomerLink(device.customer.id);
        setCustomerName(device.customer.name);
        setPhone(device.customer.phone);
        setEmail(device.customer.email);
        setAddress(device.customer.address);
        setDeviceExisting(device.photos.map((photo) => ({ id: photo.id })));
        setParts(
          device.parts.map((part) => ({
            key: part.id,
            id: part.id,
            name: part.name,
            note: part.note,
            existing: part.photos.map((photo) => ({ id: photo.id })),
            files: [],
          })),
        );
        setReady(true);
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
  }, [copy, deviceId, editing, onUnauthorized]);

  useEffect(() => {
    if (editing || !customerId) {
      return;
    }
    const controller = new AbortController();
    request<Customer>(`/api/customers/${customerId}`, { signal: controller.signal })
      .then(setSelected)
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
  }, [copy, customerId, editing, onUnauthorized]);

  useEffect(() => {
    if (selected || !customerQuery) {
      return;
    }
    const controller = new AbortController();
    const params = new URLSearchParams({ q: customerQuery, limit: "20" });
    request<Page<Customer>>(`/api/customers?${params}`, { signal: controller.signal })
      .then((page) => {
        setResults(page.items);
        setResultsMore(page.hasMore);
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        if (err instanceof ApiError && err.code === "unauthorized") {
          onUnauthorized();
        }
      });
    return () => controller.abort();
  }, [customerQuery, onUnauthorized, selected]);

  const [pausedIdentity] = usePaused(`${brand}\n${model}\n${serial}`);
  const lookupKey = pausedIdentity.split("\n").every((part) => part.trim()) ? pausedIdentity : "";
  const earlier = useWindow(editing ? "edit" : lookupKey, (after: EarlierService | null, signal) => {
    const [brandValue, modelValue, serialValue] = lookupKey.split("\n");
    if (editing || !brandValue?.trim() || !modelValue?.trim() || !serialValue?.trim()) {
      if (!editing) {
        setShopServiceId("");
      }
      return Promise.resolve({ items: [] as EarlierService[], hasMore: false, matched: false, openServiceId: "" });
    }
    const params = new URLSearchParams({ brand: brandValue, model: modelValue, serial: serialValue });
    if (after) {
      params.set("after", after.createdAt);
      params.set("afterId", after.id);
    }
    return request<Page<EarlierService> & { matched: boolean; openServiceId: string }>(`/api/devices/match?${params}`, { signal }).then(
      (page) => {
        if (!after) {
          setShopServiceId(page.matched ? page.openServiceId : "");
        }
        return page;
      },
    );
  });

  function takeFiles(files: File[]) {
    return files.map((file) => {
      const url = URL.createObjectURL(file);
      urls.current.add(url);
      return { key: localKey(), file, url };
    });
  }

  function dropFile(file: LocalFile) {
    URL.revokeObjectURL(file.url);
    urls.current.delete(file.url);
  }

  function photoCount() {
    return (
      deviceExisting.length +
      deviceFiles.length +
      parts.reduce((sum, part) => sum + part.existing.length + part.files.length, 0)
    );
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    const missing: string[] = [];
    const invalid: string[] = [];
    const linked = editing ? customerLink : (selected?.id ?? null);
    if (!linked && !customerName.trim()) {
      missing.push("customerName");
    }
    if (!brand.trim()) {
      missing.push("brand");
    }
    if (!model.trim()) {
      missing.push("model");
    }
    if (!serial.trim()) {
      missing.push("serial");
    }
    if (editing || !selected) {
      if (!phone.trim()) {
        missing.push("phone");
      } else if (!phoneOk(phone)) {
        invalid.push("phone");
      }
      if (!emailOk(email)) {
        invalid.push("email");
      }
    }
    parts.forEach((part, index) => {
      if (!part.name.trim()) {
        missing.push(`parts.${index}.name`);
      }
    });
    const locals = [...deviceFiles, ...parts.flatMap((part) => part.files)];
    let photo = "";
    for (const item of locals) {
      const problem = fileProblem(item.file);
      if (problem) {
        photo = problem;
        break;
      }
    }
    if (photoCount() > maxPhotos) {
      photo = "limit";
    }
    if (missing.length || invalid.length || photo) {
      setError(problemText(copy, new ApiError("validation", { missing, invalid, photo })));
      return;
    }

    const files: File[] = [];
    const indexes = (list: LocalFile[]) =>
      list.map((item) => {
        const index = files.length;
        files.push(item.file);
        return index;
      });
    const payload = editing
      ? {
          brand,
          model,
          serial,
          note,
          newCustomer: linked == null,
          ...(linked != null ? { customerId: linked } : {}),
          customerName,
          customerPhone: phone,
          customerEmail: email,
          customerAddress: address,
          deviceKeep: deviceExisting.map((photo) => photo.id),
          devicePhotos: indexes(deviceFiles),
          parts: parts.map((part) => ({
            id: part.id,
            name: part.name,
            note: part.note,
            keep: part.existing.map((photo) => photo.id),
            photos: indexes(part.files),
          })),
        }
      : {
          ...(selected
            ? { customerId: selected.id }
            : { customerName, customerPhone: phone, customerEmail: email, customerAddress: address }),
          brand,
          model,
          serial,
          note,
          devicePhotos: indexes(deviceFiles),
          parts: parts.map((part) => ({
            name: part.name,
            note: part.note,
            photos: indexes(part.files),
          })),
        };
    const body = new FormData();
    body.set("payload", JSON.stringify(payload));
    files.forEach((file, index) => body.append(`file${index}`, file));

    setSaving(true);
    setError(null);
    try {
      const saved = await request<Device>(editing ? `/api/services/${deviceId}` : "/api/services", {
        method: editing ? "PUT" : "POST",
        body,
      });
      onNavigate(`/services/${saved.id}`);
    } catch (err) {
      if (err instanceof ApiError && err.code === "unauthorized") {
        onUnauthorized();
        return;
      }
      if (err instanceof ApiError && err.code === "device_in_shop") {
        setShopServiceId(err.serviceId);
        setError(null);
      } else {
        setError(problemText(copy, err));
      }
      setSaving(false);
    }
  }

  if (!ready) {
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
    <form className="page form-page" onSubmit={(event) => void save(event)}>
      <header className="page-head">
        <div className="page-title">
          <div className="page-title-text">
            <h1>{editing ? copy.correct : copy.newDevice}</h1>
          </div>
        </div>
      </header>

      <section className="panel">
        <div className="panel-head">
          <div className="panel-title">
            <h2>{copy.customer}</h2>
          </div>
        </div>
        {!editing && selected ? (
          <div className="selected-customer">
            <div className="person">
              <div className="thing-text">
                <p className="muted">{copy.selectedCustomer}</p>
                <strong className="person-name">{selected.name}</strong>
                <span className="thing-sub">
                  {[selected.phone, selected.email, selected.address].filter(Boolean).join(" · ")}
                </span>
              </div>
            </div>
            <button className="button secondary" type="button" onClick={() => setSelected(null)}>
              {copy.clearCustomer}
            </button>
          </div>
        ) : (
          <div className="fields">
            <label className="field">
              {copy.search}
              <input
                value={customerDraft}
                placeholder={copy.searchCustomers}
                onChange={(event) => setCustomerDraft(event.target.value)}
              />
            </label>
            {results.length > 0 ? (
              <div className="results">
                {results.map((customer) => (
                  <button
                    className="result"
                    type="button"
                    key={customer.id}
                    onClick={() => {
                      if (editing) {
                        setCustomerLink(customer.id);
                        setCustomerName(customer.name);
                        setPhone(customer.phone);
                        setEmail(customer.email);
                        setAddress(customer.address);
                      } else {
                        setSelected(customer);
                      }
                      setResults([]);
                      setCustomerDraft("");
                    }}
                  >
                    <span className="thing-text">
                      <strong className="thing-title">{customer.name}</strong>
                      <span className="thing-sub">
                        {[customer.phone, customer.email, customer.address].filter(Boolean).join(" · ")}
                      </span>
                    </span>
                  </button>
                ))}
              </div>
            ) : null}
            {resultsMore ? <p className="muted">{copy.keepTyping}</p> : null}
            <div className="fields-2">
              <label className="field">
                {copy.customerName}
                <input value={customerName} onChange={(event) => setCustomerName(event.target.value)} />
              </label>
              <label className="field">
                {copy.phone}
                <input value={phone} inputMode="tel" autoComplete="tel" onChange={(event) => setPhone(event.target.value)} />
              </label>
              <label className="field">
                {copy.email}
                <input value={email} inputMode="email" autoComplete="email" onChange={(event) => setEmail(event.target.value)} />
              </label>
              <label className="field">
                {copy.address}
                <input value={address} autoComplete="street-address" onChange={(event) => setAddress(event.target.value)} />
              </label>
            </div>
            {editing ? (
              <button
                className="button secondary"
                type="button"
                onClick={() => {
                  setCustomerLink(null);
                  setCustomerName("");
                  setPhone("");
                  setEmail("");
                  setAddress("");
                }}
              >
                {copy.clearCustomer}
              </button>
            ) : null}
          </div>
        )}
      </section>

      <section className="panel">
        <div className="panel-head">
          <div className="panel-title">
            <h2>{copy.device}</h2>
          </div>
        </div>
        <div className="fields-2">
          <Suggest
            label={copy.brand}
            value={brand}
            options={catalog.map((item) => ({ name: item.brand, image: false }))}
            brand=""
            onChange={setBrand}
          />
          <Suggest
            label={copy.model}
            value={model}
            options={modelsFor(catalog, brand)}
            brand={brand}
            onChange={setModel}
          />
        </div>
        <label className="field">
          {copy.serial}
          <input value={serial} onChange={(event) => setSerial(event.target.value)} />
        </label>
        {!editing && shopServiceId ? (
          <p className="alert">
            {copy.deviceInShop}{" "}
            <button className="link" type="button" onClick={() => onNavigate(`/services/${shopServiceId}`)}>
              {copy.openService}
            </button>
          </p>
        ) : null}
        {!editing && earlier.items.length > 0 ? (
          <div className="earlier">
            <h3>{copy.earlierServices}</h3>
            <ul className="rows">
              {earlier.items.map((item) => (
                <li key={item.id}>
                  <div className="row row-static">
                    <span className="row-text">
                      <span className="row-title">{item.customerName}</span>
                      <span className="row-sub">
                        {formatWhen(item.createdAt, lang)} · {item.phone} · {statusLabel(item.status, states, lang)}
                      </span>
                    </span>
                  </div>
                </li>
              ))}
            </ul>
            <End enabled={earlier.hasMore && !earlier.loading} onEnd={earlier.more} />
          </div>
        ) : null}
        <label className="field">
          {copy.conditionNote}
          <textarea value={note} onChange={(event) => setNote(event.target.value)} />
        </label>
        <PhotoFields
          copy={copy}
          label={copy.photos}
          existing={deviceExisting}
          files={deviceFiles}
          deviceId={deviceId}
          onUnauthorized={onUnauthorized}
          onAdd={(list) => {
            const added = takeFiles(list);
            setDeviceFiles((current) => [...current, ...added]);
          }}
          onDropExisting={(id) => setDeviceExisting((current) => current.filter((photo) => photo.id !== id))}
          onDropFile={(file) => {
            dropFile(file);
            setDeviceFiles((current) => current.filter((item) => item.key !== file.key));
          }}
        />
      </section>

      <section className="panel">
        <div className="panel-head">
          <div className="panel-title">
            <h2>{copy.parts}</h2>
          </div>
          <div className="panel-actions">
            <button
              className="button secondary"
              type="button"
              onClick={() => {
                const key = localKey();
                setParts((current) => [...current, { key, id: "", name: "", note: "", existing: [], files: [] }]);
              }}
            >
              {copy.addPart}
            </button>
          </div>
        </div>
        {parts.length === 0 ? <p className="muted">{copy.noParts}</p> : null}
        {parts.map((part, index) => (
          <div className="part-card" key={part.key}>
            <div className="part-card-head">
              <h3>
                {copy.parts} {index + 1}
              </h3>
              <button
                className="button secondary small"
                type="button"
                onClick={() => {
                  part.files.forEach(dropFile);
                  setParts((current) => current.filter((item) => item.key !== part.key));
                }}
              >
                {copy.removePart}
              </button>
            </div>
            <label className="field">
              {copy.partName}
              <input
                value={part.name}
                onChange={(event) =>
                  setParts((current) =>
                    current.map((item) => (item.key === part.key ? { ...item, name: event.target.value } : item)),
                  )
                }
              />
            </label>
            <label className="field">
              {copy.partNote}
              <textarea
                value={part.note}
                onChange={(event) =>
                  setParts((current) =>
                    current.map((item) => (item.key === part.key ? { ...item, note: event.target.value } : item)),
                  )
                }
              />
            </label>
            <PhotoFields
              copy={copy}
              label={copy.photos}
              existing={part.existing}
              files={part.files}
              deviceId={deviceId}
              onUnauthorized={onUnauthorized}
              onAdd={(list) => {
                const added = takeFiles(list);
                setParts((current) =>
                  current.map((item) =>
                    item.key === part.key ? { ...item, files: [...item.files, ...added] } : item,
                  ),
                );
              }}
              onDropExisting={(id) =>
                setParts((current) =>
                  current.map((item) =>
                    item.key === part.key
                      ? { ...item, existing: item.existing.filter((photo) => photo.id !== id) }
                      : item,
                  ),
                )
              }
              onDropFile={(file) => {
                dropFile(file);
                setParts((current) =>
                  current.map((item) =>
                    item.key === part.key ? { ...item, files: item.files.filter((entry) => entry.key !== file.key) } : item,
                  ),
                );
              }}
            />
          </div>
        ))}
      </section>

      <div className="action-bar">
        {error ? <p className="alert">{error}</p> : null}
        <div className="action-bar-buttons">
          <button
            className="button secondary"
            type="button"
            onClick={() => onNavigate(editing && deviceId ? `/services/${deviceId}` : "/services")}
          >
            {copy.back}
          </button>
          <button className="button" type="submit" disabled={saving}>
            {copy.save}
          </button>
        </div>
      </div>
    </form>
  );
}

function Suggest({
  label,
  value,
  options,
  brand,
  onChange,
}: {
  label: string;
  value: string;
  options: CatalogModel[];
  brand: string;
  onChange: (value: string) => void;
}) {
  const id = useId();
  const listId = `${id}-list`;
  const root = useRef<HTMLDivElement>(null);
  const placed = useRef(false);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(0);
  const [box, setBox] = useState<SuggestBox | null>(null);
  const query = value.trim().toLowerCase();
  const shown = options.filter((option) => option.name.toLowerCase().includes(query));
  const index = shown.length === 0 ? 0 : Math.min(active, shown.length - 1);

  useEffect(() => {
    setActive(0);
  }, [query]);

  useLayoutEffect(() => {
    const anchor = root.current?.querySelector("input");
    if (!open || !(anchor instanceof HTMLElement)) {
      placed.current = false;
      setBox(null);
      return;
    }
    if (!placed.current) {
      placed.current = true;
      const rect = anchor.getBoundingClientRect();
      if (window.innerHeight - rect.bottom < 200) {
        anchor.scrollIntoView({ block: "center", inline: "nearest" });
      }
    }
    function place() {
      if (!(anchor instanceof HTMLElement)) {
        return;
      }
      const next = suggestBox(anchor);
      setBox((current) => (sameBox(current, next) ? current : next));
    }
    place();
    window.addEventListener("scroll", place, true);
    window.addEventListener("resize", place);
    return () => {
      window.removeEventListener("scroll", place, true);
      window.removeEventListener("resize", place);
    };
  }, [open, shown.length]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const option = document.getElementById(listId)?.querySelector<HTMLElement>(".suggest-option[aria-selected='true']");
    const menu = option?.parentElement;
    if (!option || !menu) {
      return;
    }
    const top = option.offsetTop;
    const bottom = top + option.offsetHeight;
    if (top < menu.scrollTop) {
      menu.scrollTop = top;
    } else if (bottom > menu.scrollTop + menu.clientHeight) {
      menu.scrollTop = bottom - menu.clientHeight;
    }
  }, [listId, open, index, box]);

  function choose(name: string) {
    onChange(name);
    setOpen(false);
  }

  return (
    <div className="field suggest-field" ref={root}>
      <label htmlFor={id}>{label}</label>
      <input
        id={id}
        role="combobox"
        aria-expanded={open && shown.length > 0}
        aria-controls={listId}
        aria-autocomplete="list"
        autoComplete="off"
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => setOpen(false)}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown") {
            event.preventDefault();
            if (!open || shown.length === 0) {
              setOpen(true);
              return;
            }
            setActive((index + 1) % shown.length);
            return;
          }
          if (event.key === "ArrowUp") {
            event.preventDefault();
            if (!open || shown.length === 0) {
              setOpen(true);
              return;
            }
            setActive((index - 1 + shown.length) % shown.length);
            return;
          }
          if (event.key === "Escape") {
            setOpen(false);
            return;
          }
          if (event.key === "Enter" && open && shown[index] && shown[index].name !== value) {
            event.preventDefault();
            choose(shown[index].name);
          }
        }}
      />
      {open && shown.length > 0 && box
        ? createPortal(
            <div
              className="suggest"
              id={listId}
              role="listbox"
              style={{
                top: box.top,
                bottom: box.bottom,
                left: box.left,
                width: box.width,
                maxHeight: box.maxHeight,
              }}
            >
              {shown.map((option, optionIndex) => (
                <button
                  className="suggest-option"
                  type="button"
                  role="option"
                  tabIndex={-1}
                  aria-selected={optionIndex === index}
                  data-current={option.name.toLowerCase() === query ? true : undefined}
                  key={option.name}
                  onMouseDown={(event) => event.preventDefault()}
                  onMouseEnter={() => setActive(optionIndex)}
                  onClick={() => choose(option.name)}
                >
                  {option.image ? <ModelIcon brand={brand} model={option.name} /> : null}
                  {option.name}
                </button>
              ))}
            </div>,
            document.body,
          )
        : null}
    </div>
  );
}

type SuggestBox = {
  top: number | "auto";
  bottom: number | "auto";
  left: number;
  width: number;
  maxHeight: number;
};

function sameBox(current: SuggestBox | null, next: SuggestBox) {
  return (
    current !== null &&
    current.top === next.top &&
    current.bottom === next.bottom &&
    current.left === next.left &&
    current.width === next.width &&
    current.maxHeight === next.maxHeight
  );
}

function suggestBox(anchor: HTMLElement): SuggestBox {
  const rect = anchor.getBoundingClientRect();
  const gap = 4;
  const spaceBelow = window.innerHeight - rect.bottom - gap - 8;
  const spaceAbove = rect.top - gap - 8;
  const below = spaceBelow >= 96 || spaceBelow >= spaceAbove;
  const maxHeight = Math.min(256, Math.max(120, below ? spaceBelow : spaceAbove));
  return {
    top: below ? rect.bottom + gap : "auto",
    bottom: below ? "auto" : window.innerHeight - rect.top + gap,
    left: rect.left,
    width: rect.width,
    maxHeight,
  };
}

function localKey() {
  const bytes = crypto.getRandomValues(new Uint8Array(16));
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("");
}

function modelsFor(catalog: CatalogBrand[], brand: string) {
  const key = brand.trim().toLowerCase();
  return catalog.find((item) => item.brand.toLowerCase() === key)?.models ?? [];
}

function PhotoFields({
  copy,
  label,
  existing,
  files,
  deviceId,
  onAdd,
  onDropExisting,
  onDropFile,
  onUnauthorized,
}: {
  copy: Copy;
  label: string;
  existing: ExistingPhoto[];
  files: LocalFile[];
  deviceId?: string;
  onAdd: (files: File[]) => void;
  onDropExisting: (id: string) => void;
  onDropFile: (file: LocalFile) => void;
  onUnauthorized: () => void;
}) {
  return (
    <div className="fields">
      <span className="field-label">{label}</span>
      <div className="photo-grid">
        {existing.map((photo) => (
          <div className="photo-tile" key={photo.id}>
            {deviceId ? (
              <PhotoButton
                copy={copy}
                src={photoSrc(deviceId, photo.id)}
                name={photo.id}
                onUnauthorized={onUnauthorized}
              />
            ) : null}
            <button className="photo-remove" type="button" aria-label={copy.removePhoto} onClick={() => onDropExisting(photo.id)}>
              <Icon name="close" size={12} />
            </button>
          </div>
        ))}
        {files.map((file) => (
          <div className="photo-tile" key={file.key}>
            <PhotoButton copy={copy} src={file.url} name={file.file.name} onUnauthorized={onUnauthorized} />
            <button className="photo-remove" type="button" aria-label={copy.removePhoto} onClick={() => onDropFile(file)}>
              <Icon name="close" size={12} />
            </button>
          </div>
        ))}
        <label className="photo-add">
          <Icon name="imagePlus" size={18} />
          <span>{copy.photos}</span>
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            multiple
            onChange={(event) => {
              const chosen = Array.from(event.target.files ?? []);
              onAdd(chosen);
              event.target.value = "";
            }}
          />
        </label>
      </div>
    </div>
  );
}

