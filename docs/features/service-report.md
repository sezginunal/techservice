# Service report

Status: proposed.

A signed-in technician, owner, or admin opens a printable final service report from a service. The three roles see and do the same work. A signed-out request is unauthorized. A person who cannot open the service cannot open or print its report.

The report is a paper draft. It shows the service as it is at print time. It stores no amounts and no filled prices. The operator prints it and writes money amounts by hand. The system still has no price field.

## Where

The service detail has the report controls. Every state is allowed, including `delivered` and `cancelled`. There is no separate report page and no navigation entry.

The operator chooses summary or detailed, then prints. The browser print dialog is the printer step. This feature does not save a report file and does not send the report elsewhere.

## Layouts

Two layouts. Both share the same header, customer, device, dates, current state, spare-part list, fee blanks, QR code, and signature lines. The detailed layout also prints the state history.

### Shared header

- The shop brand name and logo as in the brand feature.
- The service UUID as text.
- A QR code whose payload is that same service UUID as plain text. The QR does not encode a URL.

### Customer and device

- Customer name, phone, email when set, address when set.
- Device brand, model, and serial number.

### Dates and current state

- Registration date: the service's registration time.
- Exit date: the time the service entered `delivered`. When the current state is not `delivered`, that field is empty on the page and leaves a blank line for handwriting if needed.
- Current state: the catalog name for the active language. The code is never shown. This line is on every report, in every state.

### Spare parts

Only spare parts taken from stock for this service. Intake parts that arrived with the device are not on the report.

The list is the net quantity still on the service: uses of that spare part minus returns of that spare part. A part whose net quantity is 0 is omitted. Each row shows the spare part's current brand and model, the net quantity, and an empty amount column for handwriting.

When no spare part has a net quantity above 0, the section stays and says there were none. The English sentence is "No spare parts used." The Turkish sentence is "Yedek parça kullanılmadı."

### Fee blanks

Below the spare-part list:

- One row labeled "Technical service fee" in English and "Teknik servis ücreti" in Turkish, with an empty amount line.
- One row labeled "Total" in English and "Toplam" in Turkish, with an empty amount line.

The report never fills those amounts.

### Signatures

Two signature lines:

- "Customer" in English and "Müşteri" in Turkish.
- "Authorized" in English and "Yetkili" in Turkish.

Each line has space to sign by hand.

### Summary

The summary layout stops after the fee blanks and the signatures. It does not print the state history.

### Detailed

The detailed layout adds the state history after the spare-part list and before the fee blanks.

Each history row shows the state name for the active language, the time the service entered that state, and the time it left that state. Leaving time is the time of the next change. For the current state, the leaving time is empty. Change notes and status-change photos are not printed. Actor usernames are not printed.

The history order is oldest first, so the paper reads as a timeline.

## Language

Every label on the report follows the active language, as on the rest of the app. The shop brand name follows the brand feature and does not change with the language.

## Not in this feature

- A price, a currency, or a stored amount on a spare part or on the service.
- Filling the fee or total lines in software.
- Intake parts that came with the device.
- Service comments, device notes, status-change notes, or photos on the report.
- A PDF file, an email, or a saved report record.
- A QR payload other than the service UUID as plain text.
- Printing from the service list, the device page, the customer page, or Home.
