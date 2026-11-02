import { useEffect, useState } from "react";
import { request, type User } from "./api";
import { Home } from "./Home";
import { text, writeLang, type Lang } from "./i18n";
import { Login } from "./Login";
import { Shell, type NavItem } from "./Shell";
import { Users } from "./Users";

export function App() {
  const [lang, setLang] = useState<Lang>(() => {
    const value = localStorage.getItem("techservice.lang");
    return value === "tr" ? "tr" : "en";
  });
  const [user, setUser] = useState<User | null>(null);
  const [ready, setReady] = useState(false);
  const [path, setPath] = useState(window.location.pathname);

  const copy = text(lang);

  function changeLang(next: Lang) {
    writeLang(next);
    setLang(next);
  }

  useEffect(() => {
    document.documentElement.lang = lang;
  }, [lang]);

  useEffect(() => {
    const onPop = () => setPath(window.location.pathname);
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

  function go(next: string) {
    window.history.pushState({}, "", next);
    setPath(next);
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
    return <Login copy={copy} lang={lang} onLang={changeLang} onSuccess={setUser} />;
  }

  const items: NavItem[] = [{ path: "/", label: copy.home, icon: "home" }];
  if (user.role === "admin") {
    items.push({ path: "/users", label: copy.users, icon: "users" });
  }

  const usersPage = path === "/users" && user.role === "admin";

  return (
    <Shell
      copy={copy}
      lang={lang}
      user={user}
      path={usersPage ? "/users" : "/"}
      items={items}
      onLang={changeLang}
      onNavigate={go}
      onSignOut={() => void signOut()}
    >
      {usersPage ? <Users copy={copy} onUnauthorized={() => setUser(null)} /> : <Home copy={copy} user={user} />}
    </Shell>
  );
}
