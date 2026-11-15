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

A successful change to a record writes one row. A save that changes nothing writes nothing. Parts and photos are part of the service. They do not write their own rows. A save that changes the device and the service writes one row for each.

The row does not contain a password, device notes, or a photo file.

These events write a row:

- A customer is created. The subject is the customer name.
- A customer is corrected. The subject is the customer name. A rename shows the previous name and the new name.
- A service is registered. The subject is the brand, the model, the serial number, the service UUID, and the customer name. The screen calls this event "Service registered" in English and "Servis açıldı" in Turkish. The first service on a new device does not write a separate device row.
- A service is corrected. The subject is the brand, the model, the serial number, the service UUID, and the customer name. A change of customer shows the previous customer name and the new customer name. The screen calls this event "Service updated" in English and "Servis düzeltildi" in Turkish.
- A device is corrected. The subject is the brand, the model, the serial number, and the device UUID. The row shows the previous brand, model, and serial number, and the new brand, model, and serial number. The screen calls this event "Device updated" in English and "Cihaz düzeltildi" in Turkish.
- A service status changes. The row shows the previous status and the new status. The screen calls this event "Status changed" in English and "Durum değiştirildi" in Turkish. Intake does not change status. The device status feature writes this row.
- A user is created. The subject is the new username and the new role.
- A user's role changes. The row shows the previous role and the new role.
- A user's password changes. The subject is that username.
- A user is activated, or a user is deactivated. The subject is that username.

Registering a service with a new customer writes two rows: the customer is created, then the service is registered. The names on a row stay as they were when the row was written.

These do not write a row:

- Sign-in, sign-out, and a failed sign-in.
- A session expiring.
- Opening a list or a record.
- A refused validation.

## The screen

- The screen is one list. A row shows the time, the actor's username, the actor's role at that time, the event, and the subject. There is no second page for a row.
- The list follows the lists rule.
- The list opens on the newest row. The operator can sort by time, newest first or oldest first.
- The filters apply together. Either date may be empty. The date is the calendar day of the time shown on the screen.
- The person filter matches the actor's username.
- The event filter is the event list above. An empty choice means every event.
- The subject filter matches the customer name, the brand, the model, the serial number, the service UUID, the device UUID, or, for a user change, that username.
- The person filter and the subject filter wait until the operator pauses or submits.
- An empty log says there are no rows. A filter that matches nothing says so.
