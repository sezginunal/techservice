export type Lang = "en" | "tr";

const en = {
  appName: "Techservice",
  tagline: "Staff workspace",
  signInLead: "Enter your username and password.",
  general: "General",
  menu: "Menu",
  closeMenu: "Close menu",
  usersLead: "Create a user, change a role, or set a password.",
  inactive: "Inactive",
  username: "Username",
  password: "Password",
  signIn: "Sign in",
  signOut: "Sign out",
  invalidCredentials: "The username or password is wrong.",
  home: "Home",
  users: "Users",
  role: "Role",
  technician: "Technician",
  owner: "Owner",
  admin: "Admin",
  signedIn: "Signed in as",
  createUser: "Create user",
  save: "Save",
  active: "Active",
  newPassword: "New password",
  passwordTooShort: "Use at least 8 characters.",
  passwordTooLong: "Use at most 72 characters.",
  usernameTaken: "That username is already in use.",
  lastAdmin: "The last admin must stay an active admin.",
  invalidRole: "Choose a valid role.",
  invalidUsername: "Enter a username.",
  forbidden: "You cannot do that.",
  unexpected: "Something went wrong.",
  language: "Language",
  english: "English",
  turkish: "Turkish",
  checking: "Checking the session",
};

const tr: typeof en = {
  appName: "Teknik servis",
  tagline: "Personel alanı",
  signInLead: "Kullanıcı adı ve şifreni gir.",
  general: "Genel",
  menu: "Menü",
  closeMenu: "Menüyü kapat",
  usersLead: "Kullanıcı oluştur, rolünü değiştir veya şifre belirle.",
  inactive: "Pasif",
  username: "Kullanıcı adı",
  password: "Şifre",
  signIn: "Giriş yap",
  signOut: "Çıkış yap",
  invalidCredentials: "Kullanıcı adı veya şifre yanlış.",
  home: "Ana sayfa",
  users: "Kullanıcılar",
  role: "Rol",
  technician: "Teknisyen",
  owner: "Firma sahibi",
  admin: "Yönetici",
  signedIn: "Giriş yapan",
  createUser: "Kullanıcı oluştur",
  save: "Kaydet",
  active: "Aktif",
  newPassword: "Yeni şifre",
  passwordTooShort: "En az 8 karakter kullan.",
  passwordTooLong: "En fazla 72 karakter kullan.",
  usernameTaken: "Bu kullanıcı adı kullanılıyor.",
  lastAdmin: "Son yönetici aktif yönetici olarak kalmalı.",
  invalidRole: "Geçerli bir rol seç.",
  invalidUsername: "Bir kullanıcı adı gir.",
  forbidden: "Bunu yapamazsın.",
  unexpected: "Bir şeyler ters gitti.",
  language: "Dil",
  english: "İngilizce",
  turkish: "Türkçe",
  checking: "Oturum kontrol ediliyor",
};

const dictionaries = { en, tr };

const storageKey = "techservice.lang";

export function readLang(): Lang {
  const value = localStorage.getItem(storageKey);
  return value === "tr" ? "tr" : "en";
}

export function writeLang(lang: Lang) {
  localStorage.setItem(storageKey, lang);
}

export function text(lang: Lang) {
  return dictionaries[lang];
}

export type Copy = typeof en;
