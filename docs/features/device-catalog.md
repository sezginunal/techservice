# Device catalog

Status: approved.

A signed-in technician, owner, or admin keeps the brands and models a service can suggest, and picks one while registering a service or correcting a device. The three roles see and do the same work. A signed-out request is unauthorized.

A model here is a brand and a model. It is not a physical device. A physical device is the device feature.

The brand and the model stay required on a service. The operator picks a suggestion or types a value the catalog does not list.

The saved device stores the brand text and the model text left in the fields. A later change to the catalog does not change a device already saved.

## Identity

- The brand and the model are required.
- A required value is refused when it is empty or only space. The screen names the field, with the same sentences as the device brand and model.
- A brand or a model longer than 200 characters is refused. The screen names the field.
- The stored brand and the stored model are the text left in the fields, with surrounding space removed.
- Those two are one model. The match ignores surrounding space and letter case.
- The system assigns a UUID. The operator does not type it. This feature has no page address for that UUID.
- The source is set when the row is stored. Catalog means the startup read it from the catalog file. Added means an operator stored it on the Models screen, on registration, or on a correction. A later start does not change the source. The operator does not edit it.
- This feature does not delete a model, and it does not rename a stored brand or model.

## Catalog file

`.env` holds one setting. `.env.example` lists it.

- `DEVICE_CATALOG_PATH` is a JSON file in the `brand` directory next to the app. The value is the path the backend reads, for example `/brand/devices.json`.

The file is an array. Each element is one brand.

```json
[
  {
    "brand": "Acme",
    "models": ["Model 3", "S1"]
  }
]
```

`brand` is the brand name. `models` lists that brand's machines. The file lists each brand once. A sales kit uses that same machine name, so "Combo" and "Bundle" are not separate models.

Every start reads the file and merges it into the table.

- A pair in the file that matches no row is inserted. The row is active. The source is Catalog. The stored text is the text in the file, with surrounding space removed.
- A pair in the file that matches a row leaves that row's text, source, and active flag as they are. The start does not insert a second row.
- A Catalog row whose pair is not in the file is set inactive. An Added row is left as it is. The start does not delete a row.
- A row that is already in the state the merge would write is left unchanged.

The operator can set a Catalog row active after the file stopped listing it. The next start sets that row inactive again while the file still does not list it.

Empty brands and empty models are left out. A brand or a model longer than 200 characters is left out. A repeated brand is left out. A repeated model under one brand is left out. The process log records each of those.

When the path is empty, or the file is missing, or the JSON is invalid, the table is left as it is and the app still starts. The process log records the failed read and its cause. A failed read does not set Catalog rows inactive. The catalog file and the model images stay on the machine. The repository does not include them.

The screen reads the table. It does not read the file for suggestions.

## Model image

A model image is a file under `brand/devices/`. The JSON does not name it. The directory is the brand and the file is the model, using the stored catalog text. `Acme` / `S1` is `/brand/devices/acme/s1.webp`. `North Wind` / `Mini 2` is `/brand/devices/north-wind/mini-2.webp`. `Helios` / `XL+` is `/brand/devices/helios/xl-plus.webp`.

The name is lowercase. A `+` becomes `plus`. Any other run of characters that are not letters or digits becomes one hyphen. A hyphen at either end is removed. The file is WebP, PNG, or JPEG, in that order when more than one is present. A file larger than 2 MB, a file of another type, or a path outside `brand/devices/` is ignored. The process log records that cause. The screen shows no icon for that model, and the app still starts.

The image is not copied onto the device. Adding or removing the file changes what the screen shows the next time it loads. The match ignores surrounding space and letter case. An active row and an inactive row both match. A brand and model that match no catalog row have no icon.

## Models

The signed-in navigation adds Models after Customers. The label is "Models" in English and "Modeller" in Turkish. The list address is `/models`. On a phone the entry sits under More. Home, Services, Customers, and Spare parts stay on the tab bar.

The list follows the lists rule. The order is the brand, then the model, then the id. The operator does not choose another order. The list shows active and inactive models. The operator does not filter by active.

The lead is "Models the shop can register, by brand." in English and "Kayıt edilebilecek modeller, markaya göre." in Turkish.

The caption says "An inactive model stays on services that already use it. It is not suggested for a later service." in English and "Kapalı bir model, o modelle kayıtlı servislerde durur. Bundan sonra önerilmez." in Turkish.

A row shows the model image at icon size when that model has one, the brand, the model, the source, and the active control. The source is "Catalog" in English and "Katalog" in Turkish, or "Added" in English and "Eklenen" in Turkish. The column label is "Source" in English and "Kaynak" in Turkish. Opening a row does not leave the list.

One search box matches the brand or the model. The match ignores surrounding space and letter case. The search waits until the operator pauses or submits.

An empty catalog says "No models yet." in English and "Henüz model yok." in Turkish. A search that matches nothing says "No model matches that search." in English and "Bu aramaya uyan model yok." in Turkish.

## Active

Active is "Active" in English and "Açık" in Turkish. Inactive is "Inactive" in English and "Kapalı" in Turkish. The control saves on its own. Setting the flag to the value it already has writes nothing.

An inactive model remains stored. A device already saved with that brand and model keeps that text. The service list, the service detail, the device page, and Home still show the icon when the file is there. The brand and model fields do not suggest it.

## Adding a model

The create form is on the list screen, above the search. The fields are the brand and the model. The control is "Add model" in English and "Model ekle" in Turkish.

The brand field suggests every stored brand. When the brand matches a stored brand, the model field suggests that brand's stored models, active and inactive. The match ignores surrounding space and letter case. Selecting a suggestion fills the field with the stored text. The operator can type a brand or a model that is not suggested.

A new brand and model store an active model. The source is Added. The save stays on the list. The form clears, and the list loads again so the model is on it.

When the brand and model match a model that is already stored, active or inactive, the save is refused and writes nothing. The screen says "A model with this brand and model is already registered." in English and "Bu marka ve modelle kayıtlı bir model var." in Turkish.

## Suggestions on a service

The registration screen and the correction screen use the same brand and model fields.

The brand field suggests the active brands as the operator types. A brand is suggested when it has at least one active model. The operator selects one or leaves a brand the catalog does not list.

When the brand matches a catalog brand, the model field suggests that brand's active models. The match ignores surrounding space and letter case. When the brand matches none, the model field has no suggestions. The operator selects a model or leaves a model the catalog does not list.

Selecting a suggestion fills that field with the stored text. A suggestion row shows the model image at icon size when that model has one. The required-field checks stay as they are in intake.

The service list, the service detail, the device page, and Home show the same icon when the stored brand and model match a catalog row, active or inactive. A model with no image shows the text alone.

## Saving a new pair

A successful registration stores the brand and model in the catalog when the table has no matching row. The new row is active. The source is Added. The stored text is the text left in the fields, with surrounding space removed. That save still writes the customer, the device, and the service as in intake and the device feature.

A successful correction stores the brand and model in the catalog when the corrected pair matches no catalog row. The new row is active. The source is Added. The device correction itself stays the device feature.

A match to a stored row, active or inactive, does not insert a second row and does not change that row. An inactive match stays inactive.

A refused save writes nothing, including no catalog row. The server is the check. When two saves would insert the same new pair, the first save writes the row and the second finds that row.

The startup merge writes no audit row, including when it sets a Catalog row inactive. An operator add, a pair stored from a service, and an active change write the audit rows named in the audit feature. A save that changes nothing writes nothing.

## Not in this feature

- Deleting a model.
- Renaming a stored brand or model.
- Changing a saved device when the catalog changes.
- Uploading or removing a model image from the screen.
- A page for one model.
- Spare-part brands and models. Those stay the spare parts feature.
