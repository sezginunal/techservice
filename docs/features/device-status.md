# Device status

Status: approved.

A signed-in technician, owner, or admin lists every service, opens one, and changes its state. The three roles see and do the same work. A signed-out request is unauthorized.

## Device notes

The free-text field on the service is the device notes. Every screen shows it as "Device notes (optional)" in English and "Cihaz notları (isteğe bağlı)" in Turkish. It is optional. The operator writes any text, or leaves it empty.

The part note shows "Part note (optional)" in English and "Parça notu (isteğe bağlı)" in Turkish. It is optional.

"Status" in English and "Durum" in Turkish mean only the service state.

## States

The installation's state list is the settings feature. One installation has one list.

A service stores the state's code. The screen shows the catalog name for the active language, as in the settings feature. The code is never shown.

Registration sets `registered`. The registration screen has no state control.

`registered`, `delivered`, and `cancelled` are always active. `repair_warranty` and `repair_out_of_warranty` start active and can be made inactive. Any state the shop adds can be made inactive. The settings feature is the place that does this.

The operator moves a service to an active state. A save that keeps the same state is refused and writes nothing. A save that chooses an inactive state, or a state this installation does not have, is refused and writes nothing. The stored state stays as it was.

A service already in an inactive state keeps that state. The detail shows it. The operator can move it to an active state.

Warranty lives only in `repair_warranty` and `repair_out_of_warranty`. A service in another state, including a state the shop added, is not filtered as under warranty. The previous state stays in the history. This feature adds no separate warranty field. A state the shop adds has no rule of its own.

A cancelled service is not delivered. The in-shop count includes it. Home aging leaves out `delivered` and `cancelled`. The device feature still refuses a second service while one service is neither `delivered` nor `cancelled`.

## Service list

The service list is the list in the intake feature. This feature adds the state filter.

A state filter offers every active state, and every inactive state that is still the current state of a service, or all states. The order is the installation's order. An empty choice means all states. The service list address keeps the chosen state. The search and the state filter apply together.

An empty shop says there are no services, as in intake. A search or a state filter that matches nothing says "No service matches that search." in English and "Bu aramaya uyan servis yok." in Turkish.

Opening a row opens the service.

The customer detail lists that customer's services, as in intake. The operator changes a state only from the service detail.

## Changing a state

The service detail shows the current state. The operator picks a new state, writes a change note or leaves it empty, adds one photo or adds none, and saves.

The picker offers the active states, in the installation's order. An inactive current state is shown and is not in the picker.

The change note is optional. A photo is optional. The photo is JPEG, PNG, or WebP, and at most 8 MB. One change holds one photo. That photo does not count toward the service's 30 photos.

A refused photo or a same-state save writes nothing. The stored state stays as it was.

The correction screen still edits the device notes, the parts, and the service photos. The brand, the model, the serial number, and the customer for this service are corrected as in the device feature. The correction screen does not change the state.

## History

The service detail shows the state history, newest first.

A row shows the time, the actor's username, the previous state, the new state, the change note when one was entered, and the photo when one was added.

The history is append-only. This feature does not edit or delete a row.

A successful state change writes the audit event already named in the audit feature. That audit row shows the previous state and the new state. It does not contain the change note or the photo file.

## Not in this feature

- A separate warranty field.
- A required sequence of states.
- Adding, renaming, deactivating, reordering, or deleting a state. That is the settings feature.
- A note or a photo without a state change.
- Editing or deleting a history row.
- Changing the state from the list or from the customer detail.
- Choosing a state on the registration screen.
- More than one photo on one change.
- Using spare-part stock on a service. That is the spare parts feature. A state change does not change stock.
