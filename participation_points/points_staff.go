{{/* participation_points / STAFF — prefix command `points` (points_staff.go).
     Trigger type: Command.   Name: points.

     ┌────────────────────────────────────────────────────────────────────────┐
     │ RESTRICT THIS COMMAND TO YOUR STAFF ROLE(S) IN THE DASHBOARD.           │
     │ It has NO internal role gate (same pattern as                          │
     │ infraction_admin/infractions.go) — the dashboard Role restriction IS    │
     │ the gate. Without it, any member could award themselves points or edit  │
     │ tier thresholds.                                                        │
     └────────────────────────────────────────────────────────────────────────┘

     Members check their OWN points with the ephemeral /points slash command
     (points.go). This is the staff-side management tool.

     Points are CUMULATIVE and never reset — one lifetime total per member under
     the key `pts-total`. Tier thresholds are computed as base × N², where the
     single `base` lives under the global key `pts-base` and is what staff tune
     over time (see `tiers` below), not the scores themselves.

     Usage (prefix is your server's, e.g. -):
       points @member                look up a member
       points view @member           explicit form of the lookup
       points add @member 10         add points
       points remove @member 5       remove points (floored at 0)
       points adjust @member -5      signed adjust (+/-)
       points set @member 20         set an exact total (0 clears them)
       points balance                show YOUR OWN balance
       points balance 10             adjust your own balance by +10 (or -5, etc.)
       points list                   post the public board (top 10)
       points list 24                post the board starting at rank 24
       points tiers                  show the tier ladder + computed thresholds
       points tiers base 10          set the tier base (tier N needs base × N²)
       points tiers reset            revert the base to its default */}}

{{/* ──────────────── CONFIG ──────────────── */}}
{{- $logChannel := 0 -}}{{/* mod-log / audit channel ID for adjustments + tier edits, or 0 to disable */}}
{{- $color := 0xF4700F -}}
{{- $usage := "Usage: `points @member`, `points add|remove|adjust|set @member <n>`, `points balance [n]`, `points list [n]`, or `points tiers [base <n> | reset]`." -}}

{{/* ── TIER LADDER (config) ── ordered low → high, name + emoji only. Thresholds
     are COMPUTED: tier N needs base × N² points (quick early, slower later). The
     single `base` (points to reach tier 1) is stored in the DB (key "pts-base")
     and set live with `points tiers base <n>`. Names/emojis are placeholders —
     finalize with arcade/cas.
     ⚠ KEEP THIS $tiers BLOCK IDENTICAL to the one in points.go. */}}
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
{{/* ──────────────────────────────────────── */}}
{{/* Tier BADGE role IDs are NOT configured here — they live once in
     points_badge_sweep.go and are read from the DB key pts-tierroles inside the
     instant-reconcile block at the bottom. */}}

{{- $key := "pts-total" -}}

{{/* ── PARSE ── first token is a subcommand keyword or the looked-up member. */}}
{{- $args := .CmdArgs -}}
{{- $sub := "view" -}}{{- $userArg := "" -}}{{- $rest := "" -}}
{{- if eq (len $args) 0 -}}
{{ $usage }}
{{- return -}}
{{- end -}}
{{- $first := lower (index $args 0) -}}
{{- if in (cslice "view" "add" "remove" "adjust" "set") $first -}}
  {{- $sub = $first -}}
  {{- if gt (len $args) 1 -}}{{- $userArg = index $args 1 -}}{{- end -}}
  {{- if gt (len $args) 2 -}}{{- $rest = index $args 2 -}}{{- end -}}
{{- else if in (cslice "list" "leaders" "top" "leaderboard") $first -}}
  {{- $sub = "list" -}}
  {{- if gt (len $args) 1 -}}{{- $rest = index $args 1 -}}{{- end -}}
{{- else if eq $first "tiers" -}}
  {{- $sub = "tiers" -}}
{{- else if eq $first "balance" -}}
  {{- $sub = "balance" -}}
  {{- if gt (len $args) 1 -}}{{- $rest = index $args 1 -}}{{- end -}}
{{- else -}}
  {{- $sub = "view" -}}{{- $userArg = index $args 0 -}}
{{- end -}}

{{/* Tier base — single source of truth is the DB key pts-base, seeded by the
     interval (points_badge_sweep.go). No local default. 0 = not seeded yet. */}}
{{- $be := dbGet 0 "pts-base" -}}
{{- $base := 0 -}}{{- if $be -}}{{- $base = toInt $be.Value -}}{{- end -}}

{{/* Instant-badge trackers: any point-changing branch below sets these, and the
     reconcile block at the very bottom fixes that member's tier role right away. */}}
{{- $doReconcile := false -}}{{- $rcUser := 0 -}}{{- $rcTotal := 0 -}}

{{/* Resolve the target member for the subcommands that need one. userArg accepts
     a mention, a raw ID, or a name and returns the member (or nil). */}}
{{- $u := 0 -}}
{{- if $userArg -}}{{- $u = userArg $userArg -}}{{- end -}}

{{/* ─────────────── LOOKUP ─────────────── */}}
{{- if eq $sub "view" -}}
{{- if not $u -}}
Point to a member: `points @member`.
{{- else -}}
{{- $e := dbGet $u.ID $key -}}{{- $pts := 0 -}}{{- if $e -}}{{- $pts = toInt $e.Value -}}{{- end -}}
{{- $rank := toInt (dbRank (sdict "pattern" $key) $u.ID $key) -}}
{{/* highest tier whose computed threshold (base × N²) ≤ their total */}}
{{- $tierIdx := -1 -}}
{{- if gt $base 0 -}}{{- range $i, $t := $tiers -}}{{- $n := add $i 1 -}}{{- if ge $pts (mult $base (mult $n $n)) -}}{{- $tierIdx = $i -}}{{- end -}}{{- end -}}{{- end -}}
{{- $badge := "" -}}{{- if ge $tierIdx 0 -}}{{- $bt := index $tiers $tierIdx -}}{{- $badge = printf " — %s %s" $bt.emoji $bt.name -}}{{- end -}}
{{- if eq $rank 0 -}}
<@{{ $u.ID }}> has **0** points.
{{- else -}}
<@{{ $u.ID }}> is **#{{ $rank }}** with **{{ $pts }}** point(s){{ $badge }}.
{{- end -}}
{{- end -}}

{{/* ─────────── ADJUST / ADD / REMOVE / SET ─────────── */}}
{{- else if in (cslice "add" "remove" "adjust" "set") $sub -}}
{{- if not $u -}}
Point to a member: `points {{ $sub }} @member <amount>`.
{{- else if not (reFind `^[+-]?\d+$` $rest) -}}
Give a whole-number amount: `points {{ $sub }} @member 10`.
{{- else -}}
{{- $amount := toInt $rest -}}
{{- if eq $sub "set" -}}
{{/* absolute set; 0 or less clears the entry */}}
{{- $newTotal := 0 -}}
{{- if le $amount 0 -}}{{- dbDel $u.ID $key -}}{{- else -}}{{- dbSet $u.ID $key $amount -}}{{- $newTotal = $amount -}}{{- end -}}
{{- if $logChannel -}}{{- sendMessage $logChannel (cembed "title" "Points set" "description" (printf "Set <@%d> to **%d** — by <@%d>" $u.ID $newTotal .User.ID) "color" $color) -}}{{- end -}}
Set <@{{ $u.ID }}> to **{{ $newTotal }}** point(s).
{{- $doReconcile = true -}}{{- $rcUser = $u.ID -}}{{- $rcTotal = $newTotal -}}
{{- else if eq $amount 0 -}}
Amount must be non-zero.
{{- else -}}
{{- $mag := $amount -}}{{- if lt $amount 0 -}}{{- $mag = mult $amount -1 -}}{{- end -}}
{{- $delta := $amount -}}
{{- if eq $sub "add" -}}{{- $delta = $mag -}}{{- end -}}
{{- if eq $sub "remove" -}}{{- $delta = mult $mag -1 -}}{{- end -}}
{{- $newTotal := toInt (dbIncr $u.ID $key $delta) -}}
{{- $before := sub $newTotal $delta -}}
{{- $clamped := false -}}
{{- if le $newTotal 0 -}}{{- dbDel $u.ID $key -}}{{- if lt $newTotal 0 -}}{{- $clamped = true -}}{{- end -}}{{- $newTotal = 0 -}}{{- end -}}
{{- $verb := "Added" -}}{{- $prep := "to" -}}{{- $shown := $delta -}}
{{- if lt $delta 0 -}}{{- $verb = "Removed" -}}{{- $prep = "from" -}}{{- $shown = mult $delta -1 -}}{{- end -}}
{{- $note := "" -}}{{- if $clamped -}}{{- $note = " (floored at 0)" -}}{{- end -}}
{{- if $logChannel -}}{{- sendMessage $logChannel (cembed "title" (printf "Points %s" $verb) "description" (printf "%s **%d** %s <@%d> — %d → **%d**%s — by <@%d>" $verb $shown $prep $u.ID $before $newTotal $note .User.ID) "color" $color) -}}{{- end -}}
{{ $verb }} **{{ $shown }}** point(s) {{ $prep }} <@{{ $u.ID }}> — now **{{ $newTotal }}**{{ $note }}.
{{- $doReconcile = true -}}{{- $rcUser = $u.ID -}}{{- $rcTotal = $newTotal -}}
{{- end -}}
{{- end -}}

{{/* ─────────────── LIST (all-time board) ─────────────── */}}
{{- else if eq $sub "list" -}}
{{/* optional start rank: `points list 24` shows ranks 24–33 (offset 23). */}}
{{- $start := 1 -}}{{- if reFind `^\d+$` $rest -}}{{- $start = toInt $rest -}}{{- end -}}
{{- if lt $start 1 -}}{{- $start = 1 -}}{{- end -}}
{{- $top := dbTopEntries $key 10 (sub $start 1) -}}
{{- $out := "" -}}
{{- range $i, $e := $top -}}{{- $out = printf "%s\n**%d.** <@%d> — **%.0f**" $out (add $start $i) $e.UserID (toFloat $e.Value) -}}{{- end -}}
{{- $ttl := "Participation — top 10" -}}
{{- if gt $start 1 -}}{{- if gt (len $top) 0 -}}{{- $ttl = printf "Participation — ranks %d–%d" $start (add $start (sub (len $top) 1)) -}}{{- else -}}{{- $ttl = printf "Participation — ranks from %d" $start -}}{{- end -}}{{- end -}}
{{- if eq $out "" -}}{{- if gt $start 1 -}}{{- $out = printf "*No one is ranked at #%d or beyond.*" $start -}}{{- else -}}{{- $out = "*No points recorded yet.*" -}}{{- end -}}{{- end -}}
{{/* Post as an embed so the mentions render as names WITHOUT pinging anyone. */}}
{{- sendMessage .Channel.ID (cembed "title" $ttl "description" $out "color" $color) -}}

{{/* ─────────────── TIERS ─────────────── */}}
{{- else if eq $sub "tiers" -}}
{{- $action := "" -}}{{- if gt (len $args) 1 -}}{{- $action = lower (index $args 1) -}}{{- end -}}
{{- if eq $action "reset" -}}
{{/* clear the DB key; the badge sweep re-seeds its $baseDefault on the next run
     (the default lives ONLY there, so this command doesn't restate a number). */}}
{{- dbDel 0 "pts-base" -}}
{{- if $logChannel -}}{{- sendMessage $logChannel (cembed "title" "Tier base reset" "description" (printf "Base cleared — the badge sweep re-seeds the default on its next run — by <@%d>" .User.ID) "color" $color) -}}{{- end -}}
Tier base cleared. The badge sweep re-seeds the default on its next hourly run — or run `points tiers base <n>` to set it now.
{{- else if eq $action "base" -}}
{{- $arg := "" -}}{{- if gt (len $args) 2 -}}{{- $arg = index $args 2 -}}{{- end -}}
{{- if not (reFind `^\d+$` $arg) -}}
Give a whole number: `points tiers base 10` (tier N then needs base × N²).
{{- else if lt (toInt $arg) 1 -}}
Base must be at least **1**.
{{- else -}}
{{- $base = toInt $arg -}}
{{- dbSet 0 "pts-base" $base -}}
{{- if $logChannel -}}{{- sendMessage $logChannel (cembed "title" "Tier base updated" "description" (printf "Base set to **%d** — by <@%d>" $base .User.ID) "color" $color) -}}{{- end -}}
{{- $out := "" -}}
{{- range $i, $t := $tiers -}}{{- $n := add $i 1 -}}{{- $out = printf "%s\n%s **%s** — %d pts" $out $t.emoji $t.name (mult $base (mult $n $n)) -}}{{- end -}}
{{- sendMessage .Channel.ID (cembed "title" (printf "Tier thresholds updated — base %d (tier N = base × N²)" $base) "description" $out "color" $color) -}}
{{- end -}}
{{- else -}}
{{/* no action → show the ladder with computed thresholds (staff see the numbers). */}}
{{- if le $base 0 -}}
Tier base isn't seeded yet — the badge sweep sets it on its next run, or run `points tiers base <n>` to set it now.
{{- else -}}
{{- $out := "" -}}
{{- range $i, $t := $tiers -}}{{- $n := add $i 1 -}}{{- $out = printf "%s\n%s **%s** — %d pts" $out $t.emoji $t.name (mult $base (mult $n $n)) -}}{{- end -}}
{{- sendMessage .Channel.ID (cembed "title" (printf "Tier ladder — base %d (tier N = base × N²)" $base) "description" $out "color" $color) -}}
{{- end -}}
{{- end -}}

{{/* ─────────────── BALANCE (self) ─────────────── */}}
{{/* `points balance` shows the CALLER's own point total; `points balance <n>`
     adjusts it by a signed whole number (e.g. 10 or -5), floored at 0. Handy for
     testing tiers/badges on yourself. The badge reflects on the next sweep. */}}
{{- else if eq $sub "balance" -}}
{{- $me := .User.ID -}}
{{- if eq $rest "" -}}
{{- $e := dbGet $me $key -}}{{- $bal := 0 -}}{{- if $e -}}{{- $bal = toInt $e.Value -}}{{- end -}}
Your balance: **{{ $bal }}** point(s).
{{- else if not (reFind `^[+-]?\d+$` $rest) -}}
Give a whole-number amount: `points balance 10` or `points balance -5`.
{{- else -}}
{{- $delta := toInt $rest -}}
{{- $newTotal := toInt (dbIncr $me $key $delta) -}}
{{- if le $newTotal 0 -}}{{- dbDel $me $key -}}{{- $newTotal = 0 -}}{{- end -}}
Your balance is now **{{ $newTotal }}** point(s).
{{- $doReconcile = true -}}{{- $rcUser = $me -}}{{- $rcTotal = $newTotal -}}
{{- end -}}

{{- else -}}
{{ $usage }}
{{- end -}}

{{/* ── Instant badge reconcile. A point change above set $doReconcile; fix that
     member's tier role NOW (add the tier they're in, drop any other tier role)
     instead of waiting for the hourly sweep. No-op when the base isn't seeded or
     $tierRoles is all "". Same string-compare reconcile the sweep uses; no delay
     so the badge changes on the spot. */}}
{{- if and $doReconcile (gt $base 0) -}}
  {{/* tier role IDs from the single source (points_badge_sweep.go → pts-tierroles) */}}
  {{- $trEntry := dbGet 0 "pts-tierroles" -}}{{- $tierRoles := cslice -}}{{- if $trEntry -}}{{- $tierRoles = $trEntry.Value -}}{{- end -}}
  {{- $tt := -1 -}}
  {{- range $i, $r := $tierRoles -}}{{- $n := add $i 1 -}}{{- if ge $rcTotal (mult $base (mult $n $n)) -}}{{- $tt = $i -}}{{- end -}}{{- end -}}
  {{- $trole := "" -}}{{- if ge $tt 0 -}}{{- $trole = index $tierRoles $tt -}}{{- end -}}
  {{- $rm := getMember $rcUser -}}
  {{- if $rm -}}
    {{- $has := false -}}{{- $wrong := cslice -}}
    {{- range $rm.Roles -}}{{- $rid := str . -}}{{- range $tr := $tierRoles -}}{{- if and $tr (eq $tr $rid) -}}{{- if eq $rid $trole -}}{{- $has = true -}}{{- else -}}{{- $wrong = $wrong.Append $rid -}}{{- end -}}{{- end -}}{{- end -}}{{- end -}}
    {{- range $wr := $wrong -}}{{- takeRoleID $rcUser $wr -}}{{- end -}}
    {{- if and (ne $trole "") (not $has) -}}{{- giveRoleID $rcUser $trole -}}{{- end -}}
  {{- end -}}
{{- end -}}
