import { useEffect, useState } from "react";
import type { Copy } from "./i18n";

type Props = {
  src: string;
  name: string;
  copy: Copy;
  className?: string;
  onUnauthorized?: () => void;
};

export function PhotoButton({ src, name, copy, className, onUnauthorized }: Props) {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button className="photo-open" type="button" aria-label={copy.open} onClick={() => setOpen(true)}>
        <img className={className} src={src} alt="" />
      </button>
      {open ? (
        <PhotoView src={src} name={name} copy={copy} onClose={() => setOpen(false)} onUnauthorized={onUnauthorized} />
      ) : null}
    </>
  );
}

function PhotoView({
  src,
  name,
  copy,
  onClose,
  onUnauthorized,
}: {
  src: string;
  name: string;
  copy: Copy;
  onClose: () => void;
  onUnauthorized?: () => void;
}) {
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = previous;
      window.removeEventListener("keydown", onKey);
    };
  }, [onClose]);

  async function download() {
    setSaving(true);
    try {
      const response = await fetch(src);
      if (response.status === 401) {
        onUnauthorized?.();
        return;
      }
      if (!response.ok) {
        return;
      }
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = fileName(name, blob.type || response.headers.get("Content-Type") || "");
      link.click();
      window.setTimeout(() => URL.revokeObjectURL(url), 0);
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="photo-view" role="dialog" aria-modal="true" aria-label={copy.open} onClick={onClose}>
      <div className="photo-view-frame" onClick={(event) => event.stopPropagation()}>
        <img src={src} alt="" />
        <div className="photo-view-actions">
          <button className="button secondary" type="button" disabled={saving} onClick={() => void download()}>
            <DownloadIcon />
            {copy.download}
          </button>
          <button className="button secondary" type="button" onClick={onClose}>
            {copy.close}
          </button>
        </div>
      </div>
    </div>
  );
}

function fileName(name: string, media: string) {
  if (/\.(jpe?g|png|webp)$/i.test(name)) {
    return name;
  }
  const ext = extension(media.split(";")[0].trim());
  return ext ? `${name}.${ext}` : name;
}

function extension(type: string) {
  if (type === "image/jpeg") {
    return "jpg";
  }
  if (type === "image/png") {
    return "png";
  }
  if (type === "image/webp") {
    return "webp";
  }
  return "";
}

function DownloadIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d="M12 4v10" />
      <path d="M7 11l5 5 5-5" />
      <path d="M5 20h14" />
    </svg>
  );
}
