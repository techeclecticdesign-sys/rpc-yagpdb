{{/* =====================================================================
     In a group advert exactly ONE header line is allowed, and its visible
     text has a per-level character cap (about one line on a 1080p screen):
         #   ->  50 characters
         ##  ->  60 characters
         ### ->  70 characters
     If a post has more than one header line, or a header longer than its cap,
     this pings the poster in #rule_infractions.

     The ping is sent through the shared "Alert Sender" command after a short
     delay, so it is skipped if the post was already deleted (e.g. for length
     or cooldown). See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{/* ▼▼ Paste your "Alert Sender" command ID here (see setup.txt) ▼▼ */}}
{{ $alertSender := 0 }}

{{/* Collect the header lines by checking the post ONE LINE AT A TIME.
     A header line starts with "# ", "## " or "### ". */}}
{{ $headers := cslice }}
{{ range (split .Message.Content "\n") }}
  {{ if or (hasPrefix . "# ") (hasPrefix . "## ") (hasPrefix . "### ") }}
    {{ $headers = $headers.Append . }}
  {{ end }}
{{ end }}

{{/* Rule 1: only one header line is allowed. */}}
{{ $tooMany := gt (len $headers) 1 }}

{{/* Rule 2: each header's visible text must fit the cap for its level. */}}
{{ $tooLong := false }}
{{ range $headers }}
  {{ $line := . }}
  {{/* default to level 1 (#); bump for ## and ### — check the longest first */}}
  {{ $cap := 50 }}
  {{ $prefixLen := 2 }}
  {{ if hasPrefix $line "### " }}
    {{ $cap = 70 }}{{ $prefixLen = 4 }}
  {{ else if hasPrefix $line "## " }}
    {{ $cap = 60 }}{{ $prefixLen = 3 }}
  {{ end }}
  {{/* visible header text = the line with the "#... " prefix removed */}}
  {{ $text := slice $line $prefixLen }}
  {{/* character count (counts runes so emoji/accents aren't over-counted) */}}
  {{ $textLen := len (reFindAll "." $text) }}
  {{ if gt $textLen $cap }}{{ $tooLong = true }}{{ end }}
{{ end }}

{{ if or $tooMany $tooLong }}
  {{ execCC $alertSender nil 5 (sdict
    "channelID" .Channel.ID
    "msgID"     .Message.ID
    "text"      (printf "Hello %s ! Group adverts may only have one line of header text. You can use regular bold for additional lines. Please fix your post in %s.  Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID))
  ) }}
{{ end }}
