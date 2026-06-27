# 🐉 RPC YAGPDB Bot

This project holds the custom-command code running on the RPC YAGPDB bot. It
automates common staff tasks — mostly advert-channel enforcement — so staff
don't have to hand-check every post.

## 🗺️ Layout

Each folder holds the Go template that gets pasted into the YAGPDB dashboard, a
`setup.txt` with install instructions, and screenshots of the dashboard
settings.

## ✨ Features

### 📢 Advert enforcement — [consolidated_advert_commands](consolidated_advert_commands/)

One command per advert channel type (**quick**, **one-on-one**, **group**), each
running the whole check pipeline on every post:

- **Hard rules — delete the post + DM the author:** over the length limit,
  posting again before the cooldown expires, or already having an advert in that
  channel.
- **Advisory rules — keep the post + one ping in #rule_infractions:** links
  (quick), images/attachments (quick & 1x1), Discord headers (quick & 1x1; group
  allows a single short header line), banned words (shown in ||spoilers||), and
  the same advert copy-pasted across more than one channel.

Folding every check into a single per-channel command keeps each advert channel
under YAGPDB's cap of 3 message-triggered commands per post.

### ✅ Quick-advert reaction requirement — [quick_reactions](quick_reactions/)

Quick adverts must carry at least three approved roleplay reaction tags. A few
minutes after a post, `reaction_check` re-reads it and, if it's short on tags,
pings the author in #rule_infractions to add them.

### 🧹 Advert reaction cleanup — [autoremove_reactions](autoremove_reactions/)

On the advert channels, reactions added by anyone who isn't the original poster
or a staff member are removed automatically, and staff-only emojis are protected
from non-staff use.

### 📌 #rule_infractions sticky — [infractions_sticky](infractions_sticky/)

Keeps the sticky pinned to the bottom of #rule_infractions even though the bot's
own infraction pings would otherwise bury it — a plain sticky's `.*` trigger
never fires on the bot's own messages, so the advert commands re-pin it via
`execCC` right after they post a ping.

### 🚩 Repeat-infraction tracking — [infractions_sticky](infractions_sticky/)

Counts each member's infractions over a rolling **6-month** window — stored as a
single per-user list of timestamps (`infractionDates`), pruned on every write, so
there's one DB row per person rather than one per infraction. On the **3rd**
infraction the notice gains a warning line; on the **4th** the member is suspended
from posting adverts for **14 days** (a self-expiring `advertBan` flag the advert
commands check before accepting a post), a heads-up goes to bot-spam so staff know
a ban landed, and the infraction history is wiped so the slate is clean once the
ban is served. Bot-issued pings are counted inline by the advert commands and
`reaction_check`; manually typed staff infractions are counted by the sticky's
`.*` branch — disjoint message sets, so nothing is double-counted.

### 🛠️ Infraction admin command — [infraction_admin](infraction_admin/)

One staff **slash command**, `/infractions`, with an action menu (**view** / **clear**
/ **set**) and a native member-picker (a `user` slash-command option, so you
autocomplete the right person instead of pasting IDs): view shows the count over the
rolling 6 months plus ban status, clear wipes the history and lifts any active ban,
and set writes a count (1–99). It reads and writes the same `infractionDates` /
`advertBan` database the advert commands use. Folding the three actions into one
command keeps it to a single slash-command slot (YAGPDB allows 3 on free servers).
Restrict to staff.

### 🔁 Infraction re-check — [infraction_recheck](infraction_recheck/)

After a post is flagged with the :staffpending: reaction, re-checks it at ~10,
45, 90, and 480 minutes. The first time the post comes back clean, :staffpending:
is removed from the advert and the #rule_infractions ping is marked
:staffapproved:; if it's still broken at the 8h mark, the advert is deleted and
its ping is marked :staffapproved: to close it out. The chain is kicked off by
the sticky and walks itself forward with `scheduleUniqueCC`, so it stays within
YAGPDB's one-`execCC`-per-run limit.

### 🩹 Member intros — [member_intros_fix](member_intros_fix/)

Validates posts in the member-introductions channel: enforces the character
limit and allows only one intro per member, DMing the author and removing the
post when either rule is broken.
