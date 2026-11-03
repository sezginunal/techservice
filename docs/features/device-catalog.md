# Device catalog

Status: approved.

A signed-in technician, owner, or admin picks a brand and a model while registering a device, and while correcting one. The three roles see and do the same work. A signed-out request is unauthorized.

The brand and the model stay required. The operator picks a suggestion or types a value the catalog does not list. The device name stays a value the operator types.

The saved device stores the brand text and the model text left in the fields. A later edit of the catalog file does not change a device already saved.

## Settings

`.env` holds one setting. `.env.example` lists it.

- `DEVICE_CATALOG_PATH` is a JSON file in the `brand` directory next to the app. The value is the path the backend reads, for example `/brand/devices.json`. When the path is empty, or the file is missing, or the JSON is invalid, the screen keeps plain brand and model fields and the app still starts. The process log records the failed read and its cause.

The file is an array. Each element is one brand.

```json
[
  {
    "brand": "Acme",
    "models": ["Model 3", "S1"]
  }
]
```

`brand` is the brand name. `models` lists that brand's machines. The file lists each brand once. A sales kit uses that same machine name, so "Combo" and "Bundle" are not separate models. Empty brands and empty models are left out of the suggestions. The screen shows brands and models in file order.

Changing `.env` and starting the app again applies the new file. This feature has no screen that edits the catalog.

## Model image

A model image is a file under `brand/devices/`. The JSON does not name it. The directory is the brand and the file is the model. `Acme` / `S1` is `/brand/devices/acme/s1.webp`. `North Wind` / `Mini 2` is `/brand/devices/north-wind/mini-2.webp`. `Helios` / `XL+` is `/brand/devices/helios/xl-plus.webp`.

The name is lowercase. A `+` becomes `plus`. Any other run of characters that are not letters or digits becomes one hyphen. A hyphen at either end is removed. The file is WebP, PNG, or JPEG, in that order when more than one is present. A file larger than 2 MB, a file of another type, or a path outside `brand/devices/` is ignored. The process log records that cause. The screen shows no icon for that model, and the app still starts.

The image is not copied onto the device. Adding or removing the file changes what the screen shows the next time it loads. A brand and model that do not match a catalog model have no icon. The match ignores surrounding space and letter case.

## The screen

The registration screen and the correction screen use the same brand and model fields.

The brand field suggests the catalog brands as the operator types. The operator selects one or leaves a brand the catalog does not list.

When the brand matches a catalog brand, the model field suggests that brand's models. The match ignores surrounding space and letter case. When the brand matches none, the model field has no suggestions. The operator selects a model or leaves a model the catalog does not list.

Selecting a suggestion fills that field with the catalog text. The required-field checks stay as they are in intake.

A suggestion row shows the model image at icon size when that model has one. The device list and the device detail show the same icon when the stored brand and model match that catalog model. A model with no image shows the text alone.

## Not in this feature

- A screen that edits the catalog.
- Filling the device name from the catalog.
- Changing a saved device when the file changes.
