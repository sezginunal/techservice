import { useEffect, useState, type ReactNode } from "react";
import { logoSrc, type Brand, type User } from "./api";
import { shopName, type Copy, type Lang } from "./i18n";
import { LanguageSwitch } from "./LanguageSwitch";
import { appVersion } from "./version";

export type NavIcon = "home" | "users" | "devices" | "customers" | "audit";

export type NavItem = {
  path: string;
  label: string;
  icon: NavIcon;
};

type ShellProps = {
  copy: Copy;
  lang: Lang;
  user: User;
  brand: Brand;
  path: string;
  items: NavItem[];
  onLang: (lang: Lang) => void;
  onNavigate: (path: string) => void;
  onSignOut: () => void;
  children: ReactNode;
};

export function Shell({ copy, lang, user, brand, path, items, onLang, onNavigate, onSignOut, children }: ShellProps) {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    setOpen(false);
  }, [path]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setOpen(false);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open]);

  return (
    <div className="app">
      <aside className="sidebar" data-open={open ? "true" : "false"}>
        <button className="brand" type="button" onClick={() => onNavigate("/")}>
          <BrandMark copy={copy} brand={brand} />
        </button>
        <nav className="nav" aria-label={copy.general}>
          <p className="nav-label">{copy.general}</p>
          {items.map((item) => (
            <button
              key={item.path}
              className="nav-item"
              type="button"
              aria-current={current(path, item.path) ? "page" : undefined}
              onClick={() => onNavigate(item.path)}
            >
              <Icon name={item.icon} />
              {item.label}
            </button>
          ))}
        </nav>
        <div className="sidebar-foot">
          <div className="account">
            <span className="avatar" aria-hidden="true">
              {initials(user.username)}
            </span>
            <span className="account-text">
              <strong>{user.username}</strong>
              <span>{copy[user.role]}</span>
            </span>
          </div>
          <button className="button secondary" type="button" onClick={onSignOut}>
            {copy.signOut}
          </button>
        </div>
      </aside>
      {open ? (
        <button className="backdrop" type="button" aria-label={copy.closeMenu} onClick={() => setOpen(false)} />
      ) : null}
      <div className="inset">
        <header className="topbar">
          <button
            className="icon-button"
            type="button"
            aria-expanded={open}
            aria-label={open ? copy.closeMenu : copy.menu}
            onClick={() => setOpen((value) => !value)}
          >
            <Icon name="menu" />
          </button>
          <div className="topbar-end">
            <AppVersion copy={copy} />
            <LanguageSwitch copy={copy} lang={lang} onLang={onLang} />
          </div>
        </header>
        <main className="content">{children}</main>
      </div>
    </div>
  );
}

function current(path: string, itemPath: string) {
  if (itemPath === "/") {
    return path === "/";
  }
  return path === itemPath || path.startsWith(`${itemPath}/`);
}

export function initials(username: string) {
  const trimmed = username.trim();
  return trimmed.slice(0, 2).toLocaleUpperCase();
}

export function AppVersion({ copy }: { copy: Copy }) {
  return (
    <p className="app-version">
      {copy.version} {appVersion}
    </p>
  );
}

export function BrandMark({ copy, brand }: { copy: Copy; brand: Brand }) {
  return (
    <>
      {brand.logo ? <img className="mark" src={logoSrc} alt="" /> : <Mark />}
      <span className="brand-name">
        <strong>{shopName(copy, brand.name)}</strong>
        <span>{copy.tagline}</span>
      </span>
    </>
  );
}

function Mark() {
  return (
    <span className="mark" aria-hidden="true">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M14.7 6.3a4 4 0 0 0-5.4 5.4L4 17l3 3 5.3-5.3a4 4 0 0 0 5.4-5.4L15 12l-3-3 2.7-2.7z" />
      </svg>
    </span>
  );
}

function Icon({ name }: { name: NavIcon | "menu" }) {
  const props = {
    width: 18,
    height: 18,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 2,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true as const,
  };
  if (name === "menu") {
    return (
      <svg {...props}>
        <path d="M4 7h16" />
        <path d="M4 12h16" />
        <path d="M4 17h16" />
      </svg>
    );
  }
  if (name === "users") {
    return (
      <svg {...props}>
        <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="3" />
        <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
      </svg>
    );
  }
  if (name === "devices") {
    return (
      <svg {...props}>
        <rect x="4" y="4" width="16" height="12" rx="2" />
        <path d="M8 20h8" />
        <path d="M12 16v4" />
      </svg>
    );
  }
  if (name === "customers") {
    return (
      <svg {...props}>
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
        <circle cx="12" cy="7" r="3" />
      </svg>
    );
  }
  if (name === "audit") {
    return (
      <svg {...props}>
        <path d="M8 6h13" />
        <path d="M8 12h13" />
        <path d="M8 18h13" />
        <path d="M3 6h.01" />
        <path d="M3 12h.01" />
        <path d="M3 18h.01" />
      </svg>
    );
  }
  return (
    <svg {...props}>
      <rect x="3" y="3" width="7" height="9" rx="1" />
      <rect x="14" y="3" width="7" height="5" rx="1" />
      <rect x="14" y="12" width="7" height="9" rx="1" />
      <rect x="3" y="16" width="7" height="5" rx="1" />
    </svg>
  );
}
