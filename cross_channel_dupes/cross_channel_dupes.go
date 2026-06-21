{{/* =====================================================================
     Cross-channel duplicate advert detector.

     On a match it hands off to the shared "Alert Sender" command, which only
     pings if THIS post still exists (so if the advert command already deleted
     this post for length/cooldown, no duplicate ping is sent).

     ===================================================================== */}}

{{/* ▼▼ Paste your "Alert Sender" command ID here (see setup.txt) ▼▼ */}}
{{ $alertSender := 0 }}

{{/* --- Normalize THIS post: lowercase, drop links, keep only a-z0-9 + spaces,
       collapse whitespace. This mirrors the normalize block in the loop. --- */}}
{{ $this := .Message.Content | lower }}
{{ $this = reReplace "(?i)https?://\\S+" $this " " }}
{{ $this = reReplace "[^a-z0-9 ]+" $this " " }}
{{ $this = reReplace "\\s+" $this " " }}
{{ $this = trimSpace $this }}

{{/* Skip posts with too little text to judge meaningfully. */}}
{{ if lt (len $this) 15 }}{{ return }}{{ end }}

{{ $thisChannel := str .Channel.ID }}
{{ $dupChannel := "" }}

{{/* --- Compare against the user's ads in their OTHER advert channels.
       Note: "_" is a single-char wildcard in the pattern, so this also returns
       lastMsgTime_ keys; hasPrefix "lastMsg_" keeps only the message-ID keys. */}}
{{ range (dbGetPattern .User.ID "lastMsg_%" 100 0) }}
  {{ if and (not $dupChannel) (hasPrefix .Key "lastMsg_") }}
    {{ $channelID := slice .Key 8 }}
    {{ if ne $channelID $thisChannel }}
      {{ $other := getMessage $channelID .Value }}
      {{ if $other }}
        {{ $o := $other.Content | lower }}
        {{ $o = reReplace "(?i)https?://\\S+" $o " " }}
        {{ $o = reReplace "[^a-z0-9 ]+" $o " " }}
        {{ $o = reReplace "\\s+" $o " " }}
        {{ $o = trimSpace $o }}
        {{ if eq $o $this }}{{ $dupChannel = $channelID }}{{ end }}
      {{ end }}
    {{ end }}
  {{ end }}
{{ end }}

{{ if $dupChannel }}
  {{ execCC $alertSender nil 5 (sdict
    "channelID" .Channel.ID
    "msgID"     .Message.ID
    "text"      (printf "Hey %s ! Your advert in %s looks identical to your ad in %s. The same advert can't be posted in more than one channel — please make them distinctly different or remove one. Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID) (printf "<#%s>" $dupChannel))
  ) }}
{{ end }}
