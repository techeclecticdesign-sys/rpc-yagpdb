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
{{- $count := 0 -}}{{- $oldest := 0 -}}
{{- $prev := (dbGet $u.ID "infractionDates").Value -}}
{{- if $prev -}}{{- range $prev -}}{{- $t := toInt . -}}{{- if ge $t $cut -}}{{- $count = add $count 1 -}}{{- if or (eq $oldest 0) (lt $t $oldest) -}}{{- $oldest = $t -}}{{- end -}}{{- end -}}{{- end -}}{{- end -}}
{{- $lines := cslice (printf "<@%d> has **%d** advert-infraction(s) in the last 6 months." $u.ID $count) -}}
{{- if gt $count 0 -}}{{- $lines = $lines.Append (printf "Oldest drops off: <t:%d:D>." (add $oldest $w)) -}}{{- end -}}
{{- $ban := dbGet $u.ID "advertBan" -}}
{{- if $ban.Value -}}{{- $lines = $lines.Append (printf "⛔ **Advert-banned** until <t:%d:F>." (toInt $ban.ExpiresAt.Unix)) -}}{{- else -}}{{- $lines = $lines.Append "✅ Not advert-banned." -}}{{- end -}}
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
{{- dbSet $u.ID "infractionDates" $dates -}}
Set <@{{ $u.ID }}>'s advert-infraction count to **{{ $n }}** (counted from now, expiring in 6 months).{{ if ge $n 4 }} Note: this does not advert-ban them; the ban applies on their next infraction.{{ end }}
{{- end -}}
{{- else -}}
Unknown action.
{{- end -}}
