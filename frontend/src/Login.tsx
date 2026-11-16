import { useState, type FormEvent } from "react";
import { ApiError, request, type Brand, type User } from "./api";
import type { Copy, Lang } from "./i18n";
import { LanguageSwitch } from "./LanguageSwitch";
import { AppVersion, BrandMark } from "./Shell";

type LoginProps = {
  copy: Copy;
  lang: Lang;
  brand: Brand;
  onLang: (lang: Lang) => void;
  onSuccess: (user: User) => void;
};

export function Login({ copy, lang, brand, onLang, onSuccess }: LoginProps) {
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
      <section className="auth-art" aria-hidden="true">
        <div className="brand">
          <BrandMark copy={copy} brand={brand} />
        </div>
        <div className="art">
          <div className="art-card">
            <div className="art-row">
              <span className="dot" data-tone="teal" />
              <span className="art-line" />
              <span className="art-pill" data-tone="teal" />
            </div>
            <div className="art-row">
              <span className="dot" data-tone="amber" />
              <span className="art-line" />
              <span className="art-pill" data-tone="amber" />
            </div>
            <div className="art-row">
              <span className="dot" data-tone="green" />
              <span className="art-line" />
              <span className="art-pill" data-tone="green" />
            </div>
          </div>
          <div className="art-card art-chart">
            <span style={{ height: "35%" }} />
            <span style={{ height: "55%" }} />
            <span style={{ height: "42%" }} />
            <span style={{ height: "70%" }} />
            <span style={{ height: "88%" }} />
            <span style={{ height: "62%" }} />
            <span style={{ height: "95%" }} />
          </div>
        </div>
        <p>{copy.tagline}</p>
      </section>
      <div className="auth-main">
        <div className="auth-tools">
          <LanguageSwitch copy={copy} lang={lang} onLang={onLang} />
        </div>
        <div className="auth-card">
          <div className="brand">
            <BrandMark copy={copy} brand={brand} />
          </div>
          <div className="auth-head">
            <h1>{copy.signIn}</h1>
            <p className="muted">{copy.signInLead}</p>
          </div>
          <form className="auth-form" onSubmit={submit}>
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
        <div className="auth-foot">
          <AppVersion copy={copy} />
        </div>
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
