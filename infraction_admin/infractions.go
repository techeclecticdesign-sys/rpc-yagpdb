{{/* =====================================================================
     /infractions — staff management of the advert-infraction counter.
     One slash command, three actions (uses a single slash-command slot).

     Trigger type: Slash Command.   Name: infractions
     Dashboard → Options (add in this order; Names must match EXACTLY):
       • Name "action", Type "Text menu", Required
            choices (Value must be lowercase):  view   clear   set
       • Name "user",   Type "User",      Required
       • Name "count",  Type "Integer",   NOT required   (only used by "set")
     Read in code as .Options.action / .Options.user / .Options.count.
     Restrict to staff. See setup.txt.
     ===================================================================== */}}
{{- $action := .Options.action -}}
{{- $u := .Options.user -}}
{{- $w := 15552000 -}}{{/* 180-day window — keep equal to the advert commands */}}

{{- if eq $action "view" -}}
{{- $cut := (add (toInt currentTime.Unix) (mult $w -1)) -}}
{{- $days := cslice -}}
{{- $prev := (dbGet $u.ID "infractionDates").Value -}}
{{- if $prev -}}{{- range $prev -}}{{- $t := toInt . -}}{{- if ge $t $cut -}}{{- $days = $days.Append $t -}}{{- end -}}{{- end -}}{{- end -}}
{{- $count := len $days -}}
{{- $ban := dbGet $u.ID "advertBan" -}}
{{- $lines := cslice -}}
{{- if $ban.Value -}}{{- $lines = $lines.Append (printf "<@%d> is ⛔ Advert-banned until <t:%d:F>." $u.ID (toInt $ban.ExpiresAt.Unix)) -}}{{- else -}}{{- $lines = $lines.Append (printf "<@%d> is ✅ Not advert-banned." $u.ID) -}}{{- end -}}
{{- $lines = $lines.Append (printf "They have **%d** advert-infraction(s) in the last 6 months." $count) -}}
{{- if gt $count 0 -}}{{- $lines = $lines.Append "User has infractions on the following days:" -}}{{- range $days -}}{{- $lines = $lines.Append (printf "- <t:%d:D>" .) -}}{{- end -}}{{- end -}}
{{- ephemeralResponse -}}
{{ joinStr "\n" $lines }}
{{- else if eq $action "clear" -}}
{{- $had := 0 -}}{{- $prev := (dbGet $u.ID "infractionDates").Value -}}{{- if $prev -}}{{- $had = len $prev -}}{{- end -}}
{{- $wasBanned := false -}}{{- if (dbGet $u.ID "advertBan").Value -}}{{- $wasBanned = true -}}{{- end -}}
{{- dbDel $u.ID "infractionDates" -}}
{{- dbDel $u.ID "advertBan" -}}
Cleared <@{{ $u.ID }}>'s advert-infraction history — {{ $had }} record(s) removed{{ if $wasBanned }}, and lifted their active advert ban{{ end }}.
{{- else if eq $action "set" -}}
{{- $n := 0 -}}{{- if .Options.count -}}{{- $n = toInt .Options.count -}}{{- end -}}
{{- if or (lt $n 1) (gt $n 99) -}}
For **set**, provide a count of 1-99. (Use the **clear** action to zero someone out.)
{{- else -}}
{{- $now := toInt currentTime.Unix -}}{{- $dates := cslice -}}
{{- range seq 0 $n -}}{{- $dates = $dates.Append $now -}}{{- end -}}
{{- if ge $n 4 -}}
{{- dbSetExpire $u.ID "advertBan" $now 1209600 -}}{{/* 14-day ban — keep equal to the advert commands */}}
{{- dbDel $u.ID "infractionDates" -}}
Set <@{{ $u.ID }}>'s advert-infraction count to **{{ $n }}** — that's at or over the limit, so they've been advert-banned for 14 days (history reset, same as a live 4th infraction).
{{- else -}}
{{- $wasBanned := false -}}{{- if (dbGet $u.ID "advertBan").Value -}}{{- $wasBanned = true -}}{{- end -}}
{{- dbSet $u.ID "infractionDates" $dates -}}
{{- dbDel $u.ID "advertBan" -}}
Set <@{{ $u.ID }}>'s advert-infraction count to **{{ $n }}** (counted from now, expiring in 6 months).{{ if $wasBanned }} That's below the limit, so their active advert ban has been lifted.{{ end }}
{{- end -}}
{{- end -}}
{{- else -}}
Unknown action.
{{- end -}}
