# 🧩 Consolidated Advert Commands

## Why this exists

YAGPDB runs at most **3 message-triggered custom commands per message** on a
non-premium server (5 on premium), in order of command ID. Anything past the cap
is silently dropped. Our advert channels had grown to 5–6 message-triggered
commands each, so the advert command (length / cooldown / duplicate enforcement)
stopped firing entirely.

The fix: fold the per-channel content-checks **into** that channel's existing
advert command, which already runs on every post. Each advert channel ends up
with just **2** message-triggered commands:

```
merged advert command  +  sticky   =  2   (under the cap of 3)
```

## The merged command per channel

Each advert command becomes one linear flow:

```
1. LENGTH check        → over?     DM + delete + stop
2. COOLDOWN check      → active?   DM + delete + stop
3. DUP-IN-CHANNEL      → exists?   DM + delete + stop
   ── post is KEPT past here; it provably exists, so pings are race-free ──
4. write lastMsg_ / lastMsgTime_
5. ADVISORY checks (collect into ONE list):
     • link             (quick only)
     • header / image   (1x1 only)   • group-header (group only)
     • banned word      (all)
     • cross-channel duplicate (all)
6. if any advisory hits → ONE combined ping to #rule_infractions
7. schedule reaction_check  (quick only)
```

Because the advisory checks run **only when the post is kept**, the command knows
the post still exists, so there is no need for the old `alert_sender` delayed
re-check. Pings are sent directly.

## Subfolders

- **[group_advert](group_advert/)** — replaces the **Groups** command. (Pilot.)
- **1x1_advert** — replaces the **Normal / Long-Form** command. (After group is verified.)
- **quick_advert** — replaces the **Quick Channels** command. (Done last; most moving parts.)

## What gets DELETED after migration

These standalone commands are folded in and removed. Order matters — do not
delete the shared ones until every channel type is migrated:

| Command | Folded into | When to remove |
|---|---|---|
| `group_header_check` | group_advert | with group migration (group-only) |
| `header_alerts`, `no_images_in_1x1` | 1x1_advert | with 1x1 migration (1x1-only) |
| `link_alerts`, `quick_reactions/post_timer` | quick_advert | with quick migration (quick-only) |
| `banned_words` | all three | after ALL three migrated (shared) |
| `cross_channel_dupes` | all three | after ALL three migrated (shared) |
| `alert_sender` | (no longer used) | after ALL three migrated |

`reaction_check` (trigger type None) stays — quick_advert still calls it via execCC.
`autoremove_reactions` (reaction trigger) stays — it does not count against the cap.

## ⚠️ Duplicated logic

The banned-word list and the cross-channel-dup block are **duplicated in all three
merged commands** (YAGPDB has no include mechanism). When you change the banned
list, update it in **group_advert, 1x1_advert, and quick_advert**. Each file says
so at the top.
