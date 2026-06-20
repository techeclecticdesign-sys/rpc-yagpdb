{{/* =====================================================================
     Shared "Alert Sender" for the rule-infraction ping commands
     (link_alert, header_alert, no_images_in_1x1, group_header_check).

     Those commands run on the same post as the advert length/cooldown/
     duplicate command, which may DELETE the post. To avoid pinging about a
     post that no longer exists, they don't ping directly — they hand their
     message to this command with a short delay, and this command only pings
     if the post is still there.

     Trigger type: None  (only runs when another command calls it via execCC)
     See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{ $data := .ExecData }}
{{ if not $data }}{{ return }}{{ end }}

{{/* ▼▼ Paste your #rule_infractions channel ID here (see setup.txt) ▼▼ */}}
{{ $infractionsChannel := 0 }}

{{/* Only ping if the post survived — the advert length/cooldown/duplicate
     command may have already deleted it in the few seconds since the post. */}}
{{ if getMessage $data.channelID $data.msgID }}
  {{ sendMessage $infractionsChannel $data.text }}
{{ end }}
