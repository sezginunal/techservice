# Device status

Status: approved.

A signed-in technician, owner, or admin lists every device, opens one, and changes its state. The three roles see and do the same work. A signed-out request is unauthorized.

## Device notes

The free-text field on the device is the device notes. Every screen shows it as "Device notes (optional)" in English and "Cihaz notları (isteğe bağlı)" in Turkish. It is optional. The operator writes any text, or leaves it empty.

The part note shows "Part note (optional)" in English and "Parça notu (isteğe bağlı)" in Turkish. It is optional.

"Status" in English and "Durum" in Turkish mean only the device state.

## States

The state is one of these values. The screen shows the English name, then the Turkish name.

- `registered` — Registered — Kayıt yapıldı
- `inspecting` — Inspecting — Ön inceleme yapılıyor
- `repair_warranty` — Repair under warranty — Garanti kapsamında tamiri yapılıyor
- `repair_out_of_warranty` — Repair outside warranty — Garanti dışında tamiri yapılıyor
- `waiting_partner_part` — Waiting for a partner part — İş ortağından yedek parça bekleniyor
- `waiting_own_part` — Waiting for a part we will buy — Kendi satın alacağımız yedek parça bekleniyor
- `final_testing` — Final tests — Final testleri yapılıyor
- `preparing_delivery` — Preparing for delivery — Teslim için hazırlanıyor
- `in_storage` — In storage — Depoda
- `left_storage` — Left storage — Depodan çıkış yapıldı
- `delivered` — Delivered — Teslim edildi

Registration sets `registered`. The registration screen has no state control.

The operator moves a device from any state to any other state. A save that keeps the same state is refused and writes nothing.

Warranty lives only in `repair_warranty` and `repair_out_of_warranty`. A device in another state is not filtered as under warranty. The previous state stays in the history. This feature adds no separate warranty field.

## Device list

The device list follows the lists rule.

A row shows the UUID, brand, model, customer name, state, and registration date. It does not show photos, parts, or notes.

The search box stays. It matches the brand, model, serial number, customer name, phone, or UUID.

A state filter offers every state, or all states. An empty choice means all states. The device list address keeps the chosen state. The search and the state filter apply together.

The newest registration is first.

An empty shop says there are no devices. A search or a state filter that matches nothing says so.

Opening a row opens the device.

The customer detail lists that customer's devices with the state names above. The operator changes a state only from the device detail.

## Changing a state

The device detail shows the current state. The operator picks a new state, writes a change note or leaves it empty, adds one photo or adds none, and saves.

The change note is optional. A photo is optional. The photo is JPEG, PNG, or WebP, and at most 8 MB. One change holds one photo. That photo does not count toward the device's 30 photos.

A refused photo or a same-state save writes nothing. The stored state stays as it was.

The correction screen still edits the device notes, the device fields, the parts, and the device photos. It does not change the state.

## History

The device detail shows the state history, newest first.

A row shows the time, the actor's username, the previous state, the new state, the change note when one was entered, and the photo when one was added.

The history is append-only. This feature does not edit or delete a row.

A successful state change writes the audit event already named in the audit feature. That audit row shows the previous state and the new state. It does not contain the change note or the photo file.

## Not in this feature

- A separate warranty field.
- A required order of states.
- A note or a photo without a state change.
- Editing or deleting a history row.
- Changing the state from the list or from the customer detail.
- Choosing a state on the registration screen.
- More than one photo on one change.
