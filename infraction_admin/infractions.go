{{/* =====================================================================
     -infractions -- staff management of the advert-infraction counter.

     Trigger type: Command.   Name: infractions

       -infractions view @member       show count + dated list + ban status
       -infractions @member            alias for view
       -infractions clear @member      wipe history AND lift ban
       -infractions set @member 4      set the count to 4, applying the ban

     Restrict this command to staff in the dashboard. See setup.txt.

     DATA MODEL: infractions live in the per-user "infractionLog" entry -- a
     cslice of records {t,r,c,m}: unix time, a comma-joined reason (e.g.
     "headers, banned word"), and the offending post's channel/message id for a
     jump link. Older installs stored a bare "infractionDates" timestamp list;
     those legacy entries are read here too (shown reason-less) and are migrated
     into infractionLog the next time the member infracts. Both are pruned only
     by the 6-month window -- nothing is dropped on the 4th infraction. Every
     infraction from the 4th on (re)applies a fresh 14-day advert ban.
     ===================================================================== */}}

{{- $usage := "Usage: `-infractions view @member`, `-infractions clear @member`, or `-infractions set @member <1-99>`." -}}
{{- $action := "view" -}}
{{- $userArg := "" -}}
{{- $countArg := "" -}}

{{- if gt (len .CmdArgs) 0 -}}
  {{- $first := lower (index .CmdArgs 0) -}}
  {{- if in (cslice "view" "clear" "set") $first -}}
    {{- $action = $first -}}
    {{- if gt (len .CmdArgs) 1 -}}{{- $userArg = index .CmdArgs 1 -}}{{- end -}}
    {{- if gt (len .CmdArgs) 2 -}}{{- $countArg = index .CmdArgs 2 -}}{{- end -}}
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
{{- range $entries -}}
{{- $line := printf "- <t:%d:D>" (toInt .t) -}}
{{- if .r -}}{{- $line = printf "%s - %s" $line .r -}}{{- else -}}{{- $line = printf "%s - reason not recorded" $line -}}{{- end -}}
{{- /* Only render the jump link if the post still exists -- a link to a deleted
       message still yanks the viewer over to that channel for nothing. getMessage
       returns nil when the message is gone. Counts against the 100 API-calls/CC
       budget, but the in-window count is bounded (set caps at 99) so we're safe. */ -}}
{{- if and .c .m -}}{{- if (getMessage (toInt .c) (toInt .m)) -}}{{- $line = printf "%s ([jump to post](https://discord.com/channels/%s/%s/%s))" $line (str $.Guild.ID) (str .c) (str .m) -}}{{- end -}}{{- end -}}
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
{{- else if eq $action "set" -}}
{{- $n := 0 -}}{{- if and $countArg (reFind `^\d+$` $countArg) -}}{{- $n = toInt $countArg -}}{{- end -}}
{{- if or (lt $n 1) (gt $n 99) -}}
For **set**, provide a count of 1-99. (Use the **clear** action to zero someone out.)
{{- else -}}
{{- $now := toInt currentTime.Unix -}}{{- $recs := cslice -}}
{{- range seq 0 $n -}}{{- $recs = $recs.Append (sdict "t" $now "r" "set by staff" "c" "" "m" "") -}}{{- end -}}
{{- dbSet $u.ID "infractionLog" $recs -}}
{{- dbDel $u.ID "infractionDates" -}}
{{- if ge $n 4 -}}
{{- dbSetExpire $u.ID "advertBan" $now 1209600 -}}{{/* 14-day ban -- keep equal to the advert commands */}}
Set <@{{ $u.ID }}>'s advert-infraction count to **{{ $n }}** - that's at or over the limit, so they've been advert-banned for 14 days. Their history is kept, so any further infraction re-applies a fresh 14-day ban.
{{- else -}}
{{- $wasBanned := false -}}{{- if (dbGet $u.ID "advertBan").Value -}}{{- $wasBanned = true -}}{{- end -}}
{{- dbDel $u.ID "advertBan" -}}
Set <@{{ $u.ID }}>'s advert-infraction count to **{{ $n }}** (counted from now, expiring in 6 months).{{ if $wasBanned }} That's below the limit, so their active advert ban has been lifted.{{ end }}
{{- end -}}
{{- end -}}
{{- else -}}
{{ $usage }}
{{- end -}}
