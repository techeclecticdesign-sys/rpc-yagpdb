{{/* participation_points / BADGE SWEEP — keeps tier ROLE badges in sync with
     cumulative points (points_badge_sweep.go).

     Trigger type: Interval — every hour. Leave "run in channel" blank (it posts
     nothing). Interval commands use NO message-trigger slot. Hourly is plenty:
     the earner AND the bonus already promote instantly on tier-up, so this sweep
     only needs to catch the rarer cases — demotions (staff lowered points) and
     base rebalances. Lower the interval temporarily if you want those to converge
     faster right after a big rebalance.

     WHY AN INTERVAL RECONCILER: a member's correct tier changes when they earn
     (up), when staff adjust points (up/down), and when staff rebalance the base
     (everyone at once). This loop recomputes each member's tier from the LIVE
     base every pass and fixes their roles — so nothing has to "detect" a base
     change; the sweep just makes reality match the formula. The earner also
     promotes instantly on tier-up (points_earn.go); THIS is the safety net that
     additionally handles demotions, rebalances, and anyone the earner missed.

     The badge itself is a Discord ROLE ICON: each tier is a role whose icon is
     that tier's emoji (a Boost Level 2 perk). Discord shows the icon of a
     member's highest icon-bearing role, so each member should hold exactly ONE
     tier role. YAGPDB needs Manage Roles and its own role ABOVE every tier role.

     FREE-TIER SAFE via the same rotating-cursor pattern as advert_expiry: scan a
     modest page of pts-total behind a saved cursor, reconcile up to $changeBudget
     members per run (the role writes are the rate-limited part), wrap at the end.
     Only ~5 DB ops/run (two config publishes + cursor get + one dbTopEntries +
     cursor set). */}}

{{/* ──────────────── CONFIG ──────────────── */}}
{{/* ▼▼ SINGLE SOURCE OF TRUTH for the tier thresholds — the point ladder, tier 1
       → tier 10, ascending. THIS IS THE ONLY PLACE the ladder lives. This sweep
       publishes it to the DB key `pts-thresholds` every run, and every OTHER
       command reads it from there (no command hardcodes thresholds) — the same
       single-source pattern as the tier role IDs below. To recalibrate the ladder,
       edit these ten numbers and nowhere else; the change propagates on the next
       sweep. Keep the ORDER aligned with the $tiers names/emojis (points.go /
       points_staff.go) and the $tierRoles below. ▼▼ */}}
{{ $tierThresholds := cslice 200 400 800 1500 2300 3300 4500 6000 7700 10000 }}
{{/* ▼▼ Tier BADGE role IDs, tier 1 → tier 10, as strings ("" = no badge for that
       tier). THIS IS THE ONLY PLACE they live — edit them here and nowhere else.
       This sweep publishes them to the DB key pts-tierroles (below), and the
       earner / bonus / staff commands read them from there (single source of truth,
       same pattern as $tierThresholds → pts-thresholds). Keep the ORDER aligned with
       the $tiers names/emojis in points.go / points_staff.go. ▼▼ */}}
{{ $tierRoles := cslice "Pollen Puff" "Busy Bee" "Flower Forager" "Bumble Bard" "Honey Helper" "Meadow Muse" "Nectar Novelist" "Hive Hero" "Garden Guardian" "Bee's Knees" }}
{{/* ▼▼ Members scanned per run, and max role reconciliations per run. Keep the
       scan modest so getMember stays cheap and role writes stay under the API
       cap; the cursor wraps so everyone is reached over successive runs. Lower
       both if you ever see "too many calls"; raise for faster convergence. ▼▼ */}}
{{ $pageSize     := 50 }}
{{ $changeBudget := 5 }}
{{/* ─────────────────────────────────────── */}}

{{/* Publish the tier thresholds so the earner/bonus/staff/slash read them from ONE
     place. Overwrite every run (pure config, no live-edit command), so an edit to
     $tierThresholds above propagates on the next sweep. */}}
{{ dbSet 0 "pts-thresholds" $tierThresholds }}

{{/* Publish the tier role IDs so the earner/bonus/staff read them from ONE place.
     Overwrite every run (they're pure config with no live-edit command, so this
     makes an edit to $tierRoles above propagate on the next sweep). */}}
{{ dbSet 0 "pts-tierroles" $tierRoles }}

{{/* rotating scan cursor (owner ID 0 = sweeper globals, same idea as the advert
     sweeper's advertExpiryCursor). */}}
{{ $skip := toInt (dbGet 0 "pts-badgeCursor").Value }}
{{ $page := dbTopEntries "pts-total" $pageSize $skip }}

{{ range $page }}
  {{ if gt $changeBudget 0 }}
    {{ $uid := .UserID }}
    {{ $pts := toInt .Value }}
    {{/* target tier index for these points: highest tier (0-based) whose threshold
         (index i of the ladder) is ≤ pts. Thresholds ascend, so last pass wins. */}}
    {{ $target := -1 }}
    {{ range $i, $r := $tierRoles }}{{ if lt $i (len $tierThresholds) }}{{ if ge $pts (toInt (index $tierThresholds $i)) }}{{ $target = $i }}{{ end }}{{ end }}{{ end }}
    {{ $targetRole := "" }}{{ if ge $target 0 }}{{ $targetRole = index $tierRoles $target }}{{ end }}
    {{/* inspect the member's current tier roles (string compare — never `in`,
         which won't match a string against an int64 role id). */}}
    {{ $m := getMember $uid }}
    {{ if $m }}
      {{ $hasTarget := false }}{{ $wrong := cslice }}
      {{ range $m.Roles }}
        {{ $rid := str . }}
        {{ range $tr := $tierRoles }}{{ if and $tr (eq $tr $rid) }}
          {{ if eq $rid $targetRole }}{{ $hasTarget = true }}{{ else }}{{ $wrong = $wrong.Append $rid }}{{ end }}
        {{ end }}{{ end }}
      {{ end }}
      {{ if or (gt (len $wrong) 0) (and (ne $targetRole "") (not $hasTarget)) }}
        {{/* small delay so a member who just left can't abort the run and to
             spread the role API calls, exactly like advert_expiry's deletes. */}}
        {{ range $wr := $wrong }}{{ takeRoleID $uid $wr 3 }}{{ end }}
        {{ if and (ne $targetRole "") (not $hasTarget) }}{{ giveRoleID $uid $targetRole 3 }}{{ end }}
        {{ $changeBudget = sub $changeBudget 1 }}
      {{ end }}
    {{ end }}
  {{ end }}
{{ end }}

{{/* advance the cursor; wrap to 0 once a page comes up short (end reached). Note
     a run that exhausts $changeBudget leaves the rest of its page for the next
     lap — same catch-up behavior as the advert sweeper. */}}
{{ $newSkip := add $skip $pageSize }}
{{ if lt (len $page) $pageSize }}{{ $newSkip = 0 }}{{ end }}
{{ dbSet 0 "pts-badgeCursor" $newSkip }}
