# 🐉 RPC YAGPDB Bot

This project holds the custom-command code running on the RPC YAGPDB bot. It
automates common staff tasks — mostly advert-channel enforcement — so staff
don't have to hand-check every post.

## 🗺️ Layout

Each folder holds the Go template that gets pasted into the YAGPDB dashboard, a
`setup.txt` with install instructions, and screenshots of the dashboard
settings.

The three advert commands ship in two forms: the readable original
(`quick_advert.go`) and a minified single-line version (`quick_advert.min.go`).
**The minified file is the one you paste into YAGPDB** — it's the same logic with
the comments and whitespace stripped so it fits the dashboard's code limit. Edit
the readable `.go`, then re-minify (see [Making changes](#-making-changes)).

## ✨ Features

### 📢 Advert enforcement — [consolidated_advert_commands](consolidated_advert_commands/)

One command per advert channel type (**quick**, **one-on-one**, **group**), each
running the whole check pipeline on every post:

- **Hard rules — delete the post + DM the author:** over the length limit,
  posting again before the cooldown expires, or already having an advert in that
  channel.
- **Advisory rules — keep the post + one ping in #rule_infractions:** links
  (quick), images/attachments (quick & 1x1), Discord headers (quick & 1x1; group
  allows a single short header line), banned words, and
  the same advert copy-pasted across more than one channel.

Folding every check into a single per-channel command keeps each advert channel
under YAGPDB's cap of 3 message-triggered commands per post.

### ✅ Quick-advert reaction requirement — [quick_reactions](quick_reactions/)

Quick adverts must carry at least three approved roleplay reaction tags. A few
minutes after a post, `reaction_check` re-reads it and, if it's short on tags,
pings the author in #rule_infractions to add them. `reaction_check` is a
trigger-type-**None** command (it never fires on its own), so it costs no
message-trigger slot.

The +5-minute timer that arms it is folded into the **quick-channel sticky**
([quick_sticky](quick_sticky/)) rather than a standalone "Post Timer" command.
That drops the quick channels from 3 message-triggered commands (advert cmd +
sticky + timer) to **2** (advert cmd + sticky), so every advert channel type now
sits at 2 of YAGPDB's 3 slots.

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

### 🔤 Nickname normalizer — [nickname_normalizer](nickname_normalizer/)

Runs on every message in the Age Verification channel and tidies the poster's
server nickname: it maps "fancy font" Unicode (bold/italic/script/fraktur/
double-struck/sans/monospace, fullwidth, circled, small-caps, etc.) back to plain
ASCII (`𝕎𝕠𝕠𝕞𝕒` → Wooma, `𝓓𝑂𝑉𝐸𝑇𝑇𝐸` → Dovette), strips emojis and disallowed
punctuation (only `!@#$%&+-` are kept, never at the ends), collapses repeated
spaces, forces alphanumeric first/last characters, and title-cases each word.
The cleaned name is written back with `editNickname` (which Discord applies on the
member's next message); an unreadable, all-symbol nickname instead earns a one-per-6h
DM asking for a readable name.

### 📪 DMs-closed advert enforcement — [get_roles_dms_closed](get_roles_dms_closed/)

A reaction command on the #get_roles status message. When a member switches
themselves to **not looking | dms closed** (the `:mailbox_closed:` reaction)
while they still have an advert live, the bot DMs them a **30-minute** warning;
if their status is still *not looking* at the recheck, every advert of theirs
that's still posted is deleted. Role assignment stays on native reaction roles —
this only enforces the closed case, and switching to **neutral | advert only**
or **looking | dms open** within the window is how a member keeps their ad(s).
"Still has an advert" reuses the advert commands' `lastMsg_<channel>` records
(resolved with `getMessage`), and the recheck is the command scheduling itself
forward via `scheduleUniqueCC`.

### 🩹 Member intros — [member_intros_fix](member_intros_fix/)

Validates posts in the member-introductions channel: enforces the character
limit and allows only one intro per member, DMing the author and removing the
post when either rule is broken.

## 🧭 Design

![Advert-post execution flow](image.png)

### Two of the three regex slots

YAGPDB (free tier) runs at most **3 message-triggered — "regex" — custom
commands per message**; anything past the cap is silently dropped. That cap is
the whole reason the advert checks were consolidated. Each advert channel now
spends just **2 of its 3 slots**:

1. the **merged advert command** (all the length / cooldown / duplicate /
   advisory checks in one), and
2. the channel's own **sticky** (which also arms the quick reaction timer).

That leaves **one slot of headroom**. Everything that happens *after* a post —
the +5-minute reaction check and the infraction re-check chain — runs on
trigger-type-**None** commands fired by `scheduleUniqueCC`, which don't count
against the 3. So no matter how much follow-up logic we add, an advert channel
never uses more than 2 of its message-trigger slots.

### How to read the flowchart

The diagram above traces one advert post from the top. In plain terms:

- **Someone posts in an advert channel.** Two things fire on that post — the
  advert command (slot 1) and the sticky (slot 2).
- **The sticky** just re-pins the channel's rules reminder to the bottom so it
  never scrolls away. Done.
- **The advert command checks the hard rules first:** is the post too long, is
  the author still on cooldown, do they already have an ad in this channel, or
  are they advert-banned? If **any** hard rule fails, the bot **DMs the author
  the reason and deletes the post** — end of story.
- **If it passes the hard rules, the post stays** and the bot records it. Now it
  looks for **advisory** problems: links, images, headers, banned words, or the
  same ad copy-pasted across channels. If there are **none**, nothing else
  happens — the post is fine. If there **is** one, the bot posts **one ping in
  #rule_infractions**, adds a ⏳ `:staffpending:` reaction to the ad, and counts
  the infraction (3rd = a warning line, 4th = a 14-day advert ban).
- **Quick channels get one extra check** (armed by the sticky, so it's free):
  about **5 minutes** later the bot re-reads the post and, if it has fewer than
  **3 approved reaction tags**, pings the author to add them.
- **Every ping kicks off a re-check chain.** Ten minutes later the bot looks at
  the post again. If the author **fixed it**, the bot clears the `:staffpending:`
  flag and marks the ping approved. If the **post is gone**, it stops. If it's
  **still broken**, it checks again at +35 min, +45 min, and +390 min — and if
  it's *still* broken at the **8-hour** mark, the ad is **deleted**.

## 🔧 Making changes

You edit the readable original (e.g. `1x1_advert.go`), then **minify it and
replace the existing minified file** (`1x1_advert.min.go`) — the minified version
is what actually gets pasted into YAGPDB.

Minifying a Go template by hand is error-prone, so use one of:

- **A minifier that understands Go template (`text/template`) syntax** — a plain
  JS/HTML minifier will mangle `{{ ... }}` blocks and break the command.
- **AI.** If you go this route, use a **premium model on high thinking** — Claude
  Opus or better. Minification has to be *logically exact*, so this is not a job
  for a small/fast model.

If using AI, this two-prompt sequence is highly accurate:

> In the same fashion as the advert commands are minified, I want to take the
> current code of the originals (e.g. `1x1_advert.go`) and minify them, replacing
> the existing minified files, being careful to minify accurately and be
> logically exact.

then, to verify:

> Compare the minified versions of the advert commands with the originals and
> ensure they are logically consistent.

Running both is belt-and-suspenders — the second (verification) pass is arguably
overkill, but it's cheap insurance that the minified command behaves identically
to the original.

## 🗑️ Uninstalling

To revert the server to how it worked before this project, go to the YAGPDB
dashboard → **Custom Commands** and:

1. **Disable (or delete) every command this project added** — one per feature
   above:
   - the three consolidated advert commands (`quick_advert`, `1x1_advert`,
     `group_advert`)
   - `reaction_check` and the quick-channel sticky's reaction timer
     ([quick_sticky](quick_sticky/))
   - `autoremove_reactions`
   - the `#rule_infractions` sticky's infraction-counting / re-stick additions
     ([infractions_sticky](infractions_sticky/))
   - `/infractions` ([infraction_admin](infraction_admin/))
   - `infraction_recheck`
   - the nickname normalizer (and `nametest`)
   - `get_roles_dms_closed`
   - the member-intros fix

2. **Re-enable the three original advert-moderator commands.** These are the ones
   the consolidated commands replaced, and they sit at the **top of the command
   list** (lowest command IDs):
   - **Quick Channels**
   - **Normal / Long-Form Channels**
   - **Groups**

Once those three are back on and this project's commands are off, advert
enforcement behaves exactly as it did originally.
