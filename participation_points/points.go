{{/* participation_points / SLASH — personal points card (points.go).
     Trigger type: Slash Command.   Name: points.
     By DEFAULT the board is CENTRED on the caller: 5 ranks above them, their own
     row, and 4 below (10 rows). Callers near the top just see the top 10.

     OPTION (optional): add ONE Integer option named `start` (min value 1). When
     the caller supplies it, the board instead shows the 10 ranks beginning there
     (start:24 → ranks 24–33). In Discord this reads as `/points start:24`. Leaving
     the option off the command entirely is fine — you just lose the jump-to-rank
     ability; the centred default still works.

     Ephemeral: only the caller sees the reply. Shows the caller's lifetime
     (cumulative) points, their leaderboard rank, a 10-row slice of the board
     (centred on the caller by default, or a window from `start`), and which tier
     they're in. The caller's own rank is ALWAYS shown in the standing line,
     whether or not it falls inside the displayed window. Leave it un-restricted.

     Staff management (look up other members, add/remove points, set the tier
     base) lives in the SEPARATE `points` PREFIX command — see points_staff.go
     and setup.txt.

     DB BUDGET: dbGet(total) + dbRank + dbTopEntries + dbGet(base) = 4
     db_interactions and 2 db_multiple ops (dbRank + dbTopEntries). The free tier
     caps db_multiple at 2/run, so this sits exactly at the ceiling. Do NOT add
     another dbTop/dbBottom/dbRank/dbCount call here or it will error with "too
     many calls". */}}

{{- $color := 0xF4700F -}}
{{- $key := "pts-total" -}}
{{- $me := .User.ID -}}

{{/* ── TIER LADDER (config) ── ordered low → high, name + emoji only.
     Thresholds are COMPUTED, not listed: tier N needs base × N² points, a
     quadratic "quick early, slower later" RPG curve. The single `base` (points
     to reach tier 1) is the whole knob — staff set it live with `points tiers
     base <n>`; it's stored in the DB (key "pts-base") so both files stay in sync.
     Names/emojis below are placeholders — finalize with arcade/cas.
     ⚠ KEEP THIS $tiers BLOCK IDENTICAL to the one in points_staff.go. */}}
{{- $tiers := cslice
    (sdict "name" "Busy Bee"  "emoji" "🐝")
    (sdict "name" "Hive Hero" "emoji" "🍯")
    (sdict "name" "Tier 3"    "emoji" "🌸")
    (sdict "name" "Tier 4"    "emoji" "🌳")
    (sdict "name" "Tier 5"    "emoji" "✨")
    (sdict "name" "Tier 6"    "emoji" "🎟️")
    (sdict "name" "Tier 7"    "emoji" "💎")
    (sdict "name" "Tier 8"    "emoji" "🏆")
    (sdict "name" "Tier 9"    "emoji" "🌟")
    (sdict "name" "Tier 10"   "emoji" "👑") -}}
{{/* Caller's own lifetime total + leaderboard rank. dbRank is 1-based and returns
     0 when the caller has no entry. The exact key (no _ / % wildcards) makes the
     LIKE pattern an exact match, scoped to this guild. */}}
{{- $entry := dbGet $me $key -}}
{{- $pts := 0 -}}{{- if $entry -}}{{- $pts = toInt $entry.Value -}}{{- end -}}
{{- $rank := toInt (dbRank (sdict "pattern" $key) $me $key) -}}

{{/* Tier base — single source of truth is the DB key pts-base, seeded by the
     interval (points_badge_sweep.go). No local default. 0 = not seeded yet →
     tiers just don't show until the sweep (or `points tiers base`) sets it. */}}
{{- $be := dbGet 0 "pts-base" -}}
{{- $base := 0 -}}{{- if $be -}}{{- $base = toInt $be.Value -}}{{- end -}}

{{/* Current tier = highest index whose computed threshold (base × N²) ≤ my total.
     -1 = below tier 1 or base not seeded. Thresholds ascend, so last pass wins. */}}
{{- $tierIdx := -1 -}}
{{- if gt $base 0 -}}
  {{- range $i, $t := $tiers -}}
    {{- $n := add $i 1 -}}
    {{- if ge $pts (mult $base (mult $n $n)) -}}{{- $tierIdx = $i -}}{{- end -}}
  {{- end -}}
{{- end -}}

{{/* Board window. Default: CENTRED on the caller — start at (rank − 5) so they sit
     6th, with 5 ranks above and 4 below (10 rows). Clamped to rank 1, so a caller
     near the top just sees the top 10; a caller with 0 points (rank 0) also falls
     back to the top 10. The `start` option overrides this with an explicit window
     of 10 beginning at that rank (start:24 → offset 23 → ranks 24–33). dbTopEntries
     orders value DESC, id DESC — the same order dbRank uses — so the printed rank
     (start + row index) matches $rank and the caller lands at row 6. */}}
{{- $start := 0 -}}
{{- if .Options.start -}}{{- $start = toInt .Options.start -}}{{- else if gt $rank 0 -}}{{- $start = sub $rank 5 -}}{{- end -}}
{{- if lt $start 1 -}}{{- $start = 1 -}}{{- end -}}
{{- $top := dbTopEntries $key 10 (sub $start 1) -}}
{{- $board := "" -}}
{{- range $i, $t := $top -}}
  {{- $line := printf "**%d.** <@%d> — **%.0f**" (add $start $i) $t.UserID (toFloat $t.Value) -}}
  {{- if eq $t.UserID $me -}}{{- $line = printf "%s  ⬅️ you" $line -}}{{- end -}}
  {{- $board = printf "%s\n%s" $board $line -}}
{{- end -}}
{{- $boardTitle := "Top 10" -}}
{{- if gt $start 1 -}}
  {{- if gt (len $top) 0 -}}{{- $boardTitle = printf "Ranks %d–%d" $start (add $start (sub (len $top) 1)) -}}
  {{- else -}}{{- $boardTitle = printf "Ranks from %d" $start -}}{{- end -}}
{{- end -}}
{{- if eq $board "" -}}
  {{- if gt $start 1 -}}{{- $board = printf "\n*No one is ranked at #%d or beyond.*" $start -}}
  {{- else -}}{{- $board = "\n*No points earned yet — be the first, just post in chat!*" -}}{{- end -}}
{{- end -}}

{{/* Tier ladder — names/emojis only, NO thresholds (lucid: never show numbers),
     with the caller's current tier flagged. */}}
{{- $ladder := "" -}}
{{- range $i, $t := $tiers -}}
  {{- $mark := "" -}}{{- if eq $i $tierIdx -}}{{- $mark = "  ⬅️ you" -}}{{- end -}}
  {{- $ladder = printf "%s\n%s **%s**%s" $ladder $t.emoji $t.name $mark -}}
{{- end -}}

{{/* Caller's standing / badge line. */}}
{{- $standing := "" -}}
{{- if eq $rank 0 -}}
  {{- $standing = "You have **0** points so far — post in chat to start earning." -}}
{{- else if ge $tierIdx 0 -}}
  {{- $bt := index $tiers $tierIdx -}}
  {{- $standing = printf "**Your standing:** %s **%s** — #%d with **%d** point(s)." $bt.emoji $bt.name $rank $pts -}}
{{- else -}}
  {{- $standing = printf "**Your standing:** #%d with **%d** point(s) — not in a tier yet, keep earning!" $rank $pts -}}
{{- end -}}

{{/* Layout: your standing first, the board (top 10 or the `start` window) below
     it, then the tier ladder. */}}
{{- sendResponse nil (complexMessage "embed" (cembed "title" "Participation" "description" (printf "%s\n\n**%s**%s\n\n**Tiers**%s" $standing $boardTitle $board $ladder) "color" $color) "ephemeral" true) -}}
