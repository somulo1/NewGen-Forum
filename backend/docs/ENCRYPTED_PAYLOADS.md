# What is inside the encrypted envelopes

The server stores these as opaque bytes. The shapes below are what the seeder (acting as demo clients) writes, matching the
frontend's existing data, so the frontend can render them once it decrypts. All are UTF-8 JSON, sealed with
`e2ee.Device.Seal` using the container's conversation key and the context (org, conversation, sender account, id, epoch).

| Container / type | Payload |
|---|---|
| dm / forum, `text` | `{"text"}` ; replies set `reply_to` |
| forum `poll` / `vote` / `poll_lock` | `{"question","options":[{"id","label"}]}` / `{"option"}` (reply_to the poll) / `{"locked":true}` |
| wall `post` | `{"text","department","photo":{"label","bg"}}` ; comments: type `comment`; likes: type `reaction` `{"kind":"like"|"love"}`; `share` |
| wall `moment` | `{"caption","bg","media":{"url","isImg","name"}}` |
| wall `announcement` / `memo` | `{"title","body","to","type"}` (posted when approvals finish) |
| ticket first message | `{"title","desc","category"}` ; later messages `{"text"}` |
| note vault records | `notebook` `{"name","color"}` ; `note` `{"title","body","updated"}` with plain meta `{book,pinned}` |
| project records | `task`, `milestone`, `sprint`, `risk`, `activity`, `standup`, `comment` (plain meta only for column/due/assignee/sprint) |
| event `entry` | `{"title","forum","people","by"}` with plain meta `{date,start,end,type,repeat}` |
| folder file | `{"name","kind","size_label","by","when","imgBg","file_key"}`; the bytes are AES-GCM ciphertext in the blob store |
| secure-share `key_envelope` | file key + metadata sealed under the link key that lives only in the URL `#fragment` |

client can sign in as any demo person. Never use the demo seed outside development; the seeder refuses `HUB_ENV=prod`.
