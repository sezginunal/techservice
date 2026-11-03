import { useEffect, useId, useRef, useState } from "react";
import type { Copy, Lang } from "./i18n";

type Props = {
  copy: Copy;
  lang: Lang;
  label: string;
  value: string;
  onChange: (value: string) => void;
};

// en-GB keeps English names and day-month-year order. en-US would show month-day-year.
function localeOf(lang: Lang) {
  return lang === "tr" ? "tr-TR" : "en-GB";
}

function parseIso(value: string) {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (!match) {
    return null;
  }
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const date = new Date(year, month - 1, day);
  if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) {
    return null;
  }
  return date;
}

function toIso(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}

function sameDay(a: Date, b: Date) {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

function mondayIndex(date: Date) {
  return (date.getDay() + 6) % 7;
}

function monthDays(view: Date) {
  const first = new Date(view.getFullYear(), view.getMonth(), 1);
  const start = new Date(view.getFullYear(), view.getMonth(), 1 - mondayIndex(first));
  return Array.from({ length: 42 }, (_, index) => new Date(start.getFullYear(), start.getMonth(), start.getDate() + index));
}

function formatDay(value: string, lang: Lang) {
  const date = parseIso(value);
  if (!date) {
    return "";
  }
  return new Intl.DateTimeFormat(localeOf(lang), {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(date);
}

export function DateField({ copy, lang, label, value, onChange }: Props) {
  const locale = localeOf(lang);
  const dialogId = useId();
  const root = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const [open, setOpen] = useState(false);
  const [view, setView] = useState(() => monthOf(parseIso(value) ?? new Date()));
  const selected = parseIso(value);
  const today = new Date();

  useEffect(() => {
    if (!open) {
      return;
    }
    function onPointer(event: PointerEvent) {
      if (!root.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
        trigger.current?.focus();
      }
    }
    document.addEventListener("pointerdown", onPointer);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("pointerdown", onPointer);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  function toggle() {
    if (open) {
      setOpen(false);
      return;
    }
    setView(monthOf(selected ?? new Date()));
    setOpen(true);
  }

  function pick(day: Date) {
    onChange(toIso(day));
    setOpen(false);
    trigger.current?.focus();
  }

  function clear() {
    onChange("");
    setOpen(false);
    trigger.current?.focus();
  }

  const weekdays = weekdayLabels(locale);
  const days = monthDays(view);
  const monthLabel = new Intl.DateTimeFormat(locale, { month: "long", year: "numeric" }).format(view);
  const dayLabel = new Intl.DateTimeFormat(locale, { day: "numeric", month: "long", year: "numeric" });

  return (
    <div className="date-field" ref={root} lang={locale}>
      <label className="field">
        {label}
        <button
          ref={trigger}
          className="date-field-value"
          type="button"
          aria-haspopup="dialog"
          aria-expanded={open}
          aria-controls={dialogId}
          onClick={toggle}
        >
          <span className={value ? undefined : "date-placeholder"}>{value ? formatDay(value, lang) : copy.datePlaceholder}</span>
          <CalendarIcon />
        </button>
      </label>
      {open ? (
        <div className="date-pop" id={dialogId} role="dialog" aria-label={copy.chooseDate}>
          <div className="date-pop-head">
            <button className="date-nav" type="button" aria-label={copy.previousMonth} onClick={() => setView(shiftMonth(view, -1))}>
              <Chevron direction="left" />
            </button>
            <strong>{monthLabel}</strong>
            <button className="date-nav" type="button" aria-label={copy.nextMonth} onClick={() => setView(shiftMonth(view, 1))}>
              <Chevron direction="right" />
            </button>
          </div>
          <div className="date-weekdays" aria-hidden="true">
            {weekdays.map((name, index) => (
              <span key={index}>{name}</span>
            ))}
          </div>
          <div className="date-grid">
            {days.map((day) => {
              const pressed = selected ? sameDay(day, selected) : false;
              const outside = day.getMonth() !== view.getMonth();
              const current = sameDay(day, today);
              return (
                <button
                  key={toIso(day)}
                  className={["date-day", outside ? "outside" : "", current ? "today" : ""].filter(Boolean).join(" ")}
                  type="button"
                  aria-pressed={pressed}
                  aria-label={dayLabel.format(day)}
                  onClick={() => pick(day)}
                >
                  {day.getDate()}
                </button>
              );
            })}
          </div>
          {value ? (
            <button className="button secondary date-clear" type="button" onClick={clear}>
              {copy.clearDate}
            </button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function monthOf(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

function shiftMonth(view: Date, amount: number) {
  return new Date(view.getFullYear(), view.getMonth() + amount, 1);
}

function weekdayLabels(locale: string) {
  const format = new Intl.DateTimeFormat(locale, { weekday: "short" });
  const monday = new Date(2024, 0, 1);
  return Array.from({ length: 7 }, (_, index) => format.format(new Date(2024, 0, monday.getDate() + index)));
}

function CalendarIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <rect x="3" y="5" width="18" height="16" rx="2" />
      <path d="M3 10h18" />
      <path d="M8 3v4" />
      <path d="M16 3v4" />
    </svg>
  );
}

function Chevron({ direction }: { direction: "left" | "right" }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      {direction === "left" ? <path d="M15 6l-6 6 6 6" /> : <path d="M9 6l6 6-6 6" />}
    </svg>
  );
}
