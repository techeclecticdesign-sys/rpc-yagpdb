{{/* =====================================================================
     INFRACTION RE-CHECK — staged follow-up that clears the :staffpending:
     flag once an infracted advert has actually been fixed.

     Trigger type: None   (runs ONLY when called via scheduleUniqueCC).

     infractions_sticky kicks off stage 1 ten minutes after an advert command
     pings a post; this command then re-checks at 45 / 90 / 480 minutes (~8h).
     At each stage it re-runs that channel type's advisory checks on the CURRENT
     post:
       • clean  → remove the :staffpending: reaction, mark the ping
                  :staffapproved:, and stop.
       • gone   → stop (the reaction died with the deleted post).
       • dirty  → schedule the next stage. At the final 8h stage a still-dirty
                  post is DELETED and its ping is marked :staffapproved:.

     Every stage is a delayed (scheduleUniqueCC) call, so each runs as its own
     fresh depth-0 execution with its own 1-call runcc budget — the chain length
     is irrelevant to YAGPDB's limits.

     The unique key "infr_<msgID>" means one chain per post (and it can be
     cancelled with cancelScheduledUniqueCC if ever needed).
     See setup.txt for dashboard instructions.
     ===================================================================== */}}

{{ $d := .ExecData }}
{{ if not $d }}{{ return }}{{ end }}

{{- /* ===== CONFIG ===== */ -}}
{{/* ▼▼ This command's OWN id, so it can schedule its next stage. ▼▼ */}}
{{ $selfCC := 0 }}

{{/* ▼▼ The :staffpending: emoji as name:id — MUST match what the advert
       commands add, or it won't be removed. ▼▼ */}}
{{ $staffPending := "staffpending:1442331141771366513" }}

{{/* ▼▼ The :staffapproved: emoji as name:id — added to the #rule_infractions
       ping when a post comes back clean. Leave "" to skip. Type \:staffapproved:
       in Discord to read its id. ▼▼ */}}
{{ $staffApproved := "" }}

{{/* ▼▼ Banned words — keep identical to the three advert commands. ▼▼ */}}
{{ $banned := cslice
  "futa"
  "futanari"
  "futas"
  "futanaris"
}}

{{- /* ===== fetch the post; if it's gone, so is its reaction ===== */ -}}
{{ $msg := getMessage $d.channelID $d.msgID }}
{{ if not $msg }}{{ return }}{{ end }}

{{ $type := $d.type }}
{{ $dirty := false }}

{{- /* ===== re-run that channel type's advisory checks ===== */ -}}

{{/* links — quick only */}}
{{ if eq $type "quick" }}
  {{ if reFind "(?i)https?://\\S+" $msg.Content }}{{ $dirty = true }}{{ end }}
{{ end }}

{{/* images / attachments — quick & 1x1 (can't be edited off, so this stays
     dirty until the author deletes + reposts, which lands in the "gone" case) */}}
{{ if or (eq $type "quick") (eq $type "1x1") }}
  {{ if gt (len $msg.Attachments) 0 }}{{ $dirty = true }}{{ end }}
{{ end }}

{{/* headers — quick & 1x1 forbid any; group allows one short line */}}
{{ if or (eq $type "quick") (eq $type "1x1") }}
  {{ range (split $msg.Content "\n") }}
    {{ if or (hasPrefix . "# ") (hasPrefix . "## ") (hasPrefix . "### ") }}{{ $dirty = true }}{{ end }}
  {{ end }}
{{ else if eq $type "group" }}
  {{ $headers := cslice }}
  {{ range (split $msg.Content "\n") }}
    {{ if or (hasPrefix . "# ") (hasPrefix . "## ") (hasPrefix . "### ") }}{{ $headers = $headers.Append . }}{{ end }}
  {{ end }}
  {{ if gt (len $headers) 1 }}{{ $dirty = true }}{{ end }}
  {{ range $headers }}
    {{ $line := . }}
    {{ $cap := 50 }}{{ $prefixLen := 2 }}
    {{ if hasPrefix $line "### " }}{{ $cap = 70 }}{{ $prefixLen = 4 }}
    {{ else if hasPrefix $line "## " }}{{ $cap = 60 }}{{ $prefixLen = 3 }}{{ end }}
    {{ if gt (len (toRune (slice $line $prefixLen))) $cap }}{{ $dirty = true }}{{ end }}
  {{ end }}
{{ end }}

{{/* banned words — all types */}}
{{ if and (not $dirty) (gt (len $banned) 0) }}
  {{ $escaped := cslice }}
  {{ range $banned }}{{ $escaped = $escaped.Append (reQuoteMeta .) }}{{ end }}
  {{ if reFind (printf "(?i)\\b(?:%s)\\b" (joinStr "|" $escaped)) $msg.Content }}{{ $dirty = true }}{{ end }}
{{ end }}

{{/* cross-channel duplicate — all types. Gated behind (not $dirty) because it
     makes getMessage calls; no point paying for them once we already know the
     post is dirty. */}}
{{ if not $dirty }}
  {{ $thisNorm := lower $msg.Content }}
  {{ $thisNorm = reReplace "(?i)https?://\\S+" $thisNorm " " }}
  {{ $thisNorm = reReplace "[^a-z0-9 ]+" $thisNorm " " }}
  {{ $thisNorm = reReplace "\\s+" $thisNorm " " }}
  {{ $thisNorm = trimSpace $thisNorm }}
  {{ if ge (len $thisNorm) 15 }}
    {{ $thisChannel := str $d.channelID }}
    {{ range (dbGetPattern $msg.Author.ID "lastMsg_%" 100 0) }}
      {{ if and (not $dirty) (hasPrefix .Key "lastMsg_") }}
        {{ $cid := slice .Key 8 }}
        {{ if ne $cid $thisChannel }}
          {{ $other := getMessage $cid .Value }}
          {{ if $other }}
            {{ $o := lower $other.Content }}
            {{ $o = reReplace "(?i)https?://\\S+" $o " " }}
            {{ $o = reReplace "[^a-z0-9 ]+" $o " " }}
            {{ $o = reReplace "\\s+" $o " " }}
            {{ $o = trimSpace $o }}
            {{ if eq $o $thisNorm }}{{ $dirty = true }}{{ end }}
          {{ end }}
        {{ end }}
      {{ end }}
    {{ end }}
  {{ end }}
{{ end }}

{{- /* ===== resolved? clear the flag, approve the ping, and stop ===== */ -}}
{{ if not $dirty }}
  {{/* drop :staffpending: from the offending advert post */}}
  {{ deleteAllMessageReactions $d.channelID $d.msgID $staffPending }}
  {{/* and mark the #rule_infractions ping itself :staffapproved: */}}
  {{ if $staffApproved }}{{ if $d.infractionChannel }}{{ if $d.infractionMsgID }}
    {{ addMessageReactions $d.infractionChannel $d.infractionMsgID $staffApproved }}
  {{ end }}{{ end }}{{ end }}
  {{ return }}
{{ end }}

{{ $stage := toInt $d.stage }}

{{- /* ===== still dirty at the final (8h) stage → delete the post and close
       the ping with :staffapproved: ===== */ -}}
{{ if ge $stage 4 }}
  {{ deleteMessage $d.channelID $d.msgID 0 }}
  {{ if $staffApproved }}{{ if $d.infractionChannel }}{{ if $d.infractionMsgID }}
    {{ addMessageReactions $d.infractionChannel $d.infractionMsgID $staffApproved }}
  {{ end }}{{ end }}{{ end }}
  {{ return }}
{{ end }}

{{- /* ===== still dirty, not yet final → schedule the next stage ===== */ -}}
{{ $next := 0 }}
{{ if eq $stage 1 }}{{ $next = 2100 }}       {{/* +35m  → 45m  */}}
{{ else if eq $stage 2 }}{{ $next = 2700 }}  {{/* +45m  → 90m  */}}
{{ else if eq $stage 3 }}{{ $next = 23400 }} {{/* +390m → 480m (8h) */}}
{{ end }}
{{ if and $next $selfCC }}
  {{ scheduleUniqueCC $selfCC nil $next (joinStr "" "infr_" $d.msgID) (sdict
    "msgID"             $d.msgID
    "channelID"         $d.channelID
    "type"              $type
    "stage"             (add $stage 1)
    "infractionChannel" $d.infractionChannel
    "infractionMsgID"   $d.infractionMsgID
  ) }}
{{ end }}
