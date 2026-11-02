import { useEffect, useState } from "react";

type Health = {
  status: string;
  database: string;
};

export function App() {
  const [health, setHealth] = useState<Health | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    fetch("/api/health", { signal: controller.signal })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(String(response.status));
        }
        return response.json() as Promise<Health>;
      })
      .then((body) => {
        setHealth(body);
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") {
          return;
        }
        setError("Backend is not reachable");
      });

    return () => controller.abort();
  }, []);

  let message = "Checking the connection";
  if (health) {
    message = `Backend ${health.status}, database ${health.database}`;
  } else if (error) {
    message = error;
  }

  return (
    <main>
      <h1>Techservice</h1>
      <p>{message}</p>
    </main>
  );
}
