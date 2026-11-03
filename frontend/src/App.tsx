import { useEffect, useState } from "react";
import { logoSrc, request, type Brand, type User } from "./api";
import { Audit } from "./Audit";
import { CustomerDetail } from "./CustomerDetail";
import { Customers } from "./Customers";
import { DeviceDetail } from "./DeviceDetail";
import { DeviceForm } from "./DeviceForm";
import { Devices } from "./Devices";
import { Home } from "./Home";
import { shopName, text, writeLang, type Lang } from "./i18n";
import { Login } from "./Login";
import { Shell, type NavItem } from "./Shell";
import { Users } from "./Users";

export function App() {
  const [lang, setLang] = useState<Lang>(() => {
    const value = localStorage.getItem("techservice.lang");
    return value === "tr" ? "tr" : "en";
  });
  const [user, setUser] = useState<User | null>(null);
  const [brand, setBrand] = useState<Brand>({ name: "", logo: false });
  const [ready, setReady] = useState(false);
  const [location, setLocation] = useState(() => window.location.pathname + window.location.search);

  const copy = text(lang);
  const url = new URL(location, window.location.origin);
  const path = url.pathname;

  function changeLang(next: Lang) {
    writeLang(next);
    setLang(next);
  }

  useEffect(() => {
    document.documentElement.lang = lang;
  }, [lang]);

  useEffect(() => {
    document.title = shopName(copy, brand.name);
  }, [brand.name, copy.appName]);

  useEffect(() => {
    const current = document.querySelector<HTMLLinkElement>("link[rel='icon']");
    if (!brand.logo) {
      current?.remove();
      return;
    }
    const link = current ?? document.createElement("link");
    link.rel = "icon";
    link.href = logoSrc;
    if (!current) {
      document.head.appendChild(link);
    }
  }, [brand.logo]);

  useEffect(() => {
    const onPop = () => setLocation(window.location.pathname + window.location.search);
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    request<User>("/api/session", { signal: controller.signal })
      .then((session) => setUser(session))
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") {
          return;
        }
        setUser(null);
      })
      .finally(() => setReady(true));
    return () => controller.abort();
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    request<Brand>("/api/brand", { signal: controller.signal })
      .then((loaded) => setBrand(loaded))
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") {
          return;
        }
        setBrand({ name: "", logo: false });
      });
    return () => controller.abort();
  }, []);

  function go(next: string) {
    window.history.pushState({}, "", next);
    setLocation(window.location.pathname + window.location.search);
  }

  async function signOut() {
    try {
      await request("/api/logout", { method: "POST" });
    } finally {
      setUser(null);
      go("/");
    }
  }

  if (!ready) {
    return (
      <main className="auth">
        <p className="muted">{copy.checking}</p>
      </main>
    );
  }

  if (!user) {
    return <Login copy={copy} lang={lang} brand={brand} onLang={changeLang} onSuccess={setUser} />;
  }

  const items: NavItem[] = [
    { path: "/", label: copy.home, icon: "home" },
    { path: "/devices", label: copy.devices, icon: "devices" },
    { path: "/customers", label: copy.customers, icon: "customers" },
  ];
  if (user.role === "admin" || user.role === "owner") {
    items.push({ path: "/audit", label: copy.audit, icon: "audit" });
  }
  if (user.role === "admin") {
    items.push({ path: "/users", label: copy.users, icon: "users" });
  }

  const unauthorized = () => setUser(null);
  const deviceEdit = path.match(/^\/devices\/([0-9a-f-]{36})\/edit$/);
  const device = path.match(/^\/devices\/([0-9a-f-]{36})$/);
  const customer = path.match(/^\/customers\/(\d+)$/);
  const audit = path === "/audit" && user.role !== "technician";
  const usersPage = path === "/users" && user.role === "admin";

  let screen = <Home copy={copy} user={user} />;
  if (usersPage) {
    screen = <Users copy={copy} onUnauthorized={unauthorized} />;
  } else if (path === "/devices") {
    screen = <Devices copy={copy} lang={lang} onNavigate={go} onUnauthorized={unauthorized} />;
  } else if (path === "/devices/new") {
    screen = (
      <DeviceForm
        copy={copy}
        mode="create"
        customerId={url.searchParams.get("customer")}
        onNavigate={go}
        onUnauthorized={unauthorized}
      />
    );
  } else if (deviceEdit) {
    screen = (
      <DeviceForm
        copy={copy}
        mode="edit"
        deviceId={deviceEdit[1]}
        onNavigate={go}
        onUnauthorized={unauthorized}
      />
    );
  } else if (device) {
    screen = <DeviceDetail copy={copy} lang={lang} id={device[1]} onNavigate={go} onUnauthorized={unauthorized} />;
  } else if (path === "/customers") {
    screen = <Customers copy={copy} onNavigate={go} onUnauthorized={unauthorized} />;
  } else if (customer) {
    screen = (
      <CustomerDetail copy={copy} lang={lang} id={customer[1]} onNavigate={go} onUnauthorized={unauthorized} />
    );
  } else if (audit) {
    screen = <Audit copy={copy} lang={lang} onUnauthorized={unauthorized} />;
  }

  return (
    <Shell
      copy={copy}
      lang={lang}
      user={user}
      path={audit || usersPage || path === "/" || path.startsWith("/devices") || path.startsWith("/customers") ? path : "/"}
      items={items}
      brand={brand}
      onLang={changeLang}
      onNavigate={go}
      onSignOut={() => void signOut()}
    >
      {screen}
    </Shell>
  );
}
