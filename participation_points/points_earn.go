{{/* participation_points / EARNER — banks cumulative participation points.

     Trigger: Regex  ([\s\S]*)   (fires on EVERY message in the channels this
     command is restricted to — including attachment-only posts with no caption,
     the same way the advert commands catch a bare image. `*` matches the empty
     string, so a caption-less upload still triggers; a `.+`/`^.+$` pattern would
     NOT and would silently miss those posts.)
     Channel Restrictions: your participation / event channels ONLY. See setup.txt
     — do NOT run this in the advert or #rule_infractions channels, which are
     already at 2 of YAGPDB's 3 message-trigger slots.

     Silent: it never posts anything, it just banks points. ~3 DB ops in the
     common path (cooldown get + incr + cooldown set); +2 more (pts-base +
     pts-tierroles) only on a scoring post once the tier base is seeded. Well
     inside the free-tier cap of 10.

     CUMULATIVE, NO RESET: points accumulate forever under one lifetime key
     (pts-total). There is no monthly bucket and no reset job — a member's total
     only ever grows. Tier thresholds (see points_staff.go `tiers`) are what get
     recalibrated over time, not the scores. */}}

{{/* ──────────────── CONFIG ──────────────── */}}
{{/* ▼▼ Points per action. An attachment beats text when a post has both. ▼▼ */}}
{{ $attachmentPoints := 5 }}
{{ $textPoints       := 1 }}
{{/* ▼▼ One earn per member per this many seconds (anti-farm — stops someone
       spamming one-word messages for points). ~10s feels almost per-message in
       normal chat while still blocking copy-paste floods. ▼▼ */}}
{{ $cooldownSecs := 10 }}
{{/* ▼▼ A text post must have at least this many characters (after trimming
       whitespace) to count. ▼▼ */}}
{{ $minChars := 2 }}
{{/* ▼▼ Staff role ID(s) — members with any of these earn NOTHING, so staff never
       land on the board or in a tier (lucid: "filter staff from the list";
       cas: "staff being exempt"). Keep this identical to $staffRoles in
       points_bonus.go. As strings. ▼▼ */}}
{{ $staffRoles := cslice "300831005621878784" "322845008409395200" "479484736188973087" "1371346095380238376" "1376679985603022899" }}
{{/* ─────────────────────────────────────── */}}
{{/* Tier BADGE role IDs are NOT configured here — they live once in
     points_badge_sweep.go and are read from the DB key pts-tierroles below. */}}

{{/* Bots and system messages never reach a CC (YAGPDB ignores them), so every
     run here is a real human post. */}}

{{/* Staff earn nothing — bail before touching the DB. In a regex message trigger
     hasRoleID checks the poster. */}}
{{ range $sr := $staffRoles }}{{ if hasRoleID (toInt64 $sr) }}{{ return }}{{ end }}{{ end }}

{{/* Anti-farm gate: if this member earned within the cooldown window, stop. */}}
{{ if (dbGet .User.ID "ptsCD") }}{{ return }}{{ end }}

{{/* Decide the award: ANY attachment (image, video, file, …) scores
     $attachmentPoints; otherwise a non-trivial text post scores $textPoints. An
     attachment wins when a post has both. */}}
{{ $amt := 0 }}
{{ if gt (len .Message.Attachments) 0 }}{{ $amt = $attachmentPoints }}{{ end }}
{{- if eq $amt 0 }}
  {{- if ge (len (trimSpace .Message.Content)) $minChars }}{{ $amt = $textPoints }}{{ end }}
{{- end }}
{{ if eq $amt 0 }}{{ return }}{{ end }}

{{/* Bank it onto the lifetime total and arm the cooldown. */}}
{{/* capture the return of dbIncr — a bare {{ dbIncr }} would PRINT the new
     total into the channel; assigning it keeps the earner silent. */}}
{{ $newTotal := toInt (dbIncr .User.ID "pts-total" $amt) }}
{{ dbSetExpire .User.ID "ptsCD" 1 $cooldownSecs }}

{{/* ── Instant tier-up. If this earn crossed a threshold (tier N = base × N²),
     grant the new badge role now and drop the old one. We know the old + new
     totals from the math, so no getMember is needed. Base is read from pts-base
     (single source, seeded by the badge sweep); if it isn't seeded yet ($base 0)
     we skip and let the sweep handle it. The sweep is also the reconciler for
     demotions / rebalances / anyone this misses. */}}
{{ $be := dbGet 0 "pts-base" }}{{ $base := 0 }}{{ if $be }}{{ $base = toInt $be.Value }}{{ end }}
{{ if gt $base 0 }}
  {{/* tier role IDs from the single source (points_badge_sweep.go → pts-tierroles) */}}
  {{ $trEntry := dbGet 0 "pts-tierroles" }}{{ $tierRoles := cslice }}{{ if $trEntry }}{{ $tierRoles = $trEntry.Value }}{{ end }}
  {{ $oldTotal := sub $newTotal $amt }}
  {{ $oldTier := -1 }}{{ $newTier := -1 }}
  {{ range $i, $r := $tierRoles }}
    {{ $n := add $i 1 }}{{ $need := mult $base (mult $n $n) }}
    {{ if ge $oldTotal $need }}{{ $oldTier = $i }}{{ end }}
    {{ if ge $newTotal $need }}{{ $newTier = $i }}{{ end }}
  {{ end }}
  {{ if gt $newTier $oldTier }}
    {{ $nr := index $tierRoles $newTier }}
    {{ if $nr }}{{ addRoleID $nr }}{{ end }}
    {{ if ge $oldTier 0 }}{{ $or := index $tierRoles $oldTier }}{{ if and $or (ne $or $nr) }}{{ removeRoleID $or }}{{ end }}{{ end }}
  {{ end }}
{{ end }}
