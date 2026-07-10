{{ $maxLength := 2100 }}
{{ $lockoutHours := 96 }}
{{ $infractionsChannel := 641835326314381312 }}
{{ $infractionsSticky := 0 }}
{{ $botSpam := 406618336508510209 }}
{{ $infrWindowSecs := 15552000 }}{{/* 180 days */}}
{{ $advertBanSecs := 1209600 }}{{/* 14 days */}}
{{ $graceSecs := 600 }}{{/* 10-min delete-and-repost grace window */}}
{{ $staffPending := "staffpending:1442331141771366513" }}
{{ $askTheStaff := 324571668569915393 }}
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
      "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**Your advertising privileges will be restored on " (printf "<t:%d:F>" (toInt $ban.ExpiresAt.Unix)) ".\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. If you have any further questions please feel free to ask on " (printf "<#%d>" $askTheStaff) ".**") "inline" false))
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
    "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**" "Please adjust your post to be at or under the max length of a non-Nitro post, which is 2,000 characters. Please note all advertisements in our group channels must be kept to one Discord post, but can include a link to a Google Doc with additional information.\n\nIf you want to keep your post in the current channel, please shorten it to 2000 characters or less. Keep in mind a lot of information may be given using the Post a Plot Tags.\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. If you have any further questions please feel free to ask on " (printf "<#%d>" $askTheStaff) "." "**") "inline" false))
    "color" 14905344
    "author" (sdict "name" "Roleplay Central Database" "icon_url" "https://i.ibb.co/mt5sNFb/Main.png")
    "thumbnail" (sdict "url" "https://i.ibb.co/mt5sNFb/Main.png")
  ) }}
  {{ deleteMessage .Message.ChannelID .Message.ID 0 }}
  {{ return }}
{{ end }}

{{- /* ===== 2. COOLDOWN ===== */ -}}
{{- /* Fetch the recorded ad once: whether it still exists (getMessage) drives
     both the duplicate check below and the delete-and-repost grace window. */ -}}
{{ $lastMsgId := (dbGet .User.ID $msgKey).Value }}
{{ $oldExists := false }}
{{ if $lastMsgId }}{{ if getMessage .Message.ChannelID $lastMsgId }}{{ $oldExists = true }}{{ end }}{{ end }}
{{ $inGrace := false }}
{{ $lastMsgTime := (dbGet .User.ID $timeKey).Value }}
{{ if not $lastMsgTime }}{{ $lastMsgTime = (dbGet .User.ID $oldTimeKey).Value }}{{ end }}
{{ if $lastMsgTime }}
  {{ $minTimeToPost := (currentTime.Add (toDuration (mult $lockoutHours .TimeHour -1))) }}
  {{ $remaining := $lastMsgTime.Sub $minTimeToPost }}
  {{ if ge (toInt $remaining.Seconds) 0 }}
    {{- /* Still on cooldown — but if the recorded ad has been deleted and it was
         posted under $graceSecs ago, let this repost through instead of blocking,
         so a delete-to-fix within the window isn't hard-failed. */ -}}
    {{ $age := currentTime.Sub $lastMsgTime }}
    {{ if and (not $oldExists) (le (toInt $age.Seconds) $graceSecs) }}
      {{ $inGrace = true }}
    {{ else }}
    {{ sendDM (cembed
      "title" (joinStr "" "Hello " $name "!\n\n" "Your recent post from #" .Channel.Name " was not posted because you have posted an advertisement on this channel too recently. Here is the message that was not posted: ")
      "description" .Message.Content
      "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**You are free to wait and post again in " (humanizeDurationMinutes $remaining) ", once your post cooldown has expired. You can check your eligibility to repost in our 'Can I Post' channel.\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. If you have any further questions please feel free to ask on " (printf "<#%d>" $askTheStaff) ".**") "inline" false))
      "color" 14905344
      "author" (sdict "name" "Roleplay Central Database" "icon_url" "https://i.ibb.co/mt5sNFb/Main.png")
      "thumbnail" (sdict "url" "https://i.ibb.co/mt5sNFb/Main.png")
    ) }}
    {{ deleteMessage .Message.ChannelID .Message.ID 0 }}
    {{ return }}
    {{ end }}
  {{ end }}
{{ end }}

{{- /* ===== 3. DUPLICATE IN THIS CHANNEL ===== */ -}}
{{ if $oldExists }}
  {{ sendDM (cembed
    "title" (joinStr "" "Hello " $name "!\n\n" "Your recent post from #" .Channel.Name " was not posted because you already have an advertisement on this channel. Here is the message that was not posted: ")
    "description" .Message.Content
    "fields" (cslice (sdict "name" "**What can you do about this?**" "value" (joinStr "" "**You are free to delete your [old advert](" "https://discordapp.com/channels/" (.Message.GuildID) "/" (.Message.ChannelID) "/" ($lastMsgId) "). Once you have successfully posted a new advert your cooldown period will be restarted. You can check your eligibility to repost in our 'Can I Post' channel.\n\nFor additional information about posting advertisements, please see our " $advert_rule " channel. If you have any further questions please feel free to ask on " (printf "<#%d>" $askTheStaff) ".**") "inline" false))
    "color" 14905344
    "author" (sdict "name" "Roleplay Central Database" "icon_url" "https://i.ibb.co/mt5sNFb/Main.png")
    "thumbnail" (sdict "url" "https://i.ibb.co/mt5sNFb/Main.png")
  ) }}
  {{ deleteMessage .Message.ChannelID .Message.ID 0 }}
  {{ return }}
{{ end }}

{{- /* ===== POST IS KEPT — record it, then run advisory checks ===== */ -}}
{{ dbSet .User.ID $msgKey (str .Message.ID) }}
{{ if $inGrace }}{{ dbSet .User.ID $timeKey $lastMsgTime }}{{ else }}{{ dbSet .User.ID $timeKey .Message.Timestamp.Parse }}{{ end }}

{{ $issues := cslice }}
{{- /* $tags: a short machine tag per issue, kept in lock-step with $issues, so
     the recorded infraction can list what it was for (headers, banned word …). */ -}}
{{ $tags := cslice }}

{{- /* --- ADVISORY: no headers allowed in 1x1 (any header line; a #/##/### line
     with nothing but trailing whitespace after it counts too — Discord renders
     the NEXT line as its heading text) --- */ -}}
{{ $hasHeader := false }}
{{ range (split .Message.Content "\n") }}
  {{ if or (hasPrefix . "# ") (hasPrefix . "## ") (hasPrefix . "### ") (and (hasPrefix . "#") (or (eq (trimSpace .) "#") (eq (trimSpace .) "##") (eq (trimSpace .) "###"))) }}{{ $hasHeader = true }}{{ end }}
{{ end }}
{{ if $hasHeader }}
  {{ $issues = $issues.Append "Headers aren't allowed in the one-on-one advert channels. You're welcome to use regular **bold** instead." }}
  {{ $tags = $tags.Append "headers" }}
{{ end }}

{{- /* --- ADVISORY: no images / attachments in 1x1 --- */ -}}
{{ if gt (len .Message.Attachments) 0 }}
  {{ $issues = $issues.Append "Images and other media aren't allowed in the 1x1 advert channels. Please remove any attachments." }}
  {{ $tags = $tags.Append "image" }}
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
    {{ $tags = $tags.Append "banned word" }}
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
    {{ $issues = $issues.Append (printf "It looks identical to your ad in <#%s>. Cross-channel adverts must be distinctly different from each other and searching for different things. Please choose a channel for your advert." $dupChannel) }}
    {{ $tags = $tags.Append "dupe" }}
  {{ end }}
{{ end }}

{{- /* ===== ONE combined advisory ping (only if there were hits) ===== */ -}}
{{ if gt (len $issues) 0 }}
  {{ $body := "" }}
  {{ range $issues }}{{ $body = joinStr "" $body "\n• " . }}{{ end }}

  {{- /* infraction log — prune to the 6-month window, migrate any legacy
         plain-timestamp entries, then append this post's record. Entries are
         only ever dropped once they age past the window — NO reset on the 4th,
         so the count keeps climbing and every infraction from the 4th on
         re-applies the ban below. Each record is {t,r,c,m}: unix time, a
         comma-joined reason (from $tags), and the post's channel/message id for
         a jump link in /infractions view. */ -}}
  {{ $cutoff := (add (toInt currentTime.Unix) (mult $infrWindowSecs -1)) }}
  {{ $log := cslice }}
  {{ $legacy := (dbGet .User.ID "infractionDates").Value }}
  {{ if $legacy }}{{ range $legacy }}{{ if ge (toInt .) $cutoff }}{{ $log = $log.Append (sdict "t" (toInt .) "r" "" "c" "" "m" "") }}{{ end }}{{ end }}{{ end }}
  {{ $prevLog := (dbGet .User.ID "infractionLog").Value }}
  {{ if $prevLog }}{{ range $prevLog }}{{ if ge (toInt .t) $cutoff }}{{ $log = $log.Append . }}{{ end }}{{ end }}{{ end }}
  {{ $log = $log.Append (sdict "t" (toInt currentTime.Unix) "r" (joinStr ", " $tags) "c" (str .Channel.ID) "m" (str .Message.ID)) }}
  {{ $count := len $log }}
  {{ dbSet .User.ID "infractionLog" $log }}
  {{ if $legacy }}{{ dbDel .User.ID "infractionDates" }}{{ end }}

  {{- /* escalation line in the ping */ -}}
  {{ $suffix := "" }}
  {{ if eq $count 3 }}
    {{ $suffix = "\n\n⚠️ **This is your 3rd infraction within six months.** Further infractions will result in a temporary suspension of your advertising privileges." }}
  {{ else if ge $count 4 }}
    {{ $suffix = printf "\n\n⛔ **This is infraction #%d within six months.** Your advertising privileges have been suspended for 14 days." $count }}
  {{ end }}

  {{ $pingID := sendMessageRetID $infractionsChannel (printf "Hey %s ! A few things to fix in your post in %s:%s\n\nPlease edit your post. Thanks!%s" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID) $body $suffix) }}

  {{- /* 4th and every infraction after: (re)apply the 14-day advert ban +
         bot-spam alert. History is NOT wiped, so each further slip re-mutes for
         another 14 days. */ -}}
  {{ if ge $count 4 }}
    {{ dbSetExpire .User.ID "advertBan" (toInt currentTime.Unix) $advertBanSecs }}
    {{ if $botSpam }}{{ sendMessage $botSpam (printf "⛔ <@%d> now has **%d advert infractions in 6 months** (latest in %s) and has been suspended from posting adverts for 14 days." .User.ID $count (printf "<#%d>" .Channel.ID)) }}{{ end }}
  {{ end }}
  {{/* execCC re-sticks under this ping and hands the post to the sticky to
       start the :staffpending: re-check chain. */}}
  {{ if $infractionsSticky }}{{ execCC $infractionsSticky $infractionsChannel 1 (sdict
    "stickyChannel"  $infractionsChannel
    "recheckMsgID"   .Message.ID
    "recheckChannel" .Channel.ID
    "recheckType"    "1x1"
    "infractionMsgID" $pingID
  ) }}{{ end }}
  {{ if $staffPending }}{{ addReactions $staffPending }}{{ end }}
{{ end }}
