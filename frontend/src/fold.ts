/** Same folding as backend store.fold: trim, map Turkish letters, lower case. */
export function fold(value: string) {
  let text = value.trim();
  text = text
    .replaceAll("İ", "i")
    .replaceAll("I", "i")
    .replaceAll("ı", "i")
    .replaceAll("Ş", "s")
    .replaceAll("ş", "s")
    .replaceAll("Ğ", "g")
    .replaceAll("ğ", "g")
    .replaceAll("Ü", "u")
    .replaceAll("ü", "u")
    .replaceAll("Ö", "o")
    .replaceAll("ö", "o")
    .replaceAll("Ç", "c")
    .replaceAll("ç", "c");
  return text.toLowerCase();
}
