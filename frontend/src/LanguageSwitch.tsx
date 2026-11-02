import type { Copy, Lang } from "./i18n";

type LanguageSwitchProps = {
  copy: Copy;
  lang: Lang;
  onLang: (lang: Lang) => void;
};

export function LanguageSwitch({ copy, lang, onLang }: LanguageSwitchProps) {
  return (
    <div className="segment" role="group" aria-label={copy.language}>
      <button type="button" aria-pressed={lang === "en"} onClick={() => onLang("en")}>
        EN
      </button>
      <button type="button" aria-pressed={lang === "tr"} onClick={() => onLang("tr")}>
        TR
      </button>
    </div>
  );
}
