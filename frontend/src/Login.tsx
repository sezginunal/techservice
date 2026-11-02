import { useState, type FormEvent } from "react";
import { ApiError, request, type User } from "./api";
import type { Copy, Lang } from "./i18n";
import { LanguageSwitch } from "./LanguageSwitch";

type LoginProps = {
  copy: Copy;
  lang: Lang;
  onLang: (lang: Lang) => void;
  onSuccess: (user: User) => void;
};

export function Login({ copy, lang, onLang, onSuccess }: LoginProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      const user = await request<User>("/api/login", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });
      onSuccess(user);
    } catch (err) {
      setError(messageFor(copy, err));
    }
  }

  return (
    <main className="auth">
      <div className="auth-wrap">
        <div className="auth-tools">
          <LanguageSwitch copy={copy} lang={lang} onLang={onLang} />
        </div>
        <div className="auth-brand">
          <span className="mark" aria-hidden="true">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M14.7 6.3a4 4 0 0 0-5.4 5.4L4 17l3 3 5.3-5.3a4 4 0 0 0 5.4-5.4L15 12l-3-3 2.7-2.7z" />
            </svg>
          </span>
          <span className="brand-name">
            <strong>{copy.appName}</strong>
            <span>{copy.tagline}</span>
          </span>
        </div>
        <form className="panel stack" onSubmit={submit}>
          <div className="page-head">
            <h1>{copy.signIn}</h1>
            <p className="muted">{copy.signInLead}</p>
          </div>
          <label className="field">
            {copy.username}
            <input value={username} onChange={(event) => setUsername(event.target.value)} autoComplete="username" />
          </label>
          <label className="field">
            {copy.password}
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
            />
          </label>
          {error ? <p className="alert">{error}</p> : null}
          <button className="button wide" type="submit">
            {copy.signIn}
          </button>
        </form>
      </div>
    </main>
  );
}

export function messageFor(copy: Copy, err: unknown) {
  const code = err instanceof ApiError ? err.code : "unexpected";
  const messages: Record<string, string> = {
    invalid_credentials: copy.invalidCredentials,
    password_too_short: copy.passwordTooShort,
    password_too_long: copy.passwordTooLong,
    username_taken: copy.usernameTaken,
    last_admin: copy.lastAdmin,
    invalid_role: copy.invalidRole,
    invalid_username: copy.invalidUsername,
    forbidden: copy.forbidden,
  };
  return messages[code] ?? copy.unexpected;
}
