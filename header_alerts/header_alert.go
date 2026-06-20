{{/* =====================================================================
     Discord renders a line that starts with #, ## or ### followed by a space
     as a header. When a post contains one, this pings the poster in
     #rule_infractions asking them to remove it.

     The ping is sent through the shared "Alert Sender" command after a short
     delay, so it is skipped if the post was already deleted (e.g. for length
     or cooldown). See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{/* ▼▼ Paste your "Alert Sender" command ID here (see setup.txt) ▼▼ */}}
{{ $alertSender := 0 }}

{{ execCC $alertSender nil 5 (sdict
  "channelID" .Channel.ID
  "msgID"     .Message.ID
  "text"      (printf "Hello %s ! You are welcome to use bold, however, headers are not allowed in the one on one advert channels. Please go back to your advert in %s and remove the header. Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID))
) }}
