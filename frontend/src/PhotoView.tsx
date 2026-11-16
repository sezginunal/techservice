import { useEffect, useState } from "react";
import type { Copy } from "./i18n";
import { Icon } from "./icons";

type Props = {
  src: string;
  name: string;
  copy: Copy;
  className?: string;
  onUnauthorized?: () => void;
};

export function PhotoButton({ src, name, copy, onUnauthorized }: Props) {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button className="photo-open" type="button" aria-label={copy.open} onClick={() => setOpen(true)}>
        <img src={src} alt="" />
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
    <div className="lightbox" role="dialog" aria-modal="true" aria-label={copy.open} onClick={onClose}>
      <div className="lightbox-bar">
        <button
          className="button glass"
          type="button"
          disabled={saving}
          onClick={(event) => {
            event.stopPropagation();
            void download();
          }}
        >
          <Icon name="download" size={16} />
          {copy.download}
        </button>
        <button
          className="icon-button glass"
          type="button"
          aria-label={copy.close}
          onClick={(event) => {
            event.stopPropagation();
            onClose();
          }}
        >
          <Icon name="close" size={18} />
        </button>
      </div>
      <div className="lightbox-stage" onClick={(event) => event.stopPropagation()}>
        <img src={src} alt="" />
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
