# Intake

Status: approved.

A signed-in technician, owner, or admin records a customer and a device, then finds and corrects that record later. The three roles see and do the same work. A signed-out request is unauthorized.

The signed-in navigation adds Devices and Customers. User management stays as it is in access.

## Customer

- A customer is a separate record. A device belongs to one customer.
- The name and the phone are required. The email and the address are optional.
- The name is one field.
- A required value is refused when it is empty or only space. The screen names the field.
- A value that is not a phone number is refused. The shape is the frontend input rule.
- The stored phone is the digits of that number. Spaces, parentheses, hyphens, and a leading `+` are not stored. A leading `0090`, a leading `90`, and a single leading `0` are removed when they stand in front of the same 10-digit number, including a trunk `0` after `90`. `555 555 55 55`, `5555555555`, `05555555555`, `0555555 55 55`, and `+90 555 555 55 55` are one phone. The stored value is 10 to 15 digits.
- That stored phone belongs to one customer. Saving a second customer with the same phone is refused. The screen says the phone is already registered. The operator searches and selects the existing customer. The save does not change that customer's name.
- When an email is entered, a value that is not an email address is refused.
- The same name may exist on more than one customer. A search by that name lists every match.
- The address is kept as the operator typed it, with surrounding space removed. The screen does not check an address format.
- Changing the name, phone, email, or address updates that customer. Every device of that customer shows the new values.
- This feature does not delete a customer.

## Device

- The brand and the model are required.
- The device notes are the operator's text about the device. They are optional.
- The serial number is optional.
- On a successful save the system assigns a UUID. The operator does not type it. The device list, the device detail, and the page address use that UUID. It does not change later.
- The status is `registered`. The screen shows it as "Registered" in English and "Kayıt yapıldı" in Turkish.
- This feature has no control to change the status. Later work adds the other statuses.
- The device stays with the customer it was saved with.
- The list and the detail show the registration date. The newest device is first.
- This feature does not delete a device.

## Parts

- Parts are optional. Each part is its own row. Two cables are two rows.
- A part name is required. That part's note is optional.
- A part has no brand, model, or serial number.
- The operator can add, change, and remove a part while registering or while correcting the device.
- Removing a part removes that part's photos.

## Photos

- The operator adds photos from the file picker, on the device and on a part.
- A photo is JPEG, PNG, or WebP, and at most 8 MB.
- A device, including all of its parts, holds at most 30 photos after the save.
- The screen refuses a photo that has another type, is over 8 MB, or would pass 30. It says which of those applies.
- If a photo is refused, that save does not complete. A failed correction leaves the stored record as it was.
- The camera is a later feature.

## Registration screen

- One screen registers the device.
- The customer section comes first. The operator searches existing customers or enters a new one.
- Search matches the name, the phone, the email, or the address. A phone search uses the same digit form, so the spaced, trunk-zero, and `+90` forms find that customer. Each result shows the name, the phone, the email, and the address, so two people with the same name can be told apart.
- The operator either selects an existing customer or enters a new one. Selecting an existing customer shows that person's name, phone, email, and address and does not change them on this screen.
- The device section takes the brand, model, serial number, and device notes.
- The parts section adds part rows. Each row takes the part name, the part note, and that part's photos.
- The device photos sit on the device.
- One save writes the new customer, when the operator entered one, and the device with its parts and photos.
- A failed save writes nothing. The screen names each missing required field.
- Saving with no selected customer, no name, or no phone is refused.

## Device list and detail

- The device list and the customer list follow the lists rule.
- The device list shows the newest registration first.
- A row shows the UUID, brand, model, customer name, status, and registration date. It does not show photos or parts.
- One search box matches the brand, model, serial number, customer name, phone, address, or UUID. A phone search uses the same digit form as the customer record.
- An empty shop says there are no devices. A search that matches nothing says so.
- Opening a row opens the device. The detail shows the UUID, status, registration date, brand, model, serial number when present, device notes, customer name, phone, email, and address, the device photos, and each part with its note and photos.
- The parts section is visible when there are no parts.
- From the detail the operator corrects the brand, model, serial number, device notes, parts, and photos, and corrects the customer's name, phone, email, and address.
- The correction is kept on save. The UUID and the status stay as they are.

## Customer list and detail

- The customer list is ordered by name.
- A row shows the name, phone, email, and address. The email and the address are blank when they were not entered.
- One search box matches the name, the phone, the email, or the address. A name lists every customer with that name. A phone search uses the same digit form.
- An empty book says there are no customers. A search that matches nothing says so.
- The detail shows the name, phone, email, and address, and the devices of that customer. A device row shows the brand, model, status, and registration date. Opening it goes to the device.
- The operator corrects the name, phone, email, and address from this detail.
- When the customer has no devices, the detail says so.
- From this detail the operator starts a new device. The registration screen opens with this customer already selected and shown. The operator can select a different customer or enter a new one before saving.

## Not in this feature

- Any status other than `registered`, and any control that changes status.
- A camera capture.
- Deleting a customer or a device.
- Moving a device to another customer.
- How the device arrived.
- A quantity on a part.
