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
        <div className="auth-art-copy">
          <p className="auth-art-kicker">{copy.authArtKicker}</p>
          <p className="auth-art-title">{copy.authArtTitle}</p>
          <ul className="auth-art-points">
            <li>{copy.authArtPoint1}</li>
            <li>{copy.authArtPoint2}</li>
            <li>{copy.authArtPoint3}</li>
          </ul>
        </div>
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
