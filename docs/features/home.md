# Home

Status: proposed.

A signed-in technician, owner, or admin opens Home and sees the shop summary. The three roles see the same summary. A signed-out request is unauthorized.

The summary is counted on the server. The screen does not download the device table.

This replaces the home screen in the access feature, which showed only the signed-in name and role. The name and role stay in the sidebar. The access feature is updated to say that.

A day is a calendar day in Europe/Istanbul. Today is included.

## Registrations

The screen shows how many devices were registered in the last 7 days and in the last 30 days.

The last 7 days are today and the six days before it. The last 30 days are today and the twenty-nine days before it. A device counts on the Istanbul day of its registration time.

A chart has one bar for each of those 30 days. The last 7 bars are darker than the earlier bars.

The title is "Registrations" in English and "Kayıtlar" in Turkish. The caption says that each bar is one day and that the darker bars are the last 7 days.

## In the shop

Four figures sit under the chart.

- In the shop — Dükkanda. Every device that is not delivered. The delivered count sits beside it: "{n} delivered" in English and "{n} teslim edildi" in Turkish.
- Idle 7 days or more — 7 gündür aynı durumda. This is the count defined in the unchanged section.
- Waiting for a part — Parça bekleyen. `waiting_partner_part` and `waiting_own_part` together.
- Ready to deliver — Teslime hazır. `preparing_delivery`.

## By status

Every state from the device status feature is shown, including a state with no device. The order is the order in that feature. Each row shows the state name and the count.

Choosing a state opens the device list filtered to that state. The device list address keeps the chosen state. An empty choice on the device list still means all states. The device status feature is updated to say that.

The title is "By status" in English and "Duruma göre" in Turkish. The section links to the device list with no state filter. That link is "All devices" in English and "Tüm cihazlar" in Turkish.

## By model

The last 30 days are grouped by brand and model. The highest count is first. The screen shows at most 8 rows. Everything past those 8 is one row: "Other" in English and "Diğer" in Turkish.

A row shows the catalog image when the catalog has one.

The title is "By model" in English and "Modele göre" in Turkish. The caption says these devices were registered in the last 30 days.

When nothing was registered in those 30 days, the screen says "No device was registered in the last 30 days." in English and "Son 30 günde kayıt yok." in Turkish.

## Status unchanged

A device is listed when it is not delivered and its state has stayed the same for at least one full day. A full day is 24 hours. The clock starts at the latest state change. When the state has never changed, the clock starts at registration.

Delivered devices are absent.

The list shows at most 12 devices, the longest wait first. When more than 12 qualify, the screen says "Showing the longest waits." in English and "En uzun bekleyenler gösteriliyor." in Turkish.

A row shows the device name, brand, model, customer name, state, the number of full days, and the time that wait started. The day count is "1 day" / "1 gün", or "{n} days" / "{n} gün". The time is labeled "Same status since" in English and "Bu durumda" in Turkish. Opening the row opens the device.

The row color deepens as the day count grows:

- 1 or 2 days stays neutral.
- 3 to 6 days is amber.
- 7 to 13 days is orange.
- 14 to 29 days is red.
- 30 days or more is dark red.

The title is "Status unchanged" in English and "Durumu değişmeyenler" in Turkish. The caption says "Delivered devices are left out. A longer wait is a deeper red." in English and "Teslim edilenler yok. Bekleme uzadıkça renk kızarır." in Turkish.

When no open device has waited a full day, the screen says "No open device has waited a full day." in English and "Bir tam gündür bekleyen açık cihaz yok." in Turkish.

The figure "Idle 7 days or more" counts every non-delivered device that has waited 7 full days or more. That count includes devices past the 12 shown.

## Latest registrations

The 5 newest registrations. A row shows the device name, brand, model, state, and registration time. The catalog image shows when the catalog has one. Opening the row opens the device.

The title is "Latest registrations" in English and "Son kayıtlar" in Turkish.

When the shop has no devices, this section says "No devices yet." in English and "Henüz cihaz yok." in Turkish.

## Empty shop

Every count is zero. The status rows stay, each at zero. The model chart and the unchanged list use their empty sentences.

## Not in this feature

- Changing a device, a state, or a customer from Home.
- A date range the operator picks.
- Delivered devices in the unchanged list.
- A device-list filter for "not delivered" or for idle devices.
