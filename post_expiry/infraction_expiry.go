{{/* #rule_infractions expiry sweep — deletes channel posts older than 180 days.
     Trigger: Minute Interval, every 360 minutes, run in #bot_spam. See setup.txt.

     Messages in #rule_infractions aren't uniquely indexed the way adverts are
     (one member can have many pings), so per-message DB entries would bloat
     the entry budget. Instead the infractions sticky appends every message ID
     to a 30-day bucket entry:
       key   infrLedger_<unixTime / 2592000>   (owner ID = the channel, like
                                                the sticky's own state)
       value cslice of message-ID strings
     Only ~7 buckets exist at any time (6 months of history + the live one),
     and a bucket of even thousands of IDs sits far below the 100 kB value cap.

     Each run checks the bucket straddling the 180-day line plus the three
     before it (self-healing after downtime): IDs older than the cutoff — age
     read from the message-ID snowflake — are deleted (short deleteMessage
     delay, so a missing message can't abort the run) and the bucket is
     rewritten without them; an emptied bucket is dbDel'd. Worst case 8 DB ops
     (4 reads + 4 writes), inside the 10-op cap. Deletions are chunked to
     $maxDeletes per run; leftovers are picked up next run. */}}

{{/* ▼▼ Paste your #rule_infractions channel ID here (see setup.txt) ▼▼ */}}
{{ $infractionsChannel := 641835326314381312 }}

{{/* ▼▼ Post lifetime, seconds. 15552000 = 180 days (matches the rolling
       infraction-count window). ▼▼ */}}
{{ $ttlSecs := 15552000 }}

{{/* ▼▼ Max messages deleted per run. 12 × 4 runs/day = 48/day of headroom. ▼▼ */}}
{{ $maxDeletes := 12 }}

{{ $bucketSecs := 2592000 }}{{/* 30 days — must match the sticky's ledger */}}
{{ $cutoff := sub (toInt currentTime.Unix) $ttlSecs }}
{{ $target := div $cutoff $bucketSecs }}
{{ $budget := $maxDeletes }}

{{ range $i := seq 0 4 }}
  {{ $key := printf "infrLedger_%d" (sub $target $i) }}
  {{ $entry := dbGet $infractionsChannel $key }}
  {{ if $entry }}{{ if $entry.Value }}
    {{ $keep := cslice }}
    {{ $deleted := 0 }}
    {{ range $entry.Value }}
      {{ $mid := toInt . }}
      {{ $postedAt := add (div (div $mid 4194304) 1000) 1420070400 }}
      {{ if and (gt $budget 0) (lt $postedAt $cutoff) }}
        {{ deleteMessage $infractionsChannel $mid 5 }}
        {{ $budget = sub $budget 1 }}
        {{ $deleted = add $deleted 1 }}
      {{ else }}
        {{ $keep = $keep.Append . }}
      {{ end }}
    {{ end }}
    {{ if gt $deleted 0 }}
      {{ if len $keep }}
        {{ dbSet $infractionsChannel $key $keep }}
      {{ else }}
        {{ dbDel $infractionsChannel $key }}
      {{ end }}
    {{ end }}
  {{ end }}{{ end }}
{{ end }}
