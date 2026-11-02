import { useEffect, useState, type FormEvent } from "react";
import { ApiError, request, type Role, type User } from "./api";
import type { Copy } from "./i18n";
import { messageFor } from "./Login";
import { initials } from "./Shell";

const roles: Role[] = ["technician", "owner", "admin"];

export function Users({ copy, onUnauthorized }: { copy: Copy; onUnauthorized: () => void }) {
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
      return true;
    } catch (err) {
      fail(err);
      return false;
    }
  }

  return (
    <>
      <header className="page-head">
        <h1>{copy.users}</h1>
        <p className="muted">{copy.usersLead}</p>
      </header>
      <form className="panel stack" onSubmit={(event) => void create(event)}>
        <h2>{copy.createUser}</h2>
        <div className="form-row">
          <label className="field">
            {copy.username}
            <input
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="off"
            />
          </label>
          <label className="field">
            {copy.password}
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="new-password"
            />
          </label>
          <label className="field">
            {copy.role}
            <RoleSelect copy={copy} value={role} onChange={setRole} />
          </label>
          <button className="button" type="submit">
            {copy.createUser}
          </button>
        </div>
        {error ? <p className="alert">{error}</p> : null}
      </form>
      <div className="panel table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>{copy.username}</th>
              <th>{copy.role}</th>
              <th>{copy.newPassword}</th>
              <th>{copy.active}</th>
              <th>{copy.save}</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => (
              <UserRow key={user.id} copy={copy} user={user} onSave={save} />
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}

function UserRow({
  copy,
  user,
  onSave,
}: {
  copy: Copy;
  user: User;
  onSave: (user: User, role: Role, password: string, active: boolean) => Promise<boolean>;
}) {
  const [role, setRole] = useState(user.role);
  const [password, setPassword] = useState("");
  const [active, setActive] = useState(user.active);
  const formId = `user-${user.id}`;

  return (
    <tr>
      <td>
        <form
          id={formId}
          onSubmit={(event) => {
            event.preventDefault();
            void onSave(user, role, password, active).then((saved) => {
              if (saved) {
                setPassword("");
              }
            });
          }}
        />
        <div className="who">
          <span className="avatar" aria-hidden="true">
            {initials(user.username)}
          </span>
          <div className="stack">
            <span className="who-name">{user.username}</span>
            <span className={user.active ? "badge" : "badge off"}>{user.active ? copy.active : copy.inactive}</span>
          </div>
        </div>
      </td>
      <td data-label={copy.role}>
        <RoleSelect copy={copy} value={role} form={formId} onChange={setRole} />
      </td>
      <td data-label={copy.newPassword}>
        <input
          form={formId}
          type="password"
          value={password}
          autoComplete="new-password"
          aria-label={copy.newPassword}
          onChange={(event) => setPassword(event.target.value)}
        />
      </td>
      <td>
        <label className="check">
          <input
            form={formId}
            type="checkbox"
            checked={active}
            onChange={(event) => setActive(event.target.checked)}
          />
          {copy.active}
        </label>
      </td>
      <td>
        <button className="button" type="submit" form={formId}>
          {copy.save}
        </button>
      </td>
    </tr>
  );
}

function RoleSelect({
  copy,
  value,
  form,
  onChange,
}: {
  copy: Copy;
  value: Role;
  form?: string;
  onChange: (role: Role) => void;
}) {
  return (
    <select form={form} value={value} onChange={(event) => onChange(event.target.value as Role)}>
      {roles.map((role) => (
        <option key={role} value={role}>
          {copy[role]}
        </option>
      ))}
    </select>
  );
}
