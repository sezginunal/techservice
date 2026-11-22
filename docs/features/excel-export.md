# Excel export

Status: approved.

An owner or an admin downloads the rows of a list as an Excel file. A technician does not see the control and cannot call the export. A signed-out request is unauthorized.

## Lists

Export exists on these three list screens only:

- Customers
- Services
- Spare parts

Audit and Users have no export. Detail pages and other screens have no export.

## Who

- An owner or an admin sees the export control on each of those three lists.
- A technician does not see the control on any of those lists.
- The backend refuses an export from a technician or from a signed-out caller. Hiding the control is not the security check.

## The screen

Each of the three lists has one export control in the page header, at the start of that list's actions.

The control label is "Export Excel" in English and "Excel'e aktar" in Turkish.

Choosing the control downloads one `.xlsx` file. The page stays on that list. The download uses the search and filters that list already applies:

- Customers: the search box.
- Services: the search box and the state filter.
- Spare parts: the search box.

An empty search and an empty state filter export every row of that list. A search or a state filter that matches nothing still downloads a file that has only the header row.

While the file is building, the control is disabled. A failed download shows an error on that page. The list under the control does not change.

## Columns

The file has one sheet. The sheet name matches the list title in the current interface language.

The header row uses the same column labels the list shows, in the current interface language. The action column with the chevron is not in the file.

- Customers: name, phone, email, address.
- Services: record id, brand, model, serial number, customer, status, registered on. The status cell is the state's name in the current interface language. The registered-on cell is the same date-time text the list shows for that language.
- Spare parts: brand, model, quantity on hand.

The row order is the same order that list uses. Every matching row is in the file. The export is not limited to the rows already loaded in the browser window.

## Not in this feature

- Audit export.
- Users export.
- Export from a detail page, Home, Settings, or any other screen.
- Choosing columns, renaming sheets, or changing the file format.
- A scheduled or emailed export.
