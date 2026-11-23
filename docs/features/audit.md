# Audit

Status: approved.

An admin or an owner inspects a log of important record changes. A technician has no audit screen. A signed-out request is unauthorized. A technician's request is forbidden.

The log is append-only. This feature does not edit or delete a row.

## Role names

The stored roles stay `admin`, `owner`, and `technician`. The visible names change on every screen, including home and user management.

- `admin` is shown as "Admin" in English and "Admin" in Turkish.
- `owner` is shown as "Manager" in English and "Yönetici" in Turkish.
- `technician` is shown as "Operator" in English and "Operatör" in Turkish.

## Who sees a row

- A row stores the actor's username and the actor's role at that moment. A later role change does not rewrite the row.
- An admin sees rows for `admin`, `owner`, and `technician`.
- An owner sees rows for `owner` and `technician`. Rows for `admin` are not in the result.
- A technician sees no rows and has no navigation entry. Opening the address does not show the log.
- The navigation shows Audit to an admin and an owner.

## What writes a row

A successful change to a record writes one row. A save that changes nothing writes nothing. The parts and photos on a service are part of that service. They do not write their own rows. A spare part writes the rows named in the spare parts feature. A save that changes the device and the service writes one row for each.

The row does not contain a password, device notes, a spare-part note, or a photo file.

These events write a row:

- A customer is created. The subject is the customer name.
- A customer is corrected. The subject is the customer name. A rename shows the previous name and the new name.
- A service is registered. The subject is the brand, the model, the serial number, the service UUID, and the customer name. The screen calls this event "Service registered" in English and "Servis açıldı" in Turkish. The first service on a new device does not write a separate device row.
- A service is corrected. The subject is the brand, the model, the serial number, the service UUID, and the customer name. A change of customer shows the previous customer name and the new customer name. The screen calls this event "Service updated" in English and "Servis düzeltildi" in Turkish.
- A device is corrected. The subject is the brand, the model, the serial number, and the device UUID. The row shows the previous brand, model, and serial number, and the new brand, model, and serial number. The screen calls this event "Device updated" in English and "Cihaz düzeltildi" in Turkish.
- A model is added. The subject is the brand and the model. The screen calls this event "Model added" in English and "Model eklendi" in Turkish. The device catalog writes this row when the operator adds a model, and when a service registration or a device correction stores a new pair. The catalog file read at startup does not write this row.
- A model is set active or inactive. The subject is the brand and the model. The row shows "Active" or "Inactive" in English and "Açık" or "Kapalı" in Turkish. The screen calls this event "Model updated" in English and "Model güncellendi" in Turkish.
- A service status changes. The row shows the previous status and the new status. The screen calls this event "Status changed" in English and "Durum değiştirildi" in Turkish. Intake does not change status. The device status feature writes this row.
- A user is created. The subject is the new username and the new role.
- A user's role changes. The row shows the previous role and the new role.
- A user's password changes. The subject is that username.
- A user is activated, or a user is deactivated. The subject is that username.
- A service state is added. The subject is the new name in the active language. The row stores the English name and the Turkish name. When the state is shown on Home, the subject also says "Shown on Home" in English and "Ana sayfada" in Turkish. The screen calls this event "State added" in English and "Durum eklendi" in Turkish. The settings feature writes this row.
- A service state is updated. The subject is the name in the active language. A rename shows the previous name and the new name in that language. The row stores the previous and new English names and the previous and new Turkish names. A change between active and inactive shows "Active" or "Inactive" in English and "Açık" or "Kapalı" in Turkish. A change of the Home queue shows "Shown on Home" or "Hidden on Home" in English and "Ana sayfada" or "Ana sayfada gizli" in Turkish. The screen calls this event "State updated" in English and "Durum güncellendi" in Turkish.
- The service-state order changes. The subject is the name of the moved state in the active language. The screen calls this event "States reordered" in English and "Durum sırası değişti" in Turkish.
- A spare part is created. The subject is the brand, the model, and the quantity, including 0. The screen calls this event "Spare part added" in English and "Yedek parça eklendi" in Turkish. The spare parts feature writes this row.
- A spare part is corrected. The subject is the brand and the model. The row shows the previous brand and model and the new brand and model. The screen calls this event "Spare part updated" in English and "Yedek parça düzeltildi" in Turkish. The spare parts feature writes this row.
- Stock is added to a spare part, including the quantity added when a create matches a part that already exists. The subject is the brand, the model, and the quantity. The screen calls this event "Stock added" in English and "Stok eklendi" in Turkish. The spare parts feature writes this row.
- Spare-part stock is written off. The subject is the brand, the model, and the quantity. The screen calls this event "Stock written off" in English and "Fire kaydedildi" in Turkish. The spare parts feature writes this row.
- A spare part is used on a service. The subject is the spare part's brand and model, the quantity, the service UUID, the device's brand, model, and serial number, and the state. The screen calls this event "Spare part used" in English and "Yedek parça kullanıldı" in Turkish. The spare parts feature writes this row.
- A spare part is returned from a service. The subject is the spare part's brand and model, the quantity, the service UUID, the device's brand, model, and serial number, and the state. The screen calls this event "Spare part returned" in English and "Yedek parça iade edildi" in Turkish. The spare parts feature writes this row.

Registering a service with a new customer writes two rows: the customer is created, then the service is registered. The names on a row stay as they were when the row was written.

These do not write a row:

- Sign-in, sign-out, and a failed sign-in.
- A session expiring.
- Opening a list or a record.
- A refused validation.
- Reading the catalog file at startup, including a Catalog model set inactive because the file no longer lists it.

## The screen

- The screen is one list. A row shows the time, the actor's username, the actor's role at that time, the event, and the subject. There is no second page for a row.
- The list follows the lists rule.
- The list opens on the newest row. The operator can sort by time, newest first or oldest first.
- The filters apply together. Either date may be empty. The date is the calendar day of the time shown on the screen.
- The person filter matches the actor's username.
- The event filter is the event list above. An empty choice means every event.
- The subject filter matches the customer name, the brand, the model, the serial number, the service UUID, the device UUID, or, for a user change, that username, or, for a change to the state list, the English name or the Turkish name.
- The person filter and the subject filter wait until the operator pauses or submits.
- An empty log says there are no rows. A filter that matches nothing says so.
