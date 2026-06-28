{{ $maxLength := 2100 }}
{{ $lockoutHours := 168 }}
{{ $infractionsChannel := 0 }}
{{ $infractionsSticky := 0 }}
{{ $botSpam := 0 }}
{{ $infrWindowSecs := 15552000 }}{{/* 180 days */}}
{{ $advertBanSecs := 1209600 }}{{/* 14 days */}}
{{ $staffPending := "staffpending:1442331141771366513" }}
{{ $banned := cslice
  "futa"
  "futanari"
  "futas"
  "futanaris"
}}

{{ $argLength := (len (toRune .Message.Content)) }}
{{ $advert_rule := (joinStr "" "[#advert_rules](" "https://discordapp.com/channels/" (.Message.GuildID) "/462444993529905172)") }}
{{ $msgKey := (joinStr "" "lastMsg_" (.Message.ChannelID)) }}
{{ $timeKey := (joinStr "_" "lastMsgTime" (.Message.ChannelID) (joinStr "" $lockoutHours "h")) }}
{{ $oldTimeKey := (joinStr "" "lastMsgTime_" (.Message.ChannelID)) }}
{{ $name := .User.Username }}
{{ if .Member.Nick }}{{ $name = .Member.Nick }}{{ end }}

{{- /* ===== 0. ADVERT BAN ===== */ -}}
{{ $ban := dbGet .User.ID "advertBan" }}
{{ if $ban.Value }}
  {{ $remaining := $ban.ExpiresAt.Sub currentTime }}
  {{ if ge (toInt $remaining.Seconds) 0 }}
    {{ sendDM (cembed
      "title" (joinStr "" "Hello " $name "!\n\n" "Your recent post from #" .Channel.Name " was not posted because your advertising privileges are temporarily suspended after four or more infractions within six months. Here is the message that was not posted: ")
      "description" .Message.Content
      "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**Your advertising privileges will be restored on " (printf "<t:%d:F>" (toInt $ban.ExpiresAt.Unix)) ".\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. Please feel free to reach out to a member of the RPC moderation team if you have any further questions.**") "inline" false))
      "color" 14905344
      "author" (sdict "name" "Roleplay Central Database" "icon_url" "https://i.ibb.co/mt5sNFb/Main.png")
      "thumbnail" (sdict "url" "https://i.ibb.co/mt5sNFb/Main.png")
    ) }}
    {{ deleteMessage .Message.ChannelID .Message.ID 0 }}
    {{ return }}
  {{ end }}
{{ end }}

{{- /* ===== 1. LENGTH ===== */ -}}
{{ if gt $argLength $maxLength }}
  {{ sendDM (cembed
    "title" (joinStr "" "Hello " $name "!\n\n" "Your recent post from #" .Channel.Name " was not posted because it exceeds the 2000 character limit for our long-form ad channels. Here is the message that was not posted: ")
    "description" .Message.Content
    "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**" "Please adjust your post to be at or under the max length of a non-Nitro post, which is 2,000 characters. Please note all advertisements in our group channels must be kept to one Discord post, but can include a link to a Google Doc with additional information.\n\nIf you want to keep your post in the current channel, please shorten it to 2000 characters or less. Keep in mind a lot of information may be given using the Post a Plot Tags.\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. Please feel free to reach out to a member of the RPC moderation team if you have any further questions." "**") "inline" false))
    "color" 14905344
    "author" (sdict "name" "Roleplay Central Database" "icon_url" "https://i.ibb.co/mt5sNFb/Main.png")
    "thumbnail" (sdict "url" "https://i.ibb.co/mt5sNFb/Main.png")
  ) }}
  {{ deleteMessage .Message.ChannelID .Message.ID 0 }}
  {{ return }}
{{ end }}

{{- /* ===== 2. COOLDOWN ===== */ -}}
{{ $lastMsgTime := (dbGet .User.ID $timeKey).Value }}
{{ if not $lastMsgTime }}{{ $lastMsgTime = (dbGet .User.ID $oldTimeKey).Value }}{{ end }}
{{ if $lastMsgTime }}
  {{ $minTimeToPost := (currentTime.Add (toDuration (mult $lockoutHours .TimeHour -1))) }}
  {{ $remaining := $lastMsgTime.Sub $minTimeToPost }}
  {{ if ge (toInt $remaining.Seconds) 0 }}
    {{ sendDM (cembed
      "title" (joinStr "" "Hello " $name "!\n\n" "Your recent post from #" .Channel.Name " was not posted because you have posted an advertisement on this channel too recently. Here is the message that was not posted: ")
      "description" .Message.Content
      "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**You are free to wait and post again in " (humanizeDurationMinutes $remaining) ", once your post cooldown has expired. You can check your eligibility to repost in our 'Can I Post' channel.\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. Please feel free to reach out to a member of the RPC moderation team if you have any further questions.**") "inline" false))
      "color" 14905344
      "author" (sdict "name" "Roleplay Central Database" "icon_url" "https://i.ibb.co/mt5sNFb/Main.png")
      "thumbnail" (sdict "url" "https://i.ibb.co/mt5sNFb/Main.png")
    ) }}
    {{ deleteMessage .Message.ChannelID .Message.ID 0 }}
    {{ return }}
  {{ end }}
{{ end }}

{{- /* ===== 3. DUPLICATE IN THIS CHANNEL ===== */ -}}
{{ $lastMsgId := (dbGet .User.ID $msgKey).Value }}
{{ if getMessage .Message.ChannelID $lastMsgId }}
  {{ sendDM (cembed
    "title" (joinStr "" "Hello " $name "!\n\n" "Your recent post from #" .Channel.Name " was not posted because you already have an advertisement on this channel. Here is the message that was not posted: ")
    "description" .Message.Content
    "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**You are free to delete your [old advert](" "https://discordapp.com/channels/" (.Message.GuildID) "/" (.Message.ChannelID) "/" ($lastMsgId) "). Once you have successfully posted a new advert your cooldown period will be restarted. You can check your eligibility to repost in our 'Can I Post' channel.\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. Please feel free to reach out to a member of the RPC moderation team if you have any further questions.**") "inline" false))
    "color" 14905344
    "author" (sdict "name" "Roleplay Central Database" "icon_url" "https://i.ibb.co/mt5sNFb/Main.png")
    "thumbnail" (sdict "url" "https://i.ibb.co/mt5sNFb/Main.png")
  ) }}
  {{ deleteMessage .Message.ChannelID .Message.ID 0 }}
  {{ return }}
{{ end }}

{{- /* ===== POST IS KEPT — record it, then run advisory checks ===== */ -}}
{{ dbSet .User.ID $msgKey (str .Message.ID) }}
{{ dbSet .User.ID $timeKey .Message.Timestamp.Parse }}

{{ $issues := cslice }}

{{- /* --- ADVISORY: group header rules (one header line, within its cap) --- */ -}}
{{ $headers := cslice }}
{{ range (split .Message.Content "\n") }}
  {{ if or (hasPrefix . "# ") (hasPrefix . "## ") (hasPrefix . "### ") }}
    {{ $headers = $headers.Append . }}
  {{ end }}
{{ end }}
{{ $tooMany := gt (len $headers) 1 }}
{{ $tooLong := false }}
{{ range $headers }}
  {{ $line := . }}
  {{ $cap := 50 }}{{ $prefixLen := 2 }}
  {{ if hasPrefix $line "### " }}{{ $cap = 70 }}{{ $prefixLen = 4 }}
  {{ else if hasPrefix $line "## " }}{{ $cap = 60 }}{{ $prefixLen = 3 }}{{ end }}
  {{ if gt (len (toRune (slice $line $prefixLen))) $cap }}{{ $tooLong = true }}{{ end }}
{{ end }}
{{ if or $tooMany $tooLong }}
  {{ $issues = $issues.Append "Group adverts may only have **one** line of header text. Use regular **bold** for any additional lines." }}
{{ end }}

{{- /* --- ADVISORY: banned words (whole word, case-insensitive) --- */ -}}
{{ if gt (len $banned) 0 }}
  {{ $escaped := cslice }}
  {{ range $banned }}{{ $escaped = $escaped.Append (reQuoteMeta .) }}{{ end }}
  {{ $hits := reFindAll (printf "(?i)\\b(?:%s)\\b" (joinStr "|" $escaped)) .Message.Content }}
  {{ if gt (len $hits) 0 }}
    {{ $seen := sdict }}
    {{ $spoilered := cslice }}
    {{ range $hits }}
      {{ $k := lower . }}
      {{ if not ($seen.Get $k) }}{{ $seen.Set $k true }}{{ $spoilered = $spoilered.Append (printf "||%s||" .) }}{{ end }}
    {{ end }}
    {{ $issues = $issues.Append (printf "It contains wording that isn't allowed here: %s" (joinStr " " $spoilered)) }}
  {{ end }}
{{ end }}

{{- /* --- ADVISORY: cross-channel duplicate (normalized exact match) --- */ -}}
{{ $thisNorm := lower .Message.Content }}
{{ $thisNorm = reReplace "(?i)https?://\\S+" $thisNorm " " }}
{{ $thisNorm = reReplace "[^a-z0-9 ]+" $thisNorm " " }}
{{ $thisNorm = reReplace "\\s+" $thisNorm " " }}
{{ $thisNorm = trimSpace $thisNorm }}
{{ if ge (len $thisNorm) 15 }}
  {{ $thisChannel := str .Channel.ID }}
  {{ $dupChannel := "" }}
  {{ range (dbGetPattern .User.ID "lastMsg_%" 100 0) }}
    {{ if and (not $dupChannel) (hasPrefix .Key "lastMsg_") }}
      {{ $cid := slice .Key 8 }}
      {{ if ne $cid $thisChannel }}
        {{ $other := getMessage $cid .Value }}
        {{ if $other }}
          {{ $o := lower $other.Content }}
          {{ $o = reReplace "(?i)https?://\\S+" $o " " }}
          {{ $o = reReplace "[^a-z0-9 ]+" $o " " }}
          {{ $o = reReplace "\\s+" $o " " }}
          {{ $o = trimSpace $o }}
          {{ if eq $o $thisNorm }}{{ $dupChannel = $cid }}{{ end }}
        {{ end }}
      {{ end }}
    {{ end }}
  {{ end }}
  {{ if $dupChannel }}
    {{ $issues = $issues.Append (printf "It looks identical to your ad in <#%s>. Adverts in different channels must be distinctly different from one another." $dupChannel) }}
  {{ end }}
{{ end }}

{{- /* ===== ONE combined advisory ping (only if there were hits) ===== */ -}}
{{ if gt (len $issues) 0 }}
  {{ $body := "" }}
  {{ range $issues }}{{ $body = joinStr "" $body "\n• " . }}{{ end }}

  {{- /* infraction count — prune to 6-month window, append now */ -}}
  {{ $cutoff := (add (toInt currentTime.Unix) (mult $infrWindowSecs -1)) }}
  {{ $dates := cslice }}
  {{ $prev := (dbGet .User.ID "infractionDates").Value }}
  {{ if $prev }}{{ range $prev }}{{ if ge (toInt .) $cutoff }}{{ $dates = $dates.Append (toInt .) }}{{ end }}{{ end }}{{ end }}
  {{ $dates = $dates.Append (toInt currentTime.Unix) }}
  {{ $count := len $dates }}
  {{ dbSet .User.ID "infractionDates" $dates }}

  {{- /* escalation line in the ping */ -}}
  {{ $suffix := "" }}
  {{ if eq $count 3 }}
    {{ $suffix = "\n\n⚠️ **This is your 3rd infraction within six months.** Further infractions will result in a temporary suspension of your advertising privileges." }}
  {{ else if ge $count 4 }}
    {{ $suffix = "\n\n⛔ **This is your 4th infraction within six months.** Your advertising privileges have been suspended for 14 days." }}
  {{ end }}

  {{ $pingID := sendMessageRetID $infractionsChannel (printf "Hey %s ! A few things to fix in your post in %s:%s\n\nPlease edit your post. Thanks!%s" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID) $body $suffix) }}

  {{- /* 4th: 14-day advert ban + wipe history + bot-spam alert */ -}}
  {{ if ge $count 4 }}
    {{ dbSetExpire .User.ID "advertBan" (toInt currentTime.Unix) $advertBanSecs }}
    {{ dbDel .User.ID "infractionDates" }}
    {{ if $botSpam }}{{ sendMessage $botSpam (printf "⛔ <@%d> just hit their **4th infraction in 6 months** (latest in %s) and has been suspended from posting adverts for 14 days." .User.ID (printf "<#%d>" .Channel.ID)) }}{{ end }}
  {{ end }}
  {{/* execCC re-sticks under this ping and hands the post to the sticky to
       start the :staffpending: re-check chain. */}}
  {{ if $infractionsSticky }}{{ execCC $infractionsSticky $infractionsChannel 1 (sdict
    "stickyChannel"  $infractionsChannel
    "recheckMsgID"   .Message.ID
    "recheckChannel" .Channel.ID
    "recheckType"    "group"
    "infractionMsgID" $pingID
  ) }}{{ end }}
  {{ if $staffPending }}{{ addReactions $staffPending }}{{ end }}
  {{ end }}
