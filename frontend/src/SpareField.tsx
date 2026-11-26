import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { ApiError, request } from "./api";
import { fold } from "./fold";
import { usePaused } from "./list";

type Props = {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  brand?: string;
  onUnauthorized: () => void;
};

type Box = {
  top: number | "auto";
  bottom: number | "auto";
  left: number;
  width: number;
  maxHeight: number;
};

export function SpareField({ id, label, value, onChange, brand, onUnauthorized }: Props) {
  const root = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(0);
  const [options, setOptions] = useState<string[]>([]);
  const [box, setBox] = useState<Box | null>(null);
  const [paused] = usePaused(value);
  const [pausedBrand] = usePaused(brand ?? "");
  const query = fold(value);
  const shown = options.filter((name) => fold(name).includes(query));
  const index = shown.length === 0 ? 0 : Math.min(active, shown.length - 1);
  const listId = `${id}-list`;

  useEffect(() => {
    setActive(0);
  }, [query]);

  useEffect(() => {
    const controller = new AbortController();
    const params = new URLSearchParams();
    params.set("q", paused);
    let path = `/api/spare-parts/suggestions/brands?${params}`;
    if (brand !== undefined) {
      params.set("brand", pausedBrand);
      path = `/api/spare-parts/suggestions/models?${params}`;
    }
    request<{ items: string[] }>(path, { signal: controller.signal })
      .then((body) => {
        if (!controller.signal.aborted) {
          setOptions(body.items);
        }
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        setOptions([]);
        if (err instanceof ApiError && err.code === "unauthorized") {
          onUnauthorized();
        }
      });
    return () => controller.abort();
  }, [brand, onUnauthorized, paused, pausedBrand]);

  useLayoutEffect(() => {
    const anchor = root.current?.querySelector("input");
    if (!open || !(anchor instanceof HTMLElement) || shown.length === 0) {
      setBox(null);
      return;
    }
    function place() {
      if (!(anchor instanceof HTMLElement)) {
        return;
      }
      const rect = anchor.getBoundingClientRect();
      const gap = 4;
      const spaceBelow = window.innerHeight - rect.bottom - gap - 8;
      const spaceAbove = rect.top - gap - 8;
      const below = spaceBelow >= 96 || spaceBelow >= spaceAbove;
      setBox({
        top: below ? rect.bottom + gap : "auto",
        bottom: below ? "auto" : window.innerHeight - rect.top + gap,
        left: rect.left,
        width: rect.width,
        maxHeight: Math.min(256, Math.max(120, below ? spaceBelow : spaceAbove)),
      });
    }
    place();
    window.addEventListener("scroll", place, true);
    window.addEventListener("resize", place);
    return () => {
      window.removeEventListener("scroll", place, true);
      window.removeEventListener("resize", place);
    };
  }, [open, shown.length, value]);

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
          if (event.key === "ArrowDown" || event.key === "ArrowUp") {
            event.preventDefault();
            if (!open || shown.length === 0) {
              setOpen(true);
              return;
            }
            const step = event.key === "ArrowDown" ? 1 : -1;
            setActive((index + step + shown.length) % shown.length);
            return;
          }
          if (event.key === "Escape") {
            setOpen(false);
            return;
          }
          if (event.key === "Enter" && open && shown[index] && shown[index] !== value) {
            event.preventDefault();
            choose(shown[index]);
          }
        }}
      />
      {open && shown.length > 0 && box
        ? createPortal(
            <div
              className="suggest"
              id={listId}
              role="listbox"
              style={{ top: box.top, bottom: box.bottom, left: box.left, width: box.width, maxHeight: box.maxHeight }}
            >
              {shown.map((name, optionIndex) => (
                <button
                  className="suggest-option"
                  type="button"
                  role="option"
                  tabIndex={-1}
                  aria-selected={optionIndex === index}
                  key={name}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => choose(name)}
                >
                  {name}
                </button>
              ))}
            </div>,
            document.body,
          )
        : null}
    </div>
  );
}
