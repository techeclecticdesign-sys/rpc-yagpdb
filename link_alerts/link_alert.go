{{/* =====================================================================
     When someone posts a link in one of the quick channels, this pings the
     poster in #rule_infractions asking them to remove it from their ad.

     The ping is sent through the shared "Alert Sender" command after a short
     delay, so it is skipped if the post was already deleted (e.g. for length
     or cooldown). See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{/* ▼▼ Paste your "Alert Sender" command ID here (see setup.txt) ▼▼ */}}
{{ $alertSender := 0 }}

{{ execCC $alertSender nil 5 (sdict
  "channelID" .Channel.ID
  "msgID"     .Message.ID
  "text"      (printf "Hey %s !  Links aren’t allowed in the quick search channels. Please remove it from your ad in %s. Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID))
) }}
