{{/* Advert + member-intro expiry sweep — deletes posts older than 60 days
     (group advert channels keep posts for 120 days — see $groupChannels below).
     Trigger: Minute Interval, every 60 minutes, run in #bot_spam. See setup.txt.

     Rides entirely on the DB records the advert commands and the Member Intros
     command already write when they accept a post:
       lastMsg_<channelID>          the member's live post ID (string)
       lastMsgTime_<channelID>[...] their cooldown timestamp (time)
     so those commands need NO changes and stay at their 10-DB-op ceiling.

     A post's age is read straight out of its message-ID snowflake
     (ms = id/2^22 + 1420070400000), so nothing new is stored per post.

     Per run: 1-2 dbTopEntries pages (multi-entry cap is 2) walk the records
     behind a rotating cursor; up to $opBudget expired records are acted on
     (deleteMessage is scheduled with a short delay so a missing message can't
     abort the run, then the record is dbDel'd). The cursor wraps when a scan
     reaches the end, so anything skipped by a full budget is caught on the
     next lap. Steady state is a handful of expiries a day — far under the
     ~144/day this drains at. */}}

{{/* ▼▼ Post lifetime, seconds. 5184000 = 60 days. ▼▼ */}}
{{ $ttlSecs := 5184000 }}

{{/* ▼▼ Group advert channels keep posts LONGER: 10368000 = 120 days (4 months)
       instead of the 60-day default above. Paste every group advert channel ID
       here as strings, e.g. (cslice "123" "456"). A post whose channel is in this
       list is only swept once it passes $groupTtlSecs; every other channel still
       uses $ttlSecs. Leave the list empty to treat all channels the same. ▼▼ */}}
{{ $groupTtlSecs := 10368000 }}
{{ $groupChannels := cslice }}

{{/* ▼▼ Channel IDs (as strings) whose lastMsg_ records must NOT expire, e.g.
       (cslice "1234" "5678"). Every channel with lastMsg_ records is an advert
       channel or member intros today, so this stays empty; it's a guard in
       case a future command reuses the key scheme somewhere that shouldn't
       auto-expire. (Stale lastMsgTime_ cooldown records are cleared past the
       TTL regardless — cooldowns are hours long, so they're long dead.) ▼▼ */}}
{{ $excludeChannels := cslice }}

{{ $cutoff := sub (toInt currentTime.Unix) $ttlSecs }}
{{ $groupCutoff := sub (toInt currentTime.Unix) $groupTtlSecs }}
{{ $opBudget := 6 }}{{/* free-tier cap is 10 db_interactions/run. Overhead is 4:
     the cursor dbGet + cursor dbSet + the TWO dbTopEntries pages (dbTopEntries
     counts against db_interactions too, not only db_multiple). 10 - 4 = 6 dbDels
     left. Do NOT raise past 6 without cutting overhead, or the run errors mid-loop
     on "too many calls" and the cursor never advances. */}}

{{/* rotating scan cursor (owner ID 0 = sweeper globals, same idea as the
     stickies keeping state under the channel ID) */}}
{{ $skip := toInt (dbGet 0 "advertExpiryCursor").Value }}

{{ $entries := cslice }}
{{ $page := dbTopEntries "lastMsg%" 100 $skip }}
{{ if $page }}{{ $entries = $entries.AppendSlice $page }}{{ end }}
{{ if eq (len $entries) 100 }}
  {{ $page = dbTopEntries "lastMsg%" 100 (add $skip 100) }}
  {{ if $page }}{{ $entries = $entries.AppendSlice $page }}{{ end }}
{{ end }}

{{ range $entries }}
  {{ if hasPrefix .Key "lastMsg_" }}
    {{/* live post record — key carries the channel, value is the message ID */}}
    {{ $cid := slice .Key 8 }}
    {{ $mid := toInt .Value }}
    {{ $postedAt := add (div (div $mid 4194304) 1000) 1420070400 }}
    {{/* group advert channels get the longer 120-day window */}}
    {{ $cut := $cutoff }}{{ if in $groupChannels $cid }}{{ $cut = $groupCutoff }}{{ end }}
    {{ if and (gt $opBudget 0) (lt $postedAt $cut) (not (in $excludeChannels $cid)) }}
      {{ deleteMessage (toInt $cid) $mid 5 }}
      {{ dbDel .UserID .Key }}
      {{ $opBudget = sub $opBudget 1 }}
    {{ end }}
  {{ else if hasPrefix .Key "lastMsgTime" }}
    {{/* cooldown record — value is a time; clearing past-TTL ones just keeps
         the DB tight (covers both lastMsgTime_<cid> and lastMsgTime_<cid>_96h
         style keys) */}}
    {{ $ts := 0 }}
    {{ if .Value }}{{ $ts = toInt .Value.Unix }}{{ end }}
    {{ if and (gt $opBudget 0) (lt $ts $cutoff) }}
      {{ dbDel .UserID .Key }}
      {{ $opBudget = sub $opBudget 1 }}
    {{ end }}
  {{ end }}
{{ end }}

{{/* advance the cursor; wrap to 0 once a scan comes up short (end reached) */}}
{{ $newSkip := 0 }}
{{ if eq (len $entries) 200 }}{{ $newSkip = add $skip 200 }}{{ end }}
{{ dbSet 0 "advertExpiryCursor" $newSkip }}
