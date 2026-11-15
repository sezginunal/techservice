# Settings

Status: approved.

Each installation has its own service-state list, stored in that installation's database. A second shop that installs this project writes its own list. One installation has one list. Two services in the same shop use that list.

An owner or an admin opens Settings and edits that list. A technician has no navigation entry. Opening the address does not show the page. A signed-out request is unauthorized. A technician's request is forbidden.

The navigation shows Settings to an owner and an admin. The title is "Settings" in English and "Ayarlar" in Turkish.

## Service states

The page has one section. The title is "Service states" in English and "Servis durumları" in Turkish. The caption says "An inactive state stays on services that already use it. Nobody can choose it for a later change." in English and "Kapalı bir durum, o durumda olan servislerde durur. Bundan sonra kimse onu seçemez." in Turkish.

The screen shows every state of this installation, in the installation's order. The list is the whole catalog. It is not paged. There is no search.

A row shows the name for the current language as its title and the other language's name under that title. A state the shop can rename shows both name fields instead of that title. Other screens show one name: the English name when the interface is English, and the Turkish name when the interface is Turkish.

Each row is its own card. A handle on the left drags that row. The handle is named "Drag to reorder" in English and "Sırayı değiştirmek için sürükle" in Turkish. The operator drops the row in a new place. The row takes that place and the list saves. Dropping a row where it already is writes nothing. There are no move buttons. The new order is the order on this page, on the state picker, on the service-list filter, on Home's status rows, and on Home's state queues.

## Active

`registered`, `delivered`, and `cancelled` have no active control.

- `registered` says "Registration uses this state." in English and "Kayıt bu durumu kullanır." in Turkish.
- `delivered` says "Delivery uses this state." in English and "Teslim bu durumu kullanır." in Turkish.
- `cancelled` says "Cancelling a service uses this state." in English and "İptal bu durumu kullanır." in Turkish.

A request that deactivates one of those three is refused and writes nothing.

Every other row offers an active control. Active is "Active" in English and "Açık" in Turkish. Inactive is "Inactive" in English and "Kapalı" in Turkish. The control saves on its own.

An inactive state remains stored. A service already in that state stays there. The device status feature shows it and refuses a later change into it. Making a state inactive does not change its Home queue.

This feature does not delete a state.

## Home queue

`registered`, `delivered`, and `cancelled` have no Home control. A request that shows one of them on Home is refused and writes nothing.

Every other row offers "Show on Home" in English and "Ana sayfada göster" in Turkish. The control saves on its own. It starts off. The two choices are independent: a state can be inactive and still shown, or active and hidden.

While the choice is on, Home shows that state's queue, including when the count is zero. While the choice is off, Home has no queue for that state. The home feature says what the queue shows. Two states are two queues. This feature adds no combined queue.

## Names

A state has a code, an English name, and a Turkish name. A service stores the code. Every screen reads the name from this catalog: the English name when the interface is English, and the Turkish name when the interface is Turkish. The code is never shown.

The fixed states are `registered`, `repair_warranty`, `repair_out_of_warranty`, `cancelled`, and `delivered`. Their codes and both names are fixed. The screen shows those names as text. A request that renames a fixed state is refused and writes nothing.

Any other state, including one the shop adds, has fields for both names. The operator enters both and saves them together. The labels are "English name" in English and "İngilizce ad" in Turkish, and "Turkish name" in English and "Türkçe ad" in Turkish. One name without the other is refused.

Both names are required. The screen removes surrounding space. An empty name is refused. A name longer than 80 characters is refused. The screen names the field.

The English name and the Turkish name are each unique in this installation, ignoring case. A duplicate is refused. The screen says "That name is already in use." in English and "Bu ad zaten kullanılıyor." in Turkish.

The operator does not type a code. An added state receives a code from the English name at the moment it is added: lower case, a run of other characters becomes one underscore, and underscores at the ends are removed. An empty code, or a code already stored, is refused with the same sentence as a duplicate name. A later rename changes the names and leaves the code as it was. Services that store that code keep it.

A rename changes the name everywhere that code is shown, including service history, Home, and older status-change audit rows. The audit row for the rename keeps the previous names and the new names.

An added state has no rule of its own. Warranty, a second service, the in-shop count, and aging stay the rules named in the device status, device, and home features.

## Adding a state

The add form is its own block under the list. It has the English name, the Turkish name, the Home control, and "Add state" in English and "Durum ekle" in Turkish. Both names are required. The Home control starts off.

A successful add stores an active state at the end of the list and keeps the Home choice the operator made. A refused add writes nothing.

## First catalog

The first start that has this catalog writes these states, in this order. All five are active. The two repair states are not shown on Home. A later start does not rewrite a name, an active flag, a code, a position, or a Home choice.

1. `registered` — Registered — Kayıt yapıldı
2. `repair_warranty` — Repair under warranty — Garanti kapsamında tamiri yapılıyor
3. `repair_out_of_warranty` — Repair outside warranty — Garanti dışında tamiri yapılıyor
4. `cancelled` — Cancelled — İptal edildi
5. `delivered` — Delivered — Teslim edildi

A new installation writes only these five.

An installation that already has one of the codes below on a service, or in a service's status history, writes that code after the five, in the order below. The state is active when a service's current state is that code. It is inactive when the code appears only in history. It is not shown on Home. The shop can rename it, deactivate it, reorder it, and show it on Home. A code in this list that no stored service and no stored history row uses is not written.

6. `inspecting` — Inspecting — Ön inceleme yapılıyor
7. `waiting_partner_part` — Waiting for a partner part — İş ortağından yedek parça bekleniyor
8. `waiting_own_part` — Waiting for a part we will buy — Kendi satın alacağımız yedek parça bekleniyor
9. `final_testing` — Final tests — Final testleri yapılıyor
10. `preparing_delivery` — Preparing for delivery — Teslim için hazırlanıyor
11. `in_storage` — In storage — Depoda
12. `left_storage` — Left storage — Depodan çıkış yapıldı

## Audit

A successful add, a successful rename, active change, or Home-queue change, and a successful reorder write the audit rows named in the audit feature. A save that changes nothing writes nothing.

## Not in this feature

- A second state list in one installation.
- Deleting a state.
- Deactivating `registered`, `delivered`, or `cancelled`.
- Showing `registered`, `delivered`, or `cancelled` on Home.
- Renaming a fixed state.
- A rule that belongs only to a state the shop added.
- A required sequence of states.
- Changing a service from this page.
- Another settings section.
