# Techservice

Bench software for a repair shop. Staff sign in, record a customer, open a service for a device, attach parts and photos, and move the job through to delivery.

The interface is English and Turkish. It runs on the shop network: one machine, two containers. The browser talks only to the frontend. The frontend proxies `/api` to the backend.

## Run

```bash
cp .env.example .env
```

Set `ADMIN_USERNAME` and `ADMIN_PASSWORD` in `.env`. The password is 8 to 72 characters. That account is created on the first start, when the database has no admin. A later start leaves the stored password as it is.

```bash
docker compose up --build
```

Open `http://<this-machine>:8080`. The host port is `FRONTEND_PORT`.

SQLite and uploaded photos live on named volumes. Recreating the containers keeps both. Only the frontend port is published.

## Configuration

Settings come from `.env`. Omit a line to keep the default in [`.env.example`](.env.example).

| Variable | |
| --- | --- |
| `FRONTEND_PORT` | Host port. Default `8080`. |
| `ADMIN_USERNAME` | First admin. Used only when no admin exists. |
| `ADMIN_PASSWORD` | 8–72 characters. Stay out of the repository. |
| `COMPANY_NAME` | Shop name on the sign-in screen, the sidebar, and the browser tab. Empty keeps Techservice / Teknik servis. |
| `LOGO_PATH` | Logo file the backend reads, for example `/brand/logo.webp`. JPEG, PNG, WebP, or SVG. |
| `DEVICE_CATALOG_PATH` | Brand and model suggestions, for example `/brand/devices.json`. |

`compose.yaml` mounts `./brand` at `/brand`. The logo, the catalog, and the model images stay on the machine. A missing logo or an unreadable catalog leaves the built-in mark and plain fields. The app still starts.

A catalog file is a list of brands:

```json
[
  {
    "brand": "Acme",
    "models": ["Model 3", "S1"]
  }
]
```

A model image sits under `brand/devices/`, named from the brand and the model: `brand/devices/acme/s1.webp`. The naming rules are in [Device catalog](docs/features/device-catalog.md).

## Roles

| Stored role | On screen | |
| --- | --- | --- |
| `admin` | Admin | Users, and the full audit log. |
| `owner` | Manager / Yönetici | The audit log, except rows written by an admin. |
| `technician` | Operator / Operatör | The shop screens. |

A session lasts 7 days from sign-in, then ends.

## Layout

```
backend/          Go API, SQLite, uploaded files
frontend/         React app, served by nginx
brand/            Logo, catalog, and model images. Local.
docs/features/    What each screen does
```

The app version is [`VERSION`](VERSION). The sign-in screen shows it.

## Development

Backend tests, from `backend/`:

```bash
go test ./...
```

Frontend build, from `frontend/`. The build reads `VERSION` in the repository root.

```bash
npm ci
npm run build
```

## Screens

- [Home](docs/features/home.md) — the shop summary
- [Intake](docs/features/intake.md) — customers, services, parts, and photos
- [Device](docs/features/device.md) — one machine and its visits
- [Status](docs/features/device-status.md) — states and history
- [Photos](docs/features/photos.md) — open and download a photo
- [Audit](docs/features/audit.md) — who changed a record
- [Brand](docs/features/brand.md) — shop name and logo
- [Device catalog](docs/features/device-catalog.md) — brand and model suggestions

## License

[MIT](LICENSE)
