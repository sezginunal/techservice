# Brand

Status: approved.

The shop's own name and logo replace the built-in name and mark. A signed-out person sees them on the sign-in screen. A signed-in person sees them in the sidebar. The browser tab uses the same name, and the tab icon uses the logo when one is set.

The shown name is the company name followed by "Teknik Servis". It does not change with the language. The tagline stays as it is. When the company name is empty, the screen keeps "Techservice" in English and "Teknik servis" in Turkish.

## Settings

`.env` holds two settings. `.env.example` lists them.

- `COMPANY_NAME` is the name. When it is empty, the screen keeps "Techservice" in English and "Teknik servis" in Turkish.
- `LOGO_PATH` is a file in the `brand` directory next to the app. The value is the path the backend reads, for example `/brand/logo.png`. When the path is empty, or the file is missing, or the file is not a JPEG, PNG, WebP, or SVG, the screen keeps the current mark and the app still starts.

The sign-in screen can read the name and the logo without a session. A later request does not grant any other access.

## The screen

The shown name is the browser tab title. It is the strong name on the sign-in screen and in the sidebar.

When a logo is set, that image replaces the mark in those two places and becomes the browser tab icon. The image is the file from `LOGO_PATH`.

Changing `.env` and starting the app again applies the new name and logo. This feature has no screen that edits them.
