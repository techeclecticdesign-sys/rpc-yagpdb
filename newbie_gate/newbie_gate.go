{{/* =====================================================================
     NEWBIE GATE — onboarding sweep (kick / newbie fall-off / graduation)

     Trigger type: MINUTE INTERVAL (e.g. every 30 min), run in a quiet channel
     (e.g. #bot_spam — nothing is posted there).

     Instead of one scheduled job per member, this is ONE interval command that
     sweeps a list of pending members we keep in the database. Every join writes a
     row `gatePending` (owner = the member's id, value = their join unix time) via
     join_message.go; this command pages that list with dbTopEntries behind a
     rotating cursor (exactly like post_expiry/advert_expiry.go) and, for each
     member, reads their CURRENT roles and:
       • has a real role (not newbie/age-please/rules-please/excluded)
             → drop the newbie tag, remove them from the list (graduated)
       • has newbie, no real role
             → at 7 days, drop the newbie tag and remove them (fall-off); else leave
       • no newbie, no real role (stuck at a gate)
             → kick once they've been in the server 24h+ (leave the row; the kick's
               leave event / the next sweep clears it)
       • member left → remove the row

     WHY a sweep and not per-member scheduled jobs: YAGPDB has no cap on the number
     of pending scheduled CCs, but delayed CC runs are rate-limited to ~6/min per
     channel and OVER-BUDGET RUNS ARE SILENTLY DROPPED (customcommands/handle_timed.go).
     A dropped hop would strand a member forever (never kicked / newbie never
     removed). An interval trigger is re-armed by YAGPDB itself and can't be
     dropped, so this design is self-healing. Costs no message-trigger slots.

     Kicking uses YAGPDB's own Kick command via execAdmin — needs the Moderation
     module's Kick command enabled + the bot's Kick-Members perm with its role above
     members. Full write-up + install in setup.txt.
     ===================================================================== */}}

{{- /* ===== CONFIG ===== */ -}}

{{/* ▼▼ Newbie role ID (granted by the rules reaction). REQUIRED. ▼▼ */}}
{{ $newbieRole := "456657052136243210" }}

{{/* ▼▼ The two "waiting at a gate" role IDs. BOTH REQUIRED — worn by un-finished
       members, so both must be named to be kept out of the "real role" test,
       otherwise a member sitting at a gate reads as graduated and is never kicked. ▼▼ */}}
{{ $agePleaseRole := "735544386678554738" }}
{{ $rulesPleaseRole := "634585076524646401" }}

{{/* ▼▼ Roles that DON'T count as "picking a role" (Server Booster, bot/level roles,
       event/cosmetic roles, and — your call — the #get_roles mailbox status roles).
       Role IDs as strings, e.g. (cslice "111..." "222..."). ▼▼ */}}
{{ $excludedRoles := cslice }}

{{/* ▼▼ Timings (seconds). ▼▼ */}}
{{ $kickSeconds := 86400 }}{{/* 24h — kick a member in the server this long with no newbie role */}}
{{ $newbieFalloff := 604800 }}{{/* 7d — newbie tag auto-removal */}}

{{/* ▼▼ Kick toggle. true = kick (needs the Moderation Kick command + bot perm —
       see setup.txt). false = never kick (gate-stuck rows are dropped after the
       fall-off window so the list can't grow forever). ▼▼ */}}
{{ $enableKick := true }}
{{ $kickReason := "Did not pass the entry gates (age verification + rules) within 24 hours of joining." }}

{{/* ▼▼ Per-run limits (free-tier safety). $pageSize = members inspected per run;
       $opBudget = DB deletions per run (graduations/fall-offs/cleanups) — keep
       $opBudget ≤ 7 so total db ops (3 overhead + $opBudget) stay under 10;
       $kickBudget = kicks per run (bounds mod actions / avoids mass-kick). ▼▼ */}}
{{ $pageSize := 50 }}
{{ $opBudget := 6 }}
{{ $kickBudget := 3 }}

{{ $now := toInt currentTime.Unix }}

{{- /* rotating scan cursor (owner 0 = sweeper globals, like advert_expiry) */ -}}
{{ $skip := toInt (dbGet 0 "gateSweepCursor").Value }}
{{ $entries := dbTopEntries "gatePending" $pageSize $skip }}

{{ range $entries }}
  {{ if or (gt $opBudget 0) (gt $kickBudget 0) }}
    {{ $owner := .UserID }}
    {{ $uid := str .UserID }}
    {{ $age := sub $now (toInt .Value) }}
    {{ $m := getMember $uid }}
    {{ if not $m }}
      {{- /* member left / already kicked → drop the row */ -}}
      {{ if gt $opBudget 0 }}{{ dbDel $owner "gatePending" }}{{ $opBudget = sub $opBudget 1 }}{{ end }}
    {{ else }}
      {{ $hasNewbie := false }}
      {{ $hasReal := false }}
      {{ range $m.Roles }}
        {{ $r := str . }}
        {{ if eq $r $newbieRole }}{{ $hasNewbie = true }}
        {{ else if and (ne $agePleaseRole "0") (eq $r $agePleaseRole) }}{{/* age gate tag */}}
        {{ else if and (ne $rulesPleaseRole "0") (eq $r $rulesPleaseRole) }}{{/* rules gate tag */}}
        {{ else if in $excludedRoles $r }}{{/* excluded */}}
        {{ else }}{{ $hasReal = true }}{{ end }}
      {{ end }}
      {{ if $hasReal }}
        {{- /* graduated → drop newbie, remove from list */ -}}
        {{ if gt $opBudget 0 }}
          {{ if $hasNewbie }}{{ takeRoleID $uid $newbieRole }}{{ end }}
          {{ dbDel $owner "gatePending" }}{{ $opBudget = sub $opBudget 1 }}
        {{ end }}
      {{ else if $hasNewbie }}
        {{- /* past gates, no roles → drop the tag at 7d, else leave the row */ -}}
        {{ if and (ge $age $newbieFalloff) (gt $opBudget 0) }}
          {{ takeRoleID $uid $newbieRole }}
          {{ dbDel $owner "gatePending" }}{{ $opBudget = sub $opBudget 1 }}
        {{ end }}
      {{ else }}
        {{- /* no newbie, no real role → gate-stuck */ -}}
        {{ if $enableKick }}
          {{ if and (ge $age $kickSeconds) (gt $kickBudget 0) }}
            {{ $_ := execAdmin (printf "kick %s %s" $uid $kickReason) }}{{ $kickBudget = sub $kickBudget 1 }}
            {{- /* no dbDel: the kick's leave event (or the next sweep's getMember=nil)
                   clears the row; a failed kick keeps it so we retry next lap */ -}}
          {{ end }}
        {{ else if and (ge $age $newbieFalloff) (gt $opBudget 0) }}
          {{- /* kick disabled: stop tracking after the fall-off window (no bloat) */ -}}
          {{ dbDel $owner "gatePending" }}{{ $opBudget = sub $opBudget 1 }}
        {{ end }}
      {{ end }}
    {{ end }}
  {{ end }}
{{ end }}

{{- /* advance the cursor; wrap to 0 once a page comes up short (end reached) */ -}}
{{ $newSkip := 0 }}
{{ if eq (len $entries) $pageSize }}{{ $newSkip = add $skip $pageSize }}{{ end }}
{{ dbSet 0 "gateSweepCursor" $newSkip }}
