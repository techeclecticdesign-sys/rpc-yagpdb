{{/* participation_points / BONUS + DEDUCT — staff-reaction point adjustments
     (points_bonus.go).

     Trigger type: Reaction (a reaction is added/removed).
     Channel Restrictions: your participation / event channels ONLY (same
     whitelist as points_earn.go). A Reaction trigger fires on a REACTION event,
     not a message, so it does NOT count against YAGPDB's "3 custom commands per
     message" budget that the earner/advert/infraction regex commands share.

     TWO staff reactions, both optional (each a silent no-op until its emoji is set):
       • AWARD  ($bonusEmoji)  — adds points to the post's author.
       • DEDUCT ($deductEmoji) — removes a flat $deductPoints from the author
                                 (floored at 0). Default "" = inactive.

     AWARD amount is per-channel: $bonusByChannel maps a channel ID → the award
     for a post in that channel; channels not listed use $bonusPoints. So a staff
     award in an "event" channel can be worth more than one in a general channel.

     WEEKLY channels ($weeklyChannels): in a channel listed here, any one member can
     only receive the AWARD once per Sun–Sat calendar week (UTC, per channel). The award still
     fires once-per-post everywhere; the weekly gate additionally blocks a SECOND
     different post by the same member in that channel within the week. Deduct is
     never weekly-limited.

     Both actions are idempotent per post (separate flags), promote/demote the
     author's tier BADGE immediately (giveRoleID/takeRoleID on the author — no wait
     for the hourly sweep), and skip bots, self-reactions, and STAFF authors (staff
     are exempt from the board/tiers).

     DB ops (worst case, weekly award): dbGet(post flag) + dbGet(week flag) +
     dbSetExpire(post flag) + dbSetExpire(week flag) + dbIncr + dbGet(pts-thresholds)
     + dbGet(pts-tierroles) = 7 db_interactions, one getMember, and (only on a tier
     change) one give + one take role call — inside the free-tier cap of 10. */}}

{{/* ──────────────── CONFIG ──────────────── */}}
{{/* ▼▼ DEFAULT award for a staff reaction, used for any channel NOT listed in
       $bonusByChannel below. ▼▼ */}}
{{ $bonusPoints := 10 }}

{{/* ▼▼ PER-CHANNEL award amounts. Keys are channel IDs as STRINGS, values are the
       award for a post in that channel. A channel not listed here falls back to
       $bonusPoints. Leave as `sdict` (empty) to make every channel worth the
       default. Example:
         {{ $bonusByChannel := sdict "112233445566778899" 25 "998877665544332211" 50 }}
       ▼▼ */}}
{{ $bonusByChannel := sdict "645478254182137897" 5 "1312243973657596005" 20 "458681372219932673" 20 "461912346021986305" 20}}

{{/* ▼▼ WEEKLY channels — channel IDs as STRINGS. In a channel listed here a member
       can only receive the AWARD once per Sun–Sat calendar week (UTC, per channel).
       Empty = no weekly limit anywhere. Example:
         {{ $weeklyChannels := cslice "112233445566778899" }}
       ▼▼ */}}
{{ $weeklyChannels := cslice "328913968749871113" "645478254182137897" "1185283601928040518" "458681372219932673" "461912346021986305"}}

{{/* ▼▼ The AWARD emoji as name:id — SAME format as :staffapproved:/:staffpending:.
       Leave "" until you've uploaded the emoji; the award is a silent no-op until
       it's set. Type \:youremoji: in Discord to read its name:id. Matched by ID,
       so renaming the emoji later won't break it.
       e.g. "pointsaward:1442331141771366513"  (a bare "1442331141771366513"
       also works). ▼▼ */}}
{{ $bonusEmoji := "staffapproved:1358879847664975902" }}

{{/* ▼▼ The DEDUCT emoji as name:id, same format. OPTIONAL — leave "" to disable
       deductions entirely (default). When set, a staff reaction with this emoji
       removes $deductPoints from the author, floored at 0. Do NOT set it to the
       same emoji as $bonusEmoji (award is matched first and would win). ▼▼ */}}
{{ $deductEmoji := "spray_bottle:715379514871644222" }}
{{ $deductPoints := 15 }}

{{/* ▼▼ Staff role ID(s) as strings. Dual purpose: only these roles may GRANT/DEDUCT,
       and a member with any of these roles can never RECEIVE either (staff are
       exempt from the board/tiers). Keep identical to $staffRoles in
       points_earn.go. ▼▼ */}}
{{ $staffRoles := cslice "300831005621878784" "322845008409395200" "479484736188973087" "1371346095380238376" "1376679985603022899" }}

{{/* ▼▼ Mod-log / audit channel ID for a record of each grant/deduct, or 0 to disable. ▼▼ */}}
{{ $logChannel := 1532466633929654272 }}

{{/* ▼▼ true → also post a short public note (and ping the member) in the channel
       where the post lives. false → adjust silently. ▼▼ */}}
{{ $announce := false }}
{{ $color := 0xF4700F }}
{{/* ─────────────────────────────────────── */}}
{{/* Tier ladder + BADGE role IDs are NOT configured here — they live once in
     points_badge_sweep.go and are read from the DB keys pts-thresholds /
     pts-tierroles below. */}}

{{/* Add-only (ignore reaction removals). */}}
{{ if not .ReactionAdded }}{{ return }}{{ end }}

{{/* Which action? Match the reacted emoji's ID against the award emoji, then the
     deduct emoji. Each is a silent no-op while its config is "". Award wins if both
     are (mis)configured to the same emoji. */}}
{{ $reacted := str .Reaction.Emoji.ID }}
{{ $awardID := "" }}{{ if $bonusEmoji }}{{ $awardID = $bonusEmoji }}{{ $ep := split $bonusEmoji ":" }}{{ if eq (len $ep) 2 }}{{ $awardID = index $ep 1 }}{{ end }}{{ end }}
{{ $deductID := "" }}{{ if $deductEmoji }}{{ $deductID = $deductEmoji }}{{ $dp := split $deductEmoji ":" }}{{ if eq (len $dp) 2 }}{{ $deductID = index $dp 1 }}{{ end }}{{ end }}
{{ $mode := "" }}
{{ if and $awardID (eq $reacted $awardID) }}{{ $mode = "award" }}{{ else if and $deductID (eq $reacted $deductID) }}{{ $mode = "deduct" }}{{ end }}
{{ if eq $mode "" }}{{ return }}{{ end }}

{{/* Reactor must be staff. In a Reaction trigger the triggering member is the
     one who reacted, so hasRoleID checks the reactor. */}}
{{ $isStaff := false }}
{{ range $sr := $staffRoles }}{{ if hasRoleID (toInt64 $sr) }}{{ $isStaff = true }}{{ end }}{{ end }}
{{ if not $isStaff }}{{ return }}{{ end }}

{{/* The target post + its author. */}}
{{ $post := .ReactionMessage }}
{{ if not $post }}{{ return }}{{ end }}
{{ $author := $post.Author }}
{{ if not $author }}{{ return }}{{ end }}
{{ if $author.Bot }}{{ return }}{{ end }}
{{/* No self-adjustments. */}}
{{ if eq $author.ID .Reaction.UserID }}{{ return }}{{ end }}

{{/* Staff author is exempt (same as the earner). .ReactionMessage.Author has no
     roles, so pull the member and compare role IDs as strings (avoids the `in`
     string-vs-int match trap). */}}
{{ $am := getMember $author.ID }}
{{ if $am }}{{ range $am.Roles }}{{ $rid := str . }}{{ range $sr := $staffRoles }}{{ if eq $rid $sr }}{{ return }}{{ end }}{{ end }}{{ end }}{{ end }}

{{/* Resolve the adjustment. AWARD = per-channel amount (or the default); DEDUCT =
     a flat −$deductPoints. */}}
{{ $chanKey := str .Reaction.ChannelID }}
{{ $delta := 0 }}
{{ if eq $mode "award" }}
  {{ $award := $bonusPoints }}
  {{ $perChan := $bonusByChannel.Get $chanKey }}
  {{ if $perChan }}{{ $award = toInt $perChan }}{{ end }}
  {{ $delta = $award }}
{{ else }}
  {{ $delta = mult $deductPoints -1 }}
{{ end }}

{{/* Is this a weekly-limited channel? String compare (str on each entry so an
     unquoted ID still matches — avoids the `in` cross-type trap). */}}
{{ $isWeekly := false }}
{{ range $wc := $weeklyChannels }}{{ if eq (str $wc) $chanKey }}{{ $isWeekly = true }}{{ end }}{{ end }}

{{/* Idempotent per post, per action. Flags keyed by message ID, expiring after 90
     days. Hyphens only — DB pattern helpers treat "_" as a wildcard. */}}
{{ $flag := printf "ptsbonus-%d" .Reaction.MessageID }}
{{ if eq $mode "deduct" }}{{ $flag = printf "ptsdeduct-%d" .Reaction.MessageID }}{{ end }}
{{ if dbGet $author.ID $flag }}{{ return }}{{ end }}

{{/* Weekly gate (AWARD in a weekly channel only). Check BEFORE burning the post
     flag so a blocked post isn't marked "used". "Week" is a fixed Sun–Sat calendar
     week (UTC): the key embeds the week number so the window always resets at
     Sunday 00:00 UTC, not 7 days after the last award. $weekNo = the count of
     Sun–Sat weeks since the 1970 epoch — the epoch day is a Thursday, so +4 shifts
     the week boundary onto Sunday. (add/div return float64, so toInt each step:
     printf %d and correct week math both need ints.) Per channel, per week. */}}
{{ $days := toInt (div (toInt64 currentTime.Unix) 86400) }}
{{ $weekNo := toInt (div (toInt (add $days 4)) 7) }}
{{ $weekFlag := printf "ptsweek-%d-%d" .Reaction.ChannelID $weekNo }}
{{ if and (eq $mode "award") $isWeekly }}{{ if dbGet $author.ID $weekFlag }}{{ return }}{{ end }}{{ end }}

{{/* Commit the flags: this post is now used, and (weekly award) mark this member
     as awarded for the current Sun–Sat week in this channel. 8-day expiry — the
     week number is baked into the key, so it only needs to outlast the remainder
     of the current week; the next week uses a fresh key regardless. */}}
{{ dbSetExpire $author.ID $flag 1 7776000 }}
{{ if and (eq $mode "award") $isWeekly }}{{ dbSetExpire $author.ID $weekFlag 1 691200 }}{{ end }}

{{/* Apply to the lifetime total (same key as points_earn.go). dbIncr is atomic, so
     the raw result minus the delta is the true pre-value (needed for the tier
     math even when a deduct clamps to 0). Clear the entry when it hits 0 to keep
     the board tidy, mirroring the staff `set 0` / remove-floor behavior. */}}
{{ $raw := toInt (dbIncr $author.ID "pts-total" $delta) }}
{{ $oldTotal := sub $raw $delta }}
{{ $newTotal := $raw }}
{{ if le $newTotal 0 }}{{ dbDel $author.ID "pts-total" }}{{ $newTotal = 0 }}{{ end }}

{{/* Audit log (embed → the mentions render without pinging). */}}
{{ if $logChannel }}
  {{ if eq $mode "award" }}{{ sendMessage $logChannel (cembed "title" "Bonus points" "description" (printf "<@%d> gave <@%d> **+%d** point(s) — now **%d** total.\n[Jump to post](https://discord.com/channels/%d/%d/%d)" .Reaction.UserID $author.ID $delta $newTotal .Guild.ID .Reaction.ChannelID .Reaction.MessageID) "color" $color) }}
  {{ else }}{{ sendMessage $logChannel (cembed "title" "Points deducted" "description" (printf "<@%d> removed **%d** point(s) from <@%d> — now **%d** total.\n[Jump to post](https://discord.com/channels/%d/%d/%d)" .Reaction.UserID $deductPoints $author.ID $newTotal .Guild.ID .Reaction.ChannelID .Reaction.MessageID) "color" $color) }}{{ end }}
{{ end }}

{{/* Optional public note (NoEscape so the member actually gets pinged). */}}
{{ if $announce }}
  {{ if eq $mode "award" }}{{ sendMessageNoEscape .Reaction.ChannelID (printf "🎉 <@%d> earned **+%d** bonus participation point(s) from staff!" $author.ID $delta) }}
  {{ else }}{{ sendMessageNoEscape .Reaction.ChannelID (printf "<@%d> had **%d** participation point(s) removed by staff." $author.ID $deductPoints) }}{{ end }}
{{ end }}

{{/* ── Instant badge reconcile for the RECIPIENT (not the reactor). Bidirectional:
     an award can promote, a deduct can demote. Compute the old + new tier from the
     ladder (DB key pts-thresholds, published by the badge sweep) and, if they
     differ, give the new tier's role and take the old one. Uses giveRoleID/
     takeRoleID (by user ID) since the target is the author. If the ladder isn't
     seeded yet we skip and let the hourly sweep reconcile. */}}
{{ $thEntry := dbGet 0 "pts-thresholds" }}{{ $thresholds := cslice }}{{ if $thEntry }}{{ $thresholds = $thEntry.Value }}{{ end }}
{{ if gt (len $thresholds) 0 }}
  {{ $trEntry := dbGet 0 "pts-tierroles" }}{{ $tierRoles := cslice }}{{ if $trEntry }}{{ $tierRoles = $trEntry.Value }}{{ end }}
  {{ $oldTier := -1 }}{{ $newTier := -1 }}
  {{ range $i, $r := $tierRoles }}
    {{ if lt $i (len $thresholds) }}{{ $need := toInt (index $thresholds $i) }}
    {{ if ge $oldTotal $need }}{{ $oldTier = $i }}{{ end }}
    {{ if ge $newTotal $need }}{{ $newTier = $i }}{{ end }}{{ end }}
  {{ end }}
  {{ if ne $newTier $oldTier }}
    {{ $nr := "" }}{{ if ge $newTier 0 }}{{ $nr = index $tierRoles $newTier }}{{ end }}
    {{ if $nr }}{{ giveRoleID $author.ID $nr }}{{ end }}
    {{ if ge $oldTier 0 }}{{ $or := index $tierRoles $oldTier }}{{ if and $or (ne $or $nr) }}{{ takeRoleID $author.ID $or }}{{ end }}{{ end }}
  {{ end }}
{{ end }}
