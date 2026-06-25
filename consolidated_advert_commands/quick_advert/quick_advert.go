{{ $maxLength := 105 }}
{{ $lockoutHours := 96 }}
{{ $infractionsChannel := 0 }}
{{ $infractionsSticky := 0 }}
{{ $staffPending := "staffpending:1442331141771366513" }}
{{ $banned := cslice
  "futa"
  "futanari"
  "futas"
  "futanaris"
}}

{{ $argLength := (len (split (joinStr " " .Args) " ")) }}
{{ $advert_rule := (joinStr "" "[#advert_rules](" "https://discordapp.com/channels/" (.Message.GuildID) "/462444993529905172)") }}
{{ $msgKey := (joinStr "" "lastMsg_" (.Message.ChannelID)) }}
{{ $timeKey := (joinStr "_" "lastMsgTime" (.Message.ChannelID) (joinStr "" $lockoutHours "h")) }}
{{ $oldTimeKey := (joinStr "" "lastMsgTime_" (.Message.ChannelID)) }}
{{ $name := .User.Username }}
{{ if .Member.Nick }}{{ $name = .Member.Nick }}{{ end }}

{{- /* ===== 1. LENGTH (word count) ===== */ -}}
{{ if ge $argLength $maxLength }}
  {{ $longFormChannel := "" }}
  {{ $channelPlural := "this channel" }}
  {{ if eq .Channel.Name "quick_fandoms" }}
    {{ $longFormChannel = (joinStr "" "[#fandom_adverts](" "https://discordapp.com/channels/" (.Message.GuildID) "/504322252611780609)") }}
  {{ else if eq .Channel.Name "quick_originals" }}
    {{ $longFormChannel = (joinStr "" "[#original_adverts](" "https://discordapp.com/channels/" (.Message.GuildID) "/504322272618610688)") }}
  {{ end }}
  {{ sendDM (cembed
    "title" (joinStr "" "Hello " $name "!\n\n" "Your recent post from #" .Channel.Name " was not posted because it exceeds the hundred word limit for the quick search channels. Here is the message that was not posted: ")
    "description" .Message.Content
    "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**" "If you want to keep the current length of your post please move it to " $longFormChannel ". Please note all advertisements on " $channelPlural " must be kept to one non-Nitro length Discord post, but can include a link to a Google Doc with additional information.\n\nIf you want to keep your post in the current channel, you must shorten it to be at or under 100 words and re-send your ad once it's within that word limit. You can check your eligibility in our 'Can I post' channel. Keep in mind a lot of information may be given using the Quick Reaction Tags.\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. Please feel free to reach out to a member of the RPC moderation team if you have any further questions." "**") "inline" false))
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

{{- /* ===== POST IS KEPT — record it ===== */ -}}
{{ dbSet .User.ID $msgKey (str .Message.ID) }}
{{ dbSet .User.ID $timeKey .Message.Timestamp.Parse }}
{{- /* NOTE: the reaction check is scheduled by the QUICK-CHANNEL STICKY (the
       old post_timer logic, folded in there), NOT here. On non-premium YAGPDB a
       command may make only ONE execCC call per run, and this command spends it
       on the #rule_infractions re-stick at the bottom. The sticky fires on the
       same post and has its own budget, so it carries the reaction-check call. */ -}}

{{ $issues := cslice }}

{{- /* --- ADVISORY: no links in quick channels --- */ -}}
{{ if reFind "(?i)https?://\\S+" .Message.Content }}
  {{ $issues = $issues.Append "Links aren't allowed in the quick search channels. Please remove it from your ad." }}
{{ end }}

{{- /* --- ADVISORY: no images / attachments in quick (quick rule o2) --- */ -}}
{{ if gt (len .Message.Attachments) 0 }}
  {{ $issues = $issues.Append "Images and other media aren't allowed in the quick search channels. Please remove any attachments." }}
{{ end }}

{{- /* --- ADVISORY: no headers in adverts (general rule) --- */ -}}
{{ $hasHeader := false }}
{{ range (split .Message.Content "\n") }}
  {{ if or (hasPrefix . "# ") (hasPrefix . "## ") (hasPrefix . "### ") }}{{ $hasHeader = true }}{{ end }}
{{ end }}
{{ if $hasHeader }}
  {{ $issues = $issues.Append "Headers aren't allowed in the quick search channels. You're welcome to use regular **bold** instead." }}
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
  {{ sendMessage $infractionsChannel (printf "Hey %s ! A few things to fix in your post in %s:%s\n\nPlease edit your post. Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID) $body) }}
  {{/* Re-stick the #rule_infractions sticky beneath the ping we just posted.
       The sticky's own `.*` trigger never fires on this bot message. */}}
  {{ if $infractionsSticky }}{{ execCC $infractionsSticky $infractionsChannel 1 (sdict "stickyChannel" $infractionsChannel) }}{{ end }}
  {{/* Flag the post for staff follow-up. */}}
  {{ if $staffPending }}{{ addReactions $staffPending }}{{ end }}
{{ end }}
