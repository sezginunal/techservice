# Home

Status: approved.

A signed-in technician, owner, or admin opens Home and sees the shop summary. The three roles see the same summary. A signed-out request is unauthorized.

The summary is counted on the server. The screen does not download the service table.

This replaces the home screen in the access feature, which showed only the signed-in name and role. The name and role stay in the sidebar. The access feature is updated to say that.

A day is a calendar day in Europe/Istanbul. Today is included.

A count is a count of services. A second service for a device already seen in the window counts again.

## Registrations

The screen shows how many services were opened in the last 7 days and in the last 30 days.

The last 7 days are today and the six days before it. The last 30 days are today and the twenty-nine days before it. A service counts on the Istanbul day it was opened.

A chart has one bar for each of those 30 days. The last 7 bars are darker than the earlier bars.

The title is "Registrations" in English and "Kayıtlar" in Turkish. The caption says that each bar is one day and that the darker bars are the last 7 days.

## In the shop

Two figures sit under the chart.

- In the shop — Dükkanda. Every service that is not delivered. A cancelled service is included. The delivered count sits beside it: "{n} delivered" in English and "{n} teslim edildi" in Turkish.
- Idle 7 days or more — 7 gündür aynı durumda. This is the count defined in the unchanged section.

## State queues

A state queue is shown for each state the settings feature marks "Show on Home", in the installation's order. `registered`, `delivered`, and `cancelled` have no queue. Two states are two queues. The summary includes these rows. The screen does not download the service table to build them.

The title is the state's name. Choosing the title opens the service list filtered to that state.

The queue lists every service in that state, including a service that entered the state today. The longest wait is first. The list shows at most 12 services. When more than 12 qualify, the screen says "Showing the longest waits." in English and "En uzun bekleyenler gösteriliyor." in Turkish.

A row shows the brand, the model, the serial number, the customer name, the number of full days, and the time that wait started. The day count is "Less than a day" / "1 günden az", "1 day" / "1 gün", or "{n} days" / "{n} gün". The time is labeled "Same status since" in English and "Bu durumda" in Turkish. Opening the row opens the service.

The clock and the row color are the ones in the unchanged section. A wait under one full day stays neutral.

When the state has no service, the queue stays and says "No service is in this state." in English and "Bu durumda servis yok." in Turkish.

## By status

The rows are the active states, including a state with no service, and any inactive state that is still the current state of a service. The order is the installation's order from the settings feature. Each row shows the state name and the count.

Choosing a state opens the service list filtered to that state. The service list address keeps the chosen state. An empty choice on the service list still means all states. The device status feature is updated to say that.

The title is "By status" in English and "Duruma göre" in Turkish. The section links to the service list with no state filter. That link is "All services" in English and "Tüm servisler" in Turkish.

## By model

The last 30 days are grouped by brand and model. The highest count is first. The screen shows at most 8 rows. Everything past those 8 is one row: "Other" in English and "Diğer" in Turkish.

A row shows the catalog image when the catalog has one.

The title is "By model" in English and "Modele göre" in Turkish. The caption says "These services were opened in the last 30 days." in English and "Bu servisler son 30 günde açıldı." in Turkish.

When nothing was opened in those 30 days, the screen says "No service was opened in the last 30 days." in English and "Son 30 günde servis açılmadı." in Turkish.

## Status unchanged

A service is listed when it is neither delivered nor cancelled and its state has stayed the same for at least one full day. A full day is 24 hours. The clock starts at the latest state change. When the state has never changed, the clock starts at registration. A state the shop added uses this same clock.

Delivered and cancelled services are absent.

The list shows at most 12 services, the longest wait first. When more than 12 qualify, the screen says "Showing the longest waits." in English and "En uzun bekleyenler gösteriliyor." in Turkish.

A row shows the brand, the model, the serial number, the customer name, the state, the number of full days, and the time that wait started. The day count is "1 day" / "1 gün", or "{n} days" / "{n} gün". The time is labeled "Same status since" in English and "Bu durumda" in Turkish. Opening the row opens the service.

The row color deepens as the day count grows:

- 1 or 2 days stays neutral.
- 3 to 6 days is amber.
- 7 to 13 days is orange.
- 14 to 29 days is red.
- 30 days or more is dark red.

The title is "Status unchanged" in English and "Durumu değişmeyenler" in Turkish. The caption says "Delivered and cancelled services are left out. A longer wait is a deeper red." in English and "Teslim edilen ve iptal edilen servisler yok. Bekleme uzadıkça renk kızarır." in Turkish.

When no listed service has waited a full day, the screen says "No service has waited a full day." in English and "Bir tam gündür bekleyen servis yok." in Turkish.

The figure "Idle 7 days or more" counts every service that is neither delivered nor cancelled and has waited 7 full days or more. That count includes services past the 12 shown.

## Latest registrations

The 5 newest services. A row shows the brand, the model, the serial number, the state, and the registration time. The catalog image shows when the catalog has one. Opening the row opens the service.

The title is "Latest registrations" in English and "Son kayıtlar" in Turkish.

When the shop has no services, this section says "No services yet." in English and "Henüz servis yok." in Turkish.

## Empty shop

Every count is zero. The status rows are the active states, each at zero. A state queue appears only for a state shown on Home, and that queue uses its empty sentence. The model chart and the unchanged list use their empty sentences.

## Not in this feature

- Changing a service, a state, or a customer from Home.
- Choosing which states have a queue. That is the settings feature.
- A queue that combines two states.
- A date range the operator picks.
- Delivered or cancelled services in the unchanged list.
- A service-list filter for "not delivered" or for idle services.
