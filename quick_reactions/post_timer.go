{{/* =====================================================================
     Fires on every new post in the quick channels and schedules the
     "Reaction Check" command (reaction_check.go) to run 5 minutes later.
     See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{/* ▼▼ Paste the ID of your "Reaction Check" command here (see setup.txt) ▼▼ */}}
{{ $checkCC := 0 }}

{{/* ▼▼ How long to wait before checking the post, in seconds (600 = 10 min) ▼▼ */}}
{{ $delaySeconds := 300 }}

{{ execCC $checkCC nil $delaySeconds (sdict
  "msgID"          .Message.ID
  "channelID"      .Channel.ID
  "userMention"    (printf "<@%d>" .User.ID)
  "channelMention" (printf "<#%d>" .Channel.ID)
) }}
