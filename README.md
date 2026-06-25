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

### 🩹 Member intros — [member_intros_fix](member_intros_fix/)

Validates posts in the member-introductions channel: enforces the character
limit and allows only one intro per member, DMing the author and removing the
post when either rule is broken.
