{{/* =====================================================================
     Banned-word filter for the advert channels.

     If any WORD in a post contains one of the banned words as a SUBSTRING,
     the poster is pinged in #rule_infractions, with a link to the channel and
     the offending word(s) shown in ||spoiler tags|| so the term isn't
     displayed openly in the infractions channel.

     Substring matching is intentional (so "xSLURy" is caught), but it means a
     banned word can match inside an innocent word too (banning "ass" would
     flag "class"/"passage"). Choose entries with that in mind.

     The ping is sent through the shared "Alert Sender" command after a short
     delay, so it is skipped if the post was already deleted (e.g. for length
     or cooldown). See setup.txt for full step-by-step dashboard instructions.

     Trigger type: Regex
     Trigger: ([\s\S]*)
     Channel Restrictions: all advert channels.
     ===================================================================== */}}

{{/* ▼▼ Paste your "Alert Sender" command ID here (see setup.txt) ▼▼ */}}
{{ $alertSender := 0 }}

{{/* ▼▼ Banned words — lowercase. A post word that CONTAINS any of these
       (as a substring) is flagged. ▼▼ */}}
{{ $banned := cslice
  "exampleword"
  "anotherword"
}}

{{ if gt (len $banned) 0 }}
  {{/* Build one case-insensitive pattern that captures the WHOLE word
       containing any banned substring:  \S*(?:a|b|c)\S*
       reQuoteMeta escapes each entry so regex characters are treated as
       literal text. */}}
  {{ $escaped := cslice }}
  {{ range $banned }}{{ $escaped = $escaped.Append (reQuoteMeta .) }}{{ end }}
  {{ $pattern := printf "(?i)\\S*(?:%s)\\S*" (joinStr "|" $escaped) }}

  {{ $hits := reFindAll $pattern .Message.Content }}
  {{ if gt (len $hits) 0 }}
    {{/* De-duplicate (case-insensitively) and wrap each in spoiler tags. */}}
    {{ $seen := sdict }}
    {{ $spoilered := cslice }}
    {{ range $hits }}
      {{ $key := lower . }}
      {{ if not ($seen.Get $key) }}
        {{ $seen.Set $key true }}
        {{ $spoilered = $spoilered.Append (printf "||%s||" .) }}
      {{ end }}
    {{ end }}

    {{ execCC $alertSender nil 5 (sdict
      "channelID" .Channel.ID
      "msgID"     .Message.ID
      "text"      (printf "Hey %s ! Your post in %s contains wording that isn't allowed here: %s. Please edit it out. Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID) (joinStr " " $spoilered))
    ) }}
  {{ end }}
{{ end }}
