# Service comments

Status: approved.

A signed-in technician, owner, or admin reads and writes comments on a service. The three roles see and do the same work. A signed-out request is unauthorized.

A comment is conversation on that service. It does not change the state, the device notes, the parts, the photos, or the customer. The device notes and the status-change notes stay as they are in the device status feature.

## Where

The service detail shows the comments after the state history. Every state is allowed, including `delivered` and `cancelled`. The list and the writer stay on that page. There is no separate comments page and no navigation entry.

## Who

Every signed-in role can read every comment on a service they can open. Every signed-in role can write a new comment. Every signed-in role can edit a comment they wrote. A user cannot edit another person's comment.

## List

The comments are one flat chronological list. The oldest comment is first. A later comment is a new row under the earlier ones. There is no reply thread and no nesting.

A row shows the time, the author's username, and the text. When the author has edited the comment, the row also shows that it was edited and the edit time.

An empty list says there are no comments yet. The English sentence is "No comments yet." The Turkish sentence is "Henüz yorum yok."

The section title is "Comments" in English and "Yorumlar" in Turkish.

## Writing

The operator types the text and saves. An empty text, or text that is only space, is refused and writes nothing. The screen says "Enter a comment." in English and "Yorum gir." in Turkish.

A comment longer than 4000 characters is refused. The screen names the field.

The stored text is the text left in the field, with surrounding space removed.

A successful save adds the row at the bottom of the list and clears the writer.

## Editing

The author opens edit on their own comment, changes the text, and saves. The same empty and length rules apply. A save that leaves the text unchanged writes nothing.

After a successful edit, the row keeps its place in the list. The text updates. The row shows that it was edited and the edit time.

The English label for that mark is "Edited". The Turkish label is "Düzenlendi".

## Not in this feature

- Deleting a comment.
- Editing another person's comment.
- Nested replies or a separate reply control.
- A photo on a comment.
- Changing the state, the device notes, or a status-change note through a comment.
- An audit row for a comment. The comment text is not stored in the audit log.
- Comments on the device page, the customer page, or the service list.
