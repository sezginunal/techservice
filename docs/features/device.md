# Device

Status: approved.

A signed-in technician, owner, or admin identifies a physical device and opens a service for it, then reads that device's services later. The three roles see and do the same work. A signed-out request is unauthorized.

A device is one physical machine. A service is one time that machine is in the shop. The customer on a service is the person who brought it in. A later sale, a gift, or another person carrying a company machine is not a separate record. Two services of one device may name different customers.

Intake owns the customer, the parts, the photos, and the registration screen. This feature owns the identity and the link between a device, its services, and the customer named on each service.

## Identity

- The brand, the model, and the serial number are required.
- A required value is refused when it is empty or only space. The screen names the field. The serial field is "Serial number" in English and "Seri numarası" in Turkish. An empty serial number says "Enter the serial number." in English and "Seri numarasını gir." in Turkish.
- The stored brand, model, and serial number are the text left in the fields, with surrounding space removed.
- Those three are one device. The match ignores surrounding space and letter case.
- The same serial number on a different brand, or on a different model, is a different device.
- A machine whose serial number cannot be read is not registered.
- The system assigns a UUID to the device. The operator does not type it. The device page address uses that UUID. It does not change later.
- The system assigns a UUID to the service. The operator does not type it. The service list, the service detail, and the page address use that UUID. It does not change later.
- This feature does not delete a device or a service.

## Registration

- The registration screen is the intake screen.
- When the brand, the model, and the serial number are all present, the screen looks up the device. The lookup waits until the operator pauses or leaves the field.
- When the three match a device, the screen shows that device's services before the save. The newest service is first. The list follows the lists rule. A row shows the registration date, the customer name, the phone, and the state. The heading is "Earlier services" in English and "Önceki servisler" in Turkish.
- When the three match no device, that list is absent.
- The operator selects the customer for this service, or enters a new one, as in intake. The save does not change the customer on another service.
- One save writes the new customer, when the operator entered one. When no device matches, the save writes the device and the first service. When a device matches, the save writes a service on that device. When the brand and model match no catalog row, that save also stores the pair, as in the device catalog.
- A failed save writes nothing. The server is the check that the three match an existing device.
- A device that has a service other than `delivered` or `cancelled` cannot take another service. The save is refused and writes nothing. The screen says "This device is already in the shop." in English and "Bu cihaz zaten dükkanda." in Turkish, and offers "Open the service" in English and "Servisi aç" in Turkish. That control opens the service that is still in the shop.
- The save writes the audit rows named in the audit feature.

## Device page

- The service detail shows the brand, the model, and the serial number as one link to the device page.
- The device page shows the brand, the model, the serial number, and the catalog image when the device catalog has one.
- The page lists that device's services. The newest service is first. The list follows the lists rule. A row shows the registration date, the customer name, and the state. Opening a row opens the service.
- The heading of that list is "Services" in English and "Servisler" in Turkish.
- The device page does not change the state, the notes, the parts, the photos, or the customer.

## Correction

- From the service detail the operator corrects the brand, the model, and the serial number. The correction is that device's identity. Every service of the device shows the new values.
- A correction that matches another device is refused and writes nothing. The screen says "A device with this brand, model, and serial number is already registered." in English and "Bu marka, model ve seri numarasıyla kayıtlı bir cihaz var." in Turkish. The service stays on its device.
- When the corrected brand and model match no catalog row, the save also stores that pair, as in the device catalog. A refused correction does not.
- From the service detail the operator changes the customer for this service. The customer section works as it does on the registration screen. The other services of the device keep their customers.
- Correcting the selected customer's name, phone, email, or address updates that customer, as in intake.
- The save writes the audit rows named in the audit feature.

## Not in this feature

- A stored owner, a sale, or a transfer between customers.
- A second service while one service is neither `delivered` nor `cancelled`.
- Moving a service onto a different device.
- Deleting a device or a service.
- A device page in the navigation. The service list is the navigation entry, as in intake.
- Spare-part stock. A use is recorded on the service detail. That is the spare parts feature.
