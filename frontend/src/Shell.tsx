import { useEffect, useState, type ReactNode } from "react";
import { logoSrc, type Brand, type User } from "./api";
import { shopName, type Copy, type Lang } from "./i18n";
import { Icon, type IconName } from "./icons";
import { LanguageSwitch } from "./LanguageSwitch";
import { appVersion } from "./version";

export type NavIcon = Extract<IconName, "home" | "users" | "services" | "customers" | "spare" | "audit" | "settings">;

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

const primaryPaths = ["/", "/services", "/customers", "/spare-parts"] as const;

export function Shell({ copy, lang, user, brand, path, items, onLang, onNavigate, onSignOut, children }: ShellProps) {
  const [open, setOpen] = useState(false);
  const primary = items.filter((item) => (primaryPaths as readonly string[]).includes(item.path));
  const secondary = items.filter((item) => !(primaryPaths as readonly string[]).includes(item.path));
  const moreActive = secondary.some((item) => current(path, item.path));

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
    <div className="app" data-menu={open ? "open" : undefined}>
      <aside className="sidebar">
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
              <Icon name={item.icon} size={18} />
              {item.label}
            </button>
          ))}
        </nav>
        <div className="sidebar-foot">
          <div className="account">
            <span className="avatar" data-tone="blue" aria-hidden="true">
              {initials(user.username)}
            </span>
            <span className="account-text">
              <strong>{user.username}</strong>
              <span>{copy[user.role]}</span>
            </span>
            <button className="icon-button" type="button" aria-label={copy.signOut} onClick={onSignOut}>
              <Icon name="logout" size={18} />
            </button>
          </div>
          <div className="sidebar-tools">
            <LanguageSwitch copy={copy} lang={lang} onLang={onLang} />
            <AppVersion copy={copy} />
          </div>
        </div>
      </aside>
      {open ? (
        <button className="backdrop" type="button" aria-label={copy.closeMenu} onClick={() => setOpen(false)} />
      ) : null}
      <div className="inset">
        <header className="topbar">
          <button className="brand" type="button" onClick={() => onNavigate("/")}>
            <BrandMark copy={copy} brand={brand} />
          </button>
          <div className="topbar-end">
            <AppVersion copy={copy} />
            <LanguageSwitch copy={copy} lang={lang} onLang={onLang} />
          </div>
        </header>
        <main className="content">{children}</main>
        <nav className="tabbar" aria-label={copy.general}>
          {primary.map((item) => (
            <button
              key={item.path}
              className="tab"
              type="button"
              aria-current={current(path, item.path) ? "page" : undefined}
              onClick={() => onNavigate(item.path)}
            >
              <span className="tab-icon">
                <Icon name={item.icon} size={18} />
              </span>
              <span className="tab-label">{item.label}</span>
            </button>
          ))}
          <button
            className="tab"
            type="button"
            aria-expanded={open}
            aria-current={moreActive && !open ? "page" : undefined}
            onClick={() => setOpen((value) => !value)}
          >
            <span className="tab-icon">
              <Icon name={open ? "close" : "menu"} size={18} />
            </span>
            <span className="tab-label">{copy.more}</span>
          </button>
        </nav>
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
      {brand.logo ? <img className="mark" src={logoSrc} alt="" /> : <DefaultMark />}
      <span className="brand-name">
        <strong>{shopName(copy, brand.name)}</strong>
        <span>{copy.tagline}</span>
      </span>
    </>
  );
}

function DefaultMark() {
  return (
    <span className="mark" aria-hidden="true">
      <Icon name="services" size={16} />
    </span>
  );
}
