# Intake

Status: approved.

A signed-in technician, owner, or admin records a customer and a device, then finds and corrects that record later. The three roles see and do the same work. A signed-out request is unauthorized.

The signed-in navigation adds Devices and Customers. User management stays as it is in access.

## Customer

- A customer is a separate record. A device belongs to one customer.
- The name is required. The phone and the email are optional.
- The name is one field.
- A required value is refused when it is empty or only space. The screen names the field.
- The phone is kept as the operator typed it, with surrounding space removed. The screen does not check a phone format.
- When an email is entered, a value that is not an email address is refused.
- The same name may exist on more than one customer. The screen does not merge them.
- Changing the name, phone, or email updates that customer. Every device of that customer shows the new values.
- This feature does not delete a customer.

## Device

- The device name, brand, model, and condition note are required.
- The condition note is the operator's comment on the condition of the device.
- The serial number is optional.
- On a successful save the system assigns a UUID. The operator does not type it. The device list, the device detail, and the page address use that UUID. It does not change later.
- The status is `registered`. The screen shows it as "Registered" in English and "Kayıt yapıldı" in Turkish.
- This feature has no control to change the status. Later work adds the other statuses.
- The device stays with the customer it was saved with.
- The list and the detail show the registration date. The newest device is first.
- This feature does not delete a device.

## Parts

- Parts are optional. Each part is its own row. Two cables are two rows.
- A part name and that part's condition note are required.
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
- Search matches the name, the phone, or the email. Each result shows the name, the phone, and the email, so two people with the same name can be told apart.
- The operator either selects an existing customer or enters a new one. Selecting an existing customer shows that person's name, phone, and email and does not change them on this screen.
- The device section takes the name, brand, model, serial number, and condition note.
- The parts section adds part rows. Each row takes the part name, the condition note, and that part's photos.
- The device photos sit on the device.
- One save writes the new customer, when the operator entered one, and the device with its parts and photos.
- A failed save writes nothing. The screen names each missing required field.
- Saving with no selected customer and no new name is refused.

## Device list and detail

- The device list and the customer list follow the lists rule.
- The device list shows the newest registration first.
- A row shows the UUID, device name, brand, model, customer name, status, and registration date. It does not show photos or parts.
- One search box matches the device name, brand, model, serial number, customer name, phone, or UUID.
- An empty shop says there are no devices. A search that matches nothing says so.
- Opening a row opens the device. The detail shows the UUID, status, registration date, name, brand, model, serial number when present, condition note, customer name, phone, and email, the device photos, and each part with its note and photos.
- The parts section is visible when there are no parts.
- From the detail the operator corrects the device name, brand, model, serial number, condition note, parts, and photos, and corrects the customer's name, phone, and email.
- The correction is kept on save. The UUID and the status stay as they are.

## Customer list and detail

- The customer list is ordered by name.
- A row shows the name, phone, and email. The phone and the email are blank when they were not entered.
- One search box matches the name, the phone, or the email.
- An empty book says there are no customers. A search that matches nothing says so.
- The detail shows the name, phone, and email, and the devices of that customer. A device row shows the device name, brand, model, status, and registration date. Opening it goes to the device.
- The operator corrects the name, phone, and email from this detail.
- When the customer has no devices, the detail says so.
- From this detail the operator starts a new device. The registration screen opens with this customer already selected and shown. The operator can select a different customer or enter a new one before saving.

## Not in this feature

- Any status other than `registered`, and any control that changes status.
- A camera capture.
- Deleting a customer or a device.
- Moving a device to another customer.
- How the device arrived.
- A quantity on a part.
