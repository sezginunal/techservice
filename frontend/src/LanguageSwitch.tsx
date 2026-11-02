import type { Copy, Lang } from "./i18n";

type LanguageSwitchProps = {
  copy: Copy;
  lang: Lang;
  onLang: (lang: Lang) => void;
};

export function LanguageSwitch({ copy, lang, onLang }: LanguageSwitchProps) {
  return (
    <label className="field language">
      {copy.language}
      <select value={lang} onChange={(event) => onLang(event.target.value as Lang)}>
        <option value="en">{copy.english}</option>
        <option value="tr">{copy.turkish}</option>
      </select>
    </label>
  );
}
