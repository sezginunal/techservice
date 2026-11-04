const phonePattern = /^[0-9+()\- ]+$/;

export function phoneOk(value: string) {
  const phone = value.trim();
  if (!phone) {
    return true;
  }
  if ([...phone].length > 40 || !phonePattern.test(phone)) {
    return false;
  }
  const digits = phone.replace(/\D/g, "").length;
  return digits >= 10 && digits <= 15;
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
