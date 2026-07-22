{{/* participation_points / BONUS — staff-reaction point award (points_bonus.go).

     Trigger type: Reaction (a reaction is added/removed).
     Channel Restrictions: your participation / event channels ONLY (same
     whitelist as points_earn.go). A Reaction trigger fires on a REACTION event,
     not a message, so it does NOT count against YAGPDB's "3 custom commands per
     message" budget that the earner/advert/infraction regex commands share.

     WHAT IT DOES: when a staff member reacts to a post with the award emoji, the
     post's author gets $bonusPoints added to their lifetime participation total
     (pts-total) — once per post, ever. Reacting again, removing + re-adding, or a
     second staff member reacting does nothing extra. If the bonus pushes the
     author across a tier threshold, their badge role is updated IMMEDIATELY
     (giveRoleID/takeRoleID on the author) — no wait for the badge sweep.

     Guards: add-only (ignores reaction removals), emoji must match, reactor must
     be staff, author can't be a bot, staff can't award their OWN post, and a
     STAFF author earns nothing (staff are exempt from the board/tiers).

     DB ops: dbGet (dupe flag) + dbSetExpire (set flag) + dbIncr (award) +
     dbGet (pts-base) + dbGet (pts-tierroles) = 5 db_interactions, one getMember,
     and (only on a tier-up) one give + one take role call — well inside the cap. */}}

{{/* ──────────────── CONFIG ──────────────── */}}
{{/* ▼▼ Points a single staff award grants. ▼▼ */}}
{{ $bonusPoints := 10 }}

{{/* ▼▼ The award emoji as name:id — SAME format as :staffapproved:/:staffpending:.
       Leave "" until you've uploaded the emoji; the command is a silent no-op
       until it's set. Type \:youremoji: in Discord to read its name:id.
       Matched by ID, so renaming the emoji later won't break it.
       e.g. "pointsaward:1442331141771366513"  (a bare "1442331141771366513"
       also works). ▼▼ */}}
{{ $bonusEmoji := "" }}

{{/* ▼▼ Staff role ID(s) as strings. Dual purpose: only these roles may GRANT a
       bonus, and a member with any of these roles can never RECEIVE one (staff
       are exempt from the board/tiers). Keep identical to $staffRoles in
       points_earn.go. ▼▼ */}}
{{ $staffRoles := cslice "300831005621878784" "322845008409395200" "479484736188973087" "1371346095380238376" "1376679985603022899" }}

{{/* ▼▼ Mod-log / audit channel ID for a record of each grant, or 0 to disable. ▼▼ */}}
{{ $logChannel := 0 }}

{{/* ▼▼ true → also post a short public "🎉 +N bonus points" note (and ping the
       member) in the channel where the post lives. false → award silently. ▼▼ */}}
{{ $announce := false }}
{{ $color := 0xF4700F }}
{{/* ─────────────────────────────────────── */}}
{{/* Tier BADGE role IDs are NOT configured here — they live once in
     points_badge_sweep.go and are read from the DB key pts-tierroles below. */}}

{{/* Add-only, and only for the configured emoji. Silent no-op until $bonusEmoji
     is set. Match on the emoji ID (the part after the colon). */}}
{{ if not .ReactionAdded }}{{ return }}{{ end }}
{{ if not $bonusEmoji }}{{ return }}{{ end }}
{{ $wantID := $bonusEmoji }}
{{ $ep := split $bonusEmoji ":" }}{{ if eq (len $ep) 2 }}{{ $wantID = index $ep 1 }}{{ end }}
{{ if ne (str .Reaction.Emoji.ID) $wantID }}{{ return }}{{ end }}

{{/* Reactor must be staff. In a Reaction trigger the triggering member is the
     one who reacted, so hasRoleID checks the reactor. */}}
{{ $isStaff := false }}
{{ range $sr := $staffRoles }}{{ if hasRoleID (toInt64 $sr) }}{{ $isStaff = true }}{{ end }}{{ end }}
{{ if not $isStaff }}{{ return }}{{ end }}

{{/* The rewarded post + its author. */}}
{{ $post := .ReactionMessage }}
{{ if not $post }}{{ return }}{{ end }}
{{ $author := $post.Author }}
{{ if not $author }}{{ return }}{{ end }}
{{ if $author.Bot }}{{ return }}{{ end }}
{{/* No self-awards. */}}
{{ if eq $author.ID .Reaction.UserID }}{{ return }}{{ end }}

{{/* Staff author earns nothing — same exemption the earner applies to posters.
     .ReactionMessage.Author is a user with no roles, so pull the member and
     compare role IDs as strings (same idiom as newbie_gate/get_roles — avoids the
     `in` string-vs-int match trap). */}}
{{ $am := getMember $author.ID }}
{{ if $am }}{{ range $am.Roles }}{{ $rid := str . }}{{ range $sr := $staffRoles }}{{ if eq $rid $sr }}{{ return }}{{ end }}{{ end }}{{ end }}{{ end }}

{{/* Idempotent: one bonus per post, ever. Flag keyed by message ID, expiring
     after 90 days (nobody re-reacts to ancient posts; keeps the DB tidy).
     Hyphens only — DB pattern helpers treat "_" as a wildcard. */}}
{{ $flag := printf "ptsbonus-%d" .Reaction.MessageID }}
{{ if dbGet $author.ID $flag }}{{ return }}{{ end }}
{{ dbSetExpire $author.ID $flag 1 7776000 }}

{{/* Bank the bonus onto the lifetime total (same key as points_earn.go). */}}
{{ $newTotal := toInt (dbIncr $author.ID "pts-total" $bonusPoints) }}

{{/* Audit log (embed → the mentions render without pinging). */}}
{{ if $logChannel }}{{ sendMessage $logChannel (cembed "title" "Bonus points" "description" (printf "<@%d> gave <@%d> **+%d** bonus point(s) — now **%d** total.\n[Jump to post](https://discord.com/channels/%d/%d/%d)" .Reaction.UserID $author.ID $bonusPoints $newTotal .Guild.ID .Reaction.ChannelID .Reaction.MessageID) "color" $color) }}{{ end }}

{{/* Optional public celebration (NoEscape so the member actually gets pinged). */}}
{{ if $announce }}{{ sendMessageNoEscape .Reaction.ChannelID (printf "🎉 <@%d> earned **+%d** bonus participation point(s) from staff!" $author.ID $bonusPoints) }}{{ end }}

{{/* ── Instant tier-up for the RECIPIENT (not the reactor). Same math as the
     earner, but the target is $author, so use giveRoleID/takeRoleID (by user ID)
     rather than addRoleID/removeRoleID. Base comes from pts-base (single source,
     seeded by the badge sweep); if unseeded ($base 0) we skip and let the sweep
     handle it. The sweep still owns demotions / rebalances. */}}
{{ $be := dbGet 0 "pts-base" }}{{ $base := 0 }}{{ if $be }}{{ $base = toInt $be.Value }}{{ end }}
{{ if gt $base 0 }}
  {{/* tier role IDs from the single source (points_badge_sweep.go → pts-tierroles) */}}
  {{ $trEntry := dbGet 0 "pts-tierroles" }}{{ $tierRoles := cslice }}{{ if $trEntry }}{{ $tierRoles = $trEntry.Value }}{{ end }}
  {{ $oldTotal := sub $newTotal $bonusPoints }}
  {{ $oldTier := -1 }}{{ $newTier := -1 }}
  {{ range $i, $r := $tierRoles }}
    {{ $n := add $i 1 }}{{ $need := mult $base (mult $n $n) }}
    {{ if ge $oldTotal $need }}{{ $oldTier = $i }}{{ end }}
    {{ if ge $newTotal $need }}{{ $newTier = $i }}{{ end }}
  {{ end }}
  {{ if gt $newTier $oldTier }}
    {{ $nr := index $tierRoles $newTier }}
    {{ if $nr }}{{ giveRoleID $author.ID $nr }}{{ end }}
    {{ if ge $oldTier 0 }}{{ $or := index $tierRoles $oldTier }}{{ if and $or (ne $or $nr) }}{{ takeRoleID $author.ID $or }}{{ end }}{{ end }}
  {{ end }}
{{ end }}
