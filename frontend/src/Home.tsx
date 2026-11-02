import type { User } from "./api";
import type { Copy } from "./i18n";
import { initials } from "./Shell";

export function Home({ copy, user }: { copy: Copy; user: User }) {
  return (
    <>
      <header className="page-head">
        <h1>{copy.home}</h1>
      </header>
      <section className="panel identity">
        <span className="avatar" aria-hidden="true">
          {initials(user.username)}
        </span>
        <div className="stack">
          <p className="muted">{copy.signedIn}</p>
          <h2>{user.username}</h2>
          <span className="badge role">{copy[user.role]}</span>
        </div>
      </section>
    </>
  );
}
