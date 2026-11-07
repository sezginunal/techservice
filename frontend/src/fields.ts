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
  return name.length > 0 && name.length <= 64;
}

export function passwordProblem(value: string) {
  if (value.length < 8) {
    return "short" as const;
  }
  if (value.length > 72) {
    return "long" as const;
  }
  return null;
}
