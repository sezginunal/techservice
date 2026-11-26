import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";

export function useWindow<T>(
  key: string,
  load: (after: T | null, signal: AbortSignal) => Promise<{ items: T[]; hasMore: boolean }>,
) {
  const loadRef = useRef(load);
  loadRef.current = load;
  const [items, setItems] = useState<T[]>([]);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const itemsRef = useRef(items);
  itemsRef.current = items;
  const hasMoreRef = useRef(false);
  hasMoreRef.current = hasMore;
  const busy = useRef(false);
  const moreAbort = useRef<AbortController | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    busy.current = true;
    setLoading(true);
    setError(null);
    loadRef
      .current(null, controller.signal)
      .then((page) => {
        if (controller.signal.aborted) {
          return;
        }
        setItems(page.items);
        setHasMore(page.hasMore);
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        setError(err);
        setItems([]);
        setHasMore(false);
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          busy.current = false;
          setLoading(false);
        }
      });
    return () => {
      controller.abort();
      moreAbort.current?.abort();
      busy.current = false;
    };
  }, [key]);

  const more = useCallback(() => {
    if (busy.current || !hasMoreRef.current) {
      return;
    }
    const after = itemsRef.current.at(-1);
    if (!after) {
      return;
    }
    const controller = new AbortController();
    moreAbort.current = controller;
    busy.current = true;
    setLoading(true);
    loadRef
      .current(after, controller.signal)
      .then((page) => {
        if (controller.signal.aborted) {
          return;
        }
        setItems((current) => [...current, ...page.items]);
        setHasMore(page.hasMore);
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        setError(err);
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          busy.current = false;
          setLoading(false);
        }
      });
  }, []);

  return { items, hasMore, loading, error, more };
}

export function End({ enabled, onEnd }: { enabled: boolean; onEnd: () => void }) {
  const ref = useRef<HTMLDivElement>(null);
  const onEndRef = useRef(onEnd);
  onEndRef.current = onEnd;

  useEffect(() => {
    const node = ref.current;
    if (!enabled || !node) {
      return;
    }
    const root = node.closest(".content");
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          onEndRef.current();
        }
      },
      { root: root instanceof Element ? root : null, rootMargin: "200px" },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [enabled]);

  return <div ref={ref} />;
}

export function usePaused(value: string) {
  const [paused, setPaused] = useState(value.trim());
  useEffect(() => {
    const timer = window.setTimeout(() => setPaused(value.trim()), 400);
    return () => window.clearTimeout(timer);
  }, [value]);
  return [paused, setPaused] as const;
}

export function useQuery() {
  const [value, setValue] = useState("");
  const [query, setQuery] = usePaused(value);
  return {
    value,
    setValue,
    query,
    commit(event?: FormEvent) {
      event?.preventDefault();
      setQuery(value.trim());
    },
  };
}
