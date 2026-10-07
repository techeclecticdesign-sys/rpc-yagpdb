{{/* =====================================================================
     -infractions -- staff management of the advert-infraction counter.

     Trigger type: Command.   Name: infractions

       -infractions view @member       show count + numbered list + ban status
       -infractions @member            alias for view
       -infractions clear @member      wipe history AND lift ban
       -infractions drop @member 2     drop a single record by its list index
                                       (the number shown by `view`). If this
                                       takes them under 4, any active ban is
                                       lifted; the ban is never extended.
       -infractions add @member reason append a hand-written infraction dated
                                       now with the given reason (the reason may
                                       be several words). If this reaches 4+, a
                                       fresh 14-day advert ban is applied.

     Restrict this command to staff in the dashboard. See setup.txt.

     DATA MODEL: infractions live in the per-user "infractionLog" entry -- a
     cslice of records {t,r,c,m}: unix time, a comma-joined reason (e.g.
     "headers, banned word"), and the offending post's channel/message id for a
     jump link. Older installs stored a bare "infractionDates" timestamp list;
     those legacy entries are read here too (shown reason-less) and are migrated
     into infractionLog the next time the member infracts. Both are pruned only
     by the 6-month window -- nothing is dropped on the 4th infraction. Every
     infraction from the 4th on (re)applies a fresh 14-day advert ban.

     Records are stored oldest-first (each new infraction is appended) and the
     `view`/`drop` index is 1-based over that same in-window order, so the number
     you see in `view` is the number you pass to `drop`. `drop` rebuilds the log
     without that one record (migrating any legacy entries in the process); `add`
     appends a fresh record dated now with a staff-written reason.
     ===================================================================== */}}

{{- $usage := "Usage: `-infractions view @member`, `-infractions clear @member`, `-infractions drop @member <index>`, or `-infractions add @member <reason>`." -}}
{{- $action := "view" -}}
{{- $userArg := "" -}}
{{- $idxArg := "" -}}
{{- $reasonArg := "" -}}

{{- if gt (len .CmdArgs) 0 -}}
  {{- $first := lower (index .CmdArgs 0) -}}
  {{- if in (cslice "view" "clear" "drop" "add") $first -}}
    {{- $action = $first -}}
    {{- if gt (len .CmdArgs) 1 -}}{{- $userArg = index .CmdArgs 1 -}}{{- end -}}
    {{- if and (eq $first "drop") (gt (len .CmdArgs) 2) -}}{{- $idxArg = index .CmdArgs 2 -}}{{- end -}}
    {{- /* "add" reason is everything after the user -- the last argument may be several words. */ -}}
    {{- if and (eq $first "add") (gt (len .CmdArgs) 2) -}}{{- $reasonArg = joinStr " " (slice .CmdArgs 2) -}}{{- end -}}
  {{- else -}}
    {{- $userArg = index .CmdArgs 0 -}}
  {{- end -}}
{{- end -}}

{{- $u := 0 -}}
{{- if $userArg -}}{{- $u = userArg $userArg -}}{{- end -}}
{{- if not $u -}}
{{ $usage }}
{{- return -}}
{{- end -}}

{{- $w := 15552000 -}}{{/* 180-day window -- keep equal to the advert commands */}}
{{- $cut := (add (toInt currentTime.Unix) (mult $w -1)) -}}

{{- /* Merge legacy timestamps + rich log into one list of {t,r,c,m} records,
       in-window only. The two keys never both hold live data (a write migrates
       + deletes the legacy key), so this can't double-count. */ -}}
{{- $entries := cslice -}}
{{- $legacy := (dbGet $u.ID "infractionDates").Value -}}
{{- if $legacy -}}{{- range $legacy -}}{{- if ge (toInt .) $cut -}}{{- $entries = $entries.Append (sdict "t" (toInt .) "r" "" "c" "" "m" "") -}}{{- end -}}{{- end -}}{{- end -}}
{{- $log := (dbGet $u.ID "infractionLog").Value -}}
{{- if $log -}}{{- range $log -}}{{- if ge (toInt .t) $cut -}}{{- $entries = $entries.Append . -}}{{- end -}}{{- end -}}{{- end -}}
{{- $count := len $entries -}}

{{- if eq $action "view" -}}
{{- $ban := dbGet $u.ID "advertBan" -}}
{{- $lines := cslice -}}
{{- if $ban.Value -}}{{- $lines = $lines.Append (printf "<@%d> is advert-banned until <t:%d:F>." $u.ID (toInt $ban.ExpiresAt.Unix)) -}}{{- else -}}{{- $lines = $lines.Append (printf "<@%d> is not advert-banned." $u.ID) -}}{{- end -}}
{{- $lines = $lines.Append (printf "They have **%d** advert-infraction(s) in the last 6 months." $count) -}}
{{- if gt $count 0 -}}
{{- $lines = $lines.Append "Infractions:" -}}
{{- range $i, $e := $entries -}}
{{- $line := printf "%d\\. <t:%d:D>" (toInt (add $i 1)) (toInt $e.t) -}}
{{- if $e.r -}}{{- $line = printf "%s - %s" $line $e.r -}}{{- else -}}{{- $line = printf "%s - reason not recorded" $line -}}{{- end -}}
{{- /* Only render the jump link if the post still exists -- a link to a deleted
       message still yanks the viewer over to that channel for nothing. getMessage
       returns nil when the message is gone. Counts against the 100 API-calls/CC
       budget, but only real recorded posts carry a channel/message (hand-added
       records don't), and members are banned at 4, so the count stays small. */ -}}
{{- if and $e.c $e.m -}}{{- if (getMessage (toInt $e.c) (toInt $e.m)) -}}{{- $line = printf "%s ([jump to post](https://discord.com/channels/%s/%s/%s))" $line (str $.Guild.ID) (str $e.c) (str $e.m) -}}{{- end -}}{{- end -}}
{{- $lines = $lines.Append $line -}}
{{- end -}}
{{- end -}}
{{ joinStr "\n" $lines }}
{{- else if eq $action "clear" -}}
{{- $wasBanned := false -}}{{- if (dbGet $u.ID "advertBan").Value -}}{{- $wasBanned = true -}}{{- end -}}
{{- dbDel $u.ID "infractionLog" -}}
{{- dbDel $u.ID "infractionDates" -}}
{{- dbDel $u.ID "advertBan" -}}
Cleared <@{{ $u.ID }}>'s advert-infraction history - {{ $count }} record(s) removed{{ if $wasBanned }}, and lifted their active advert ban{{ end }}.
{{- else if eq $action "drop" -}}
{{- $idx := 0 -}}{{- if and $idxArg (reFind `^\d+$` $idxArg) -}}{{- $idx = toInt $idxArg -}}{{- end -}}
{{- if eq $count 0 -}}
<@{{ $u.ID }}> has no advert-infractions to drop.
{{- else if or (lt $idx 1) (gt $idx $count) -}}
For **drop**, provide an index of 1-{{ $count }}. Run `-infractions view @member` to see the numbered list.
{{- else -}}
{{- /* Rebuild the log without the 1-based $idx record. $entries is the same
       in-window, oldest-first list `view` numbers, so the index lines up.
       Writing $recs migrates any legacy timestamps into infractionLog. */ -}}
{{- $recs := cslice -}}{{- $removed := sdict "t" 0 "r" "" -}}
{{- range $i, $e := $entries -}}{{- if eq (toInt (add $i 1)) $idx -}}{{- $removed = $e -}}{{- else -}}{{- $recs = $recs.Append $e -}}{{- end -}}{{- end -}}
{{- dbSet $u.ID "infractionLog" $recs -}}
{{- dbDel $u.ID "infractionDates" -}}
{{- $newCount := len $recs -}}
{{- $wasBanned := false -}}{{- if (dbGet $u.ID "advertBan").Value -}}{{- $wasBanned = true -}}{{- end -}}
{{- $lifted := false -}}{{- if and $wasBanned (lt $newCount 4) -}}{{- dbDel $u.ID "advertBan" -}}{{- $lifted = true -}}{{- end -}}
Dropped infraction #{{ $idx }} for <@{{ $u.ID }}> (dated <t:{{ toInt $removed.t }}:D>{{ if $removed.r }} - {{ $removed.r }}{{ end }}). They now have **{{ $newCount }}** advert-infraction(s) in the last 6 months.{{ if $lifted }} That's below the limit, so their active advert ban has been lifted.{{ end }}
{{- end -}}
{{- else if eq $action "add" -}}
{{- if not $reasonArg -}}
For **add**, provide a reason: `-infractions add @member <reason>`.
{{- else -}}
{{- /* Neutralise pings in the free-text reason before it is stored -- a zero-width
       space after every "@" breaks @everyone/@here and <@id>/<@&role> mentions, so
       the reason can't ping when it is echoed here or re-shown by `view`. */ -}}
{{- $reasonArg = reReplace "@" $reasonArg "@​" -}}
{{- /* Append a hand-written record dated now. Writing $recs also migrates any
       legacy timestamps into infractionLog, so infractionDates can be dropped. */ -}}
{{- $now := toInt currentTime.Unix -}}
{{- $recs := $entries.Append (sdict "t" $now "r" $reasonArg "c" "" "m" "") -}}
{{- dbSet $u.ID "infractionLog" $recs -}}
{{- dbDel $u.ID "infractionDates" -}}
{{- $newCount := len $recs -}}
{{- if ge $newCount 4 -}}
{{- dbSetExpire $u.ID "advertBan" $now 1209600 -}}{{/* 14-day ban -- keep equal to the advert commands */}}
Added an advert-infraction for <@{{ $u.ID }}> - {{ $reasonArg }}. They now have **{{ $newCount }}** in the last 6 months, which is at or over the limit, so they've been advert-banned for 14 days. Any further infraction re-applies a fresh 14-day ban.
{{- else -}}
Added an advert-infraction for <@{{ $u.ID }}> - {{ $reasonArg }}. They now have **{{ $newCount }}** in the last 6 months.
{{- end -}}
{{- end -}}
{{- else -}}
{{ $usage }}
{{- end -}}
