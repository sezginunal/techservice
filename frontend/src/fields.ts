const phonePattern = /^[0-9+()\- ]+$/;

function canonicalPhone(digits: string): string {
  if (digits.startsWith("0090") && digits.length >= 14) {
    return canonicalPhone(digits.slice(4));
  }
  if (digits.startsWith("90") && digits.length >= 12 && digits.length <= 13) {
    return canonicalPhone(digits.slice(2));
  }
  if (digits.startsWith("0") && digits.length === 11) {
    return digits.slice(1);
  }
  return digits;
}

export function phoneOk(value: string) {
  const phone = value.trim();
  if (!phone) {
    return false;
  }
  if ([...phone].length > 40 || !phonePattern.test(phone)) {
    return false;
  }
  const key = canonicalPhone(phone.replace(/\D/g, ""));
  return key.length >= 10 && key.length <= 15;
}

export function emailOk(value: string) {
  const email = value.trim();
  if (!email) {
    return true;
  }
  if ([...email].length > 254 || /\s/.test(email)) {
    return false;
  }
  const at = email.indexOf("@");
  if (at <= 0 || at !== email.lastIndexOf("@")) {
    return false;
  }
  const domain = email.slice(at + 1);
  return domain.includes(".") && !domain.startsWith(".") && !domain.endsWith(".") && !domain.includes("..");
}

export function usernameOk(value: string) {
  const name = value.trim();
  const chars = [...name].length;
  return chars > 0 && chars <= 64;
}

export function wholeQuantity(value: string, min: number, max: number) {
  const text = value.trim();
  if (!text) {
    return "empty" as const;
  }
  if (!/^\d+$/.test(text)) {
    return "range" as const;
  }
  const quantity = Number(text);
  if (!Number.isSafeInteger(quantity) || quantity < min || quantity > max) {
    return "range" as const;
  }
  return quantity;
}

export function passwordProblem(value: string) {
  const chars = [...value].length;
  if (chars < 8) {
    return "short" as const;
  }
  if (chars > 72 || new TextEncoder().encode(value).length > 72) {
    return "long" as const;
  }
  return null;
}
