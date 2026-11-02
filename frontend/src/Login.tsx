import { useState, type FormEvent } from "react";
import { ApiError, request, type User } from "./api";
import type { Copy } from "./i18n";
import { LanguageSwitch } from "./LanguageSwitch";

type LoginProps = {
  copy: Copy;
  lang: "en" | "tr";
  onLang: (lang: "en" | "tr") => void;
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
    <main className="page">
      <form className="panel stack" onSubmit={submit}>
        <div className="row">
          <h1>{copy.appName}</h1>
          <LanguageSwitch copy={copy} lang={lang} onLang={onLang} />
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
        <button className="button" type="submit">
          {copy.signIn}
        </button>
      </form>
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
