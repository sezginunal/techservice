# Spare parts

Status: approved.

A signed-in technician, owner, or admin keeps spare-part stock and records each use on a service. The three roles see and do the same work. A signed-out request is unauthorized.

The signed-in navigation adds Spare parts after Models. The label is "Spare parts" in English and "Yedek parçalar" in Turkish. The list address is `/spare-parts`.

The parts on a service stay as they are in the intake feature. Those rows have a name, a note, and photos. A spare part is a stock line. It has a brand, a model, and a quantity on hand.

## Identity

- The brand and the model are required. A spare part has no separate name.
- A required value is refused when it is empty or only space. The screen names the field, with the same sentences as the device brand and model.
- A brand or a model longer than 200 characters is refused. The screen names the field.
- The stored brand and the stored model are the text left in the fields, with surrounding space removed.
- Those two are one spare part. The match ignores surrounding space and letter case.
- The system assigns a UUID. The operator does not type it. The spare part address is `/spare-parts/` followed by that UUID. It does not change later.
- This feature does not delete a spare part.

The same brand and model are one line for every physical kind of part the operator types that way. Two kinds that share a brand and a model share one count.

## Quantity

A quantity is a whole number. The on-hand quantity is never below 0 and never above 99999.

The operator enters a quantity in these cases:

- Creating a spare part: 0 through 99999. 0 leaves the part with nothing on hand.
- Adding stock: 1 through 99999. The result must still be at most 99999.
- Using a part on a service: 1 through the quantity on hand.
- Returning a part from a service: 1 through the quantity that service can return.
- Writing stock off: 1 through the quantity on hand.

An empty quantity is refused. The screen says "Enter the quantity." in English and "Adedi gir." in Turkish.

A value outside the range for that case is refused. The screen says "Enter a whole number from {min} to {max}." in English and "{min} ile {max} arasında bir tam sayı gir." in Turkish. The screen checks this before the request.

The server checks the quantity on hand again on a use, a return, and a write-off. When that quantity has fallen and the entered quantity is now too high, the save is refused. The screen says "Only {n} are in stock." in English and "Stokta {n} adet var." in Turkish.

An add or a return that would put more than 99999 on hand is refused. The screen says "Stock cannot pass 99999." in English and "Stok 99999 adedi geçemez." in Turkish.

A refused save writes nothing. The server is the check. When two saves would take the same last unit, the first save writes and the second is refused.

## Spare part list

The list follows the lists rule. The order is the brand, then the model, then the id. The operator does not choose another order.

The lead is "Parts in stock, by brand." in English and "Stoktaki parçalar, markaya göre." in Turkish.

A row shows the brand, the model, and the quantity on hand. The quantity label is "In stock" in English and "Stok" in Turkish. Opening a row opens the spare part.

One search box matches the brand or the model. The match ignores surrounding space and letter case. The search waits until the operator pauses or submits.

An empty stock says "No spare parts yet." in English and "Henüz yedek parça yok." in Turkish. A search that matches nothing says "No spare part matches that search." in English and "Bu aramaya uyan yedek parça yok." in Turkish.

## Creating a spare part

The create form is on the list screen, above the search. The fields are the brand, the model, and the quantity. The control is "Add spare part" in English and "Yedek parça ekle" in Turkish.

The brand field suggests the brands already stored on spare parts. When the brand matches a stored brand, the model field suggests that brand's stored models. The match ignores surrounding space and letter case. Selecting a suggestion fills the field with the stored text. The operator can type a brand or a model that is not suggested. These suggestions are spare parts already stored. They are not the device catalog.

A new brand and model create the spare part. A quantity above 0 also writes the opening receipt. A quantity of 0 writes the spare part and no receipt. The save stays on the list. The form clears, and the list loads again so the spare part is on it.

When the brand and model match a spare part that is already stored:

- A quantity above 0 adds that quantity to the existing part as a receipt. The save stays on the list. The form clears, and the list loads again. The screen says "This spare part is already registered. The quantity was added." in English and "Bu yedek parça zaten kayıtlı. Adet stoğa eklendi." in Turkish. That receipt has no note.
- A quantity of 0 writes nothing. The screen says "This spare part is already registered." in English and "Bu yedek parça zaten kayıtlı." in Turkish, and offers "Open the spare part" in English and "Yedek parçayı aç" in Turkish.

## Spare part page

The page shows the brand, the model, and the quantity on hand.

The operator corrects the brand and the model and saves. The control is "Save" in English and "Kaydet" in Turkish. A save that leaves both values the same writes nothing. A correction that matches another spare part is refused and writes nothing. The screen says "A spare part with this brand and model is already registered." in English and "Bu marka ve modelle kayıtlı bir yedek parça var." in Turkish. The uses stay on this part.

The brand and model fields use the same suggestions as the create form.

### Add stock

The operator enters a quantity and a note, or leaves the note empty, and saves. The control is "Add stock" in English and "Stoğa ekle" in Turkish. The note label is "Note (optional)" in English and "Not (isteğe bağlı)" in Turkish. A note longer than 4000 characters is refused. The screen names the field. A write-off note and a return note use the same limit.

A successful save writes a receipt and raises the quantity on hand.

### Write off

The operator enters a quantity and a note, or leaves the note empty, and saves. The control is "Write off" in English and "Fire" in Turkish. The note uses the same label as a receipt.

A write-off is stock that left without being used on a service: damaged, lost, or a count correction. A successful save lowers the quantity on hand.

While the quantity on hand is 0, the write-off save is disabled. The fields stay on the page.

### Used on services

This list is the places the part was used, and the returns from those services. Receipts and write-offs are not in this list.

The list follows the lists rule. The newest row is first.

The heading is "Used on services" in English and "Kullanıldığı servisler" in Turkish.

A row shows the time, the actor's username, the quantity, the kind, the device brand, the device model, the serial number, and the state. The kind is "Used" or "Returned" in English and "Kullanıldı" or "İade edildi" in Turkish. The state is the catalog name for the active language. The code is never shown. A return shows its note when one was entered.

The row stores the spare part and the state code. The screen shows the part's current brand and model, so a later correction changes the name on these rows. A renamed state shows its current name.

Opening a row opens the service.

An empty list says "This spare part has not been used." in English and "Bu yedek parça kullanılmadı." in Turkish.

### Stock changes

This list is the receipts and the write-offs. Uses and returns are not in this list.

The list follows the lists rule. The newest row is first.

The heading is "Stock changes" in English and "Stok hareketleri" in Turkish.

A row shows the time, the actor's username, the kind, the quantity, and the note when one was entered. The kind is "Added" or "Written off" in English and "Eklendi" or "Fire" in Turkish. The opening quantity is a receipt and appears here.

An empty list says "No stock changes yet." in English and "Henüz stok hareketi yok." in Turkish.

The spare part page does not use a part on a service and does not return one.

## Using a spare part

The service detail shows a spare-part section after the parts section from intake. The heading is "Spare parts" in English and "Yedek parçalar" in Turkish.

The section lists the spare parts still on this service. A part is still on the service when its uses on this service minus its returns on this service are at least 1. The list follows the lists rule. The order is the brand, then the model, then the id. A row shows the current brand, the current model, and that net quantity. Opening the brand and model opens the spare part.

The section is visible when it has no rows. It says "No spare part is on this service." in English and "Bu serviste yedek parça yok." in Turkish.

The operator uses a spare part from this section on any service, including a delivered or a cancelled service. The control is "Use a spare part" in English and "Yedek parça kullan" in Turkish. The registration screen, the correction screen, the device page, the customer page, and the spare part page have no such control.

The control opens a dialog. Opening it closes a return. The dialog lists every spare part, including a part with nothing on hand. The list follows the lists rule. The order is the brand, then the model, then the id. One search box matches the brand or the model, as on the spare part list. Opening the dialog loads the current quantity on hand.

A row shows the brand, the model, and the quantity on hand. A row with nothing on hand is shown and cannot be chosen. A row with stock can be chosen.

Choosing a part adds it in the dialog at quantity 1. The operator can change that quantity, from 1 through the quantity on hand, or remove the part. The remove control is "Remove" in English and "Çıkar" in Turkish. Choosing a part that is already in the dialog does not add a second line.

The save control is "Save" in English and "Kaydet" in Turkish. It is disabled until the dialog has a part. One save uses every line still in the dialog. Each line stores the time, the actor's username, the quantity, the service, and the service's state code at that moment. It lowers the quantity on hand by that quantity. It does not change the service state. A later use of the same part on the same service adds to that part's net quantity. A successful save closes the dialog. The part appears on the service, or its quantity there rises.

The save checks every quantity on hand again before it writes. A line that no longer fits refuses the save, names that part, and writes nothing from the dialog. The screen says "Only {n} are in stock." in English and "Stokta {n} adet var." in Turkish.

The control "Cancel" in English and "Vazgeç" in Turkish closes the dialog and writes nothing.

## Returning a spare part

A return puts stock back from a service. It is how a mistaken use is corrected. The use stays on the spare part page, in the used-on-services list.

Each row has the control "Return to stock" in English and "Stoğa iade" in Turkish. It opens a dialog for that part only. Opening it closes the use dialog. The quantity starts at 1. It is a whole number from 1 through that row's net quantity. The dialog asks "Return {n} to stock?" in English and "{n} adet stoğa iade edilsin mi?" in Turkish, with the quantity currently in the field. The save control is "Return to stock" in English and "Stoğa iade" in Turkish. The control "Cancel" in English and "Vazgeç" in Turkish closes the dialog and writes nothing.

One save returns that one part. The save stores the time, the actor's username, the quantity, an empty note, the service, and the service's state code at that moment. It raises the quantity on hand by that quantity. It does not change the service state. A return that would put more than 99999 on hand is refused and writes nothing. A successful save closes the dialog. A part whose net quantity falls to 0 leaves the list.

A return is not tied to one use row. Several uses of one part on one service add up to one net quantity.

## History

Receipts, write-offs, uses, and returns are append-only. This feature does not edit or delete a row.

A mistaken use is corrected by a return. A mistaken receipt or write-off is corrected by a later receipt or write-off.

## Audit

A successful create, correction, receipt, write-off, use, or return writes the audit row named in the audit feature. A save that changes nothing writes nothing. The audit row does not contain the note. The names on an audit row stay as they were when the row was written.

## Not in this feature

- A part name separate from the brand and the model.
- A price, a serial number for one unit, or a photo of the spare part.
- Deleting a spare part or a history row.
- Editing a receipt, a write-off, a use, or a return.
- Merging two spare parts into one line.
- Hiding a spare part that has nothing on hand.
- Using or returning a spare part from anywhere except the service detail.
- Changing the service state when stock changes.
- The device catalog as the source of the brand and the model.
- Spare parts on the service list, the device page, the customer page, or Home.
- A note on a use.
