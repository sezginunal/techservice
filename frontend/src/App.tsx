import { useEffect, useState, type FormEvent } from "react";
import { ApiError, request, type Role, type User } from "./api";
import { text, writeLang, type Copy, type Lang } from "./i18n";
import { LanguageSwitch } from "./LanguageSwitch";
import { Login, messageFor } from "./Login";

const roles: Role[] = ["technician", "owner", "admin"];

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
      <main className="page">
        <section className="panel">
          <p>{copy.checking}</p>
        </section>
      </main>
    );
  }

  if (!user) {
    return <Login copy={copy} lang={lang} onLang={changeLang} onSuccess={setUser} />;
  }

  return (
    <main className="page">
      <header className="row">
        <h1>{copy.appName}</h1>
        <div className="row">
          <LanguageSwitch copy={copy} lang={lang} onLang={changeLang} />
          <button className="button secondary" type="button" onClick={() => void signOut()}>
            {copy.signOut}
          </button>
        </div>
      </header>
      <nav className="row">
        <button className="button secondary" type="button" onClick={() => go("/")}>
          {copy.home}
        </button>
        {user.role === "admin" ? (
          <button className="button secondary" type="button" onClick={() => go("/users")}>
            {copy.users}
          </button>
        ) : null}
      </nav>
      {path === "/users" && user.role === "admin" ? (
        <Users copy={copy} onUnauthorized={() => setUser(null)} />
      ) : (
        <section className="panel stack">
          <p>
            {copy.signedIn} {user.username}
          </p>
          <p>
            {copy.role}: {copy[user.role]}
          </p>
        </section>
      )}
    </main>
  );
}

function Users({ copy, onUnauthorized }: { copy: Copy; onUnauthorized: () => void }) {
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<Role>("technician");

  function fail(err: unknown) {
    if (err instanceof ApiError && err.code === "unauthorized") {
      onUnauthorized();
      return;
    }
    setError(messageFor(copy, err));
  }

  useEffect(() => {
    request<User[]>("/api/users")
      .then(setUsers)
      .catch(fail);
  }, []);

  async function create(event: FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      const created = await request<User>("/api/users", {
        method: "POST",
        body: JSON.stringify({ username, password, role }),
      });
      setUsers((current) => [...current, created].sort((a, b) => a.username.localeCompare(b.username)));
      setUsername("");
      setPassword("");
      setRole("technician");
    } catch (err) {
      fail(err);
    }
  }

  async function save(user: User, nextRole: Role, nextPassword: string, nextActive: boolean) {
    setError(null);
    try {
      const saved = await request<User>(`/api/users/${user.id}`, {
        method: "PATCH",
        body: JSON.stringify({
          role: nextRole,
          active: nextActive,
          ...(nextPassword ? { password: nextPassword } : {}),
        }),
      });
      setUsers((current) => current.map((item) => (item.id === saved.id ? saved : item)));
    } catch (err) {
      fail(err);
    }
  }

  return (
    <section className="stack">
      <form className="panel stack" onSubmit={(event) => void create(event)}>
        <h2>{copy.createUser}</h2>
        <label className="field">
          {copy.username}
          <input value={username} onChange={(event) => setUsername(event.target.value)} />
        </label>
        <label className="field">
          {copy.password}
          <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
        </label>
        <label className="field">
          {copy.role}
          <RoleSelect copy={copy} value={role} onChange={setRole} />
        </label>
        {error ? <p className="alert">{error}</p> : null}
        <button className="button" type="submit">
          {copy.createUser}
        </button>
      </form>
      <div className="stack">
        {users.map((user) => (
          <UserRow key={user.id} copy={copy} user={user} onSave={save} />
        ))}
      </div>
    </section>
  );
}

function UserRow({
  copy,
  user,
  onSave,
}: {
  copy: Copy;
  user: User;
  onSave: (user: User, role: Role, password: string, active: boolean) => Promise<void>;
}) {
  const [role, setRole] = useState(user.role);
  const [password, setPassword] = useState("");
  const [active, setActive] = useState(user.active);

  return (
    <form
      className="panel stack"
      onSubmit={(event) => {
        event.preventDefault();
        void onSave(user, role, password, active).then(() => setPassword(""));
      }}
    >
      <h2>{user.username}</h2>
      <label className="field">
        {copy.role}
        <RoleSelect copy={copy} value={role} onChange={setRole} />
      </label>
      <label className="field">
        {copy.newPassword}
        <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
      </label>
      <label className="check">
        <input type="checkbox" checked={active} onChange={(event) => setActive(event.target.checked)} />
        {copy.active}
      </label>
      <button className="button" type="submit">
        {copy.save}
      </button>
    </form>
  );
}

function RoleSelect({ copy, value, onChange }: { copy: Copy; value: Role; onChange: (role: Role) => void }) {
  return (
    <select value={value} onChange={(event) => onChange(event.target.value as Role)}>
      {roles.map((role) => (
        <option key={role} value={role}>
          {copy[role]}
        </option>
      ))}
    </select>
  );
}
