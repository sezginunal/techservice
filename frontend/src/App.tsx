import { useEffect, useState } from "react";
import { logoSrc, request, type Brand, type ServiceState, type Shop, type User } from "./api";
import { Audit } from "./Audit";
import { CustomerDetail } from "./CustomerDetail";
import { Customers } from "./Customers";
import { DeviceDetail } from "./DeviceDetail";
import { DeviceForm } from "./DeviceForm";
import { DevicePage } from "./DevicePage";
import { Devices } from "./Devices";
import { Home } from "./Home";
import { shopName, storedLang, text, writeLang, type Lang } from "./i18n";
import { Login } from "./Login";
import { Settings } from "./Settings";
import { Shell, type NavItem } from "./Shell";
import { SparePartPage } from "./SparePartPage";
import { SpareParts } from "./SpareParts";
import { Users } from "./Users";

export function App() {
  const [stored] = useState<Lang | null>(storedLang);
  const [lang, setLang] = useState<Lang>(stored ?? "en");
  const [user, setUser] = useState<User | null>(null);
  const [states, setStates] = useState<ServiceState[]>([]);
  const [brand, setBrand] = useState<Brand>({ name: "", logo: false });
  const [ready, setReady] = useState(false);
  const [langReady, setLangReady] = useState(stored !== null);
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
    if (!user) {
      setStates([]);
      return;
    }
    const controller = new AbortController();
    request<ServiceState[]>("/api/states", { signal: controller.signal })
      .then((items) => {
        if (!controller.signal.aborted) {
          setStates(items);
        }
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") {
          return;
        }
        if (!controller.signal.aborted) {
          setStates([]);
        }
      });
    return () => controller.abort();
  }, [user]);

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
    request<Shop>("/api/brand", { signal: controller.signal })
      .then((loaded) => {
        setBrand({ name: loaded.name, logo: loaded.logo });
        if (stored === null) {
          setLang(loaded.lang === "tr" ? "tr" : "en");
        }
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") {
          return;
        }
        setBrand({ name: "", logo: false });
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setLangReady(true);
        }
      });
    return () => controller.abort();
  }, [stored]);

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

  if (!langReady) {
    return <main className="auth-loading" />;
  }

  if (!ready) {
    return (
      <main className="auth-loading">
        <span className="spinner" aria-hidden="true" />
        <p>{copy.checking}</p>
      </main>
    );
  }

  if (!user) {
    return <Login copy={copy} lang={lang} brand={brand} onLang={changeLang} onSuccess={setUser} />;
  }

  const items: NavItem[] = [
    { path: "/", label: copy.home, icon: "home" },
    { path: "/services", label: copy.devices, icon: "services" },
    { path: "/customers", label: copy.customers, icon: "customers" },
    { path: "/spare-parts", label: copy.spareParts, icon: "spare" },
  ];
  if (user.role === "admin" || user.role === "owner") {
    items.push({ path: "/audit", label: copy.audit, icon: "audit" });
    items.push({ path: "/settings", label: copy.settings, icon: "settings" });
  }
  if (user.role === "admin") {
    items.push({ path: "/users", label: copy.users, icon: "users" });
  }

  const unauthorized = () => setUser(null);
  const serviceEdit = path.match(/^\/services\/([0-9a-f-]{36})\/edit$/);
  const service = path.match(/^\/services\/([0-9a-f-]{36})$/);
  const device = path.match(/^\/devices\/([0-9a-f-]{36})$/);
  const customer = path.match(/^\/customers\/(\d+)$/);
  const sparePart = path.match(/^\/spare-parts\/([0-9a-f-]{36})$/);
  const audit = path === "/audit" && user.role !== "technician";
  const settingsPage = path === "/settings" && user.role !== "technician";
  const usersPage = path === "/users" && user.role === "admin";

  let screen = <Home copy={copy} lang={lang} states={states} onNavigate={go} onUnauthorized={unauthorized} />;
  if (usersPage) {
    screen = <Users copy={copy} onUnauthorized={unauthorized} />;
  } else if (path === "/services") {
    screen = (
      <Devices
        copy={copy}
        lang={lang}
        role={user.role}
        states={states}
        status={url.searchParams.get("status") ?? ""}
        onNavigate={go}
        onUnauthorized={unauthorized}
      />
    );
  } else if (path === "/services/new") {
    screen = (
      <DeviceForm
        copy={copy}
        lang={lang}
        states={states}
        mode="create"
        customerId={url.searchParams.get("customer")}
        onNavigate={go}
        onUnauthorized={unauthorized}
      />
    );
  } else if (serviceEdit) {
    screen = (
      <DeviceForm
        copy={copy}
        lang={lang}
        states={states}
        mode="edit"
        deviceId={serviceEdit[1]}
        onNavigate={go}
        onUnauthorized={unauthorized}
      />
    );
  } else if (service) {
    screen = (
      <DeviceDetail
        copy={copy}
        lang={lang}
        brand={brand}
        states={states}
        user={user}
        id={service[1]}
        onNavigate={go}
        onUnauthorized={unauthorized}
      />
    );
  } else if (device) {
    screen = (
      <DevicePage copy={copy} lang={lang} states={states} id={device[1]} onNavigate={go} onUnauthorized={unauthorized} />
    );
  } else if (sparePart) {
    screen = (
      <SparePartPage
        copy={copy}
        lang={lang}
        states={states}
        id={sparePart[1]}
        notice={url.searchParams.get("notice") ?? ""}
        onNavigate={go}
        onUnauthorized={unauthorized}
      />
    );
  } else if (path === "/spare-parts") {
    screen = <SpareParts copy={copy} lang={lang} role={user.role} onNavigate={go} onUnauthorized={unauthorized} />;
  } else if (path === "/customers") {
    screen = <Customers copy={copy} lang={lang} role={user.role} onNavigate={go} onUnauthorized={unauthorized} />;
  } else if (customer) {
    screen = (
      <CustomerDetail
        copy={copy}
        lang={lang}
        states={states}
        id={customer[1]}
        onNavigate={go}
        onUnauthorized={unauthorized}
      />
    );
  } else if (audit) {
    screen = <Audit copy={copy} lang={lang} states={states} onUnauthorized={unauthorized} />;
  } else if (settingsPage) {
    screen = <Settings copy={copy} lang={lang} states={states} onStates={setStates} onUnauthorized={unauthorized} />;
  }

  return (
    <Shell
      copy={copy}
      lang={lang}
      user={user}
      path={
        audit ||
        settingsPage ||
        usersPage ||
        path === "/" ||
        path.startsWith("/services") ||
        path.startsWith("/devices") ||
        path.startsWith("/customers") ||
        path.startsWith("/spare-parts")
          ? path
          : "/"
      }
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
