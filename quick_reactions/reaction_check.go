{{/* =====================================================================
     Runs @delaySeconds after a post is made. Counts the unique approved
     reactions on the post; if there are fewer than 3, pings the author in
     #rule_infractions.
     See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{ $data := .ExecData }}
{{ if not $data }}{{ return }}{{ end }}

{{/* ▼▼ Paste your #rule_infractions channel ID here (see setup.txt) ▼▼ */}}
{{ $infractionsChannel := 641835326314381312 }}

{{/* ▼▼ Paste your #advert_rules channel ID here (see setup.txt) ▼▼ */}}
{{ $advertRules := 462444993529905172 }}

{{/* ▼▼ Paste your "#rule_infractions sticky" command ID here (see setup.txt) ▼▼ */}}
{{ $infractionsSticky := 60 }}

{{/* ▼▼ The :staffpending: emoji as name:id — MUST match the advert commands and
       the Infraction Re-check command, or the re-check won't be able to remove
       it. Leave "" to skip flagging + auto-resolve for reaction infractions. ▼▼ */}}
{{ $staffPending := "staffpending:1442331141771366513" }}

{{/* ▼▼ Bot-spam channel ID (4th-infraction ban alert). 0 to skip. ▼▼ */}}
{{ $botSpam := 406618336508510209 }}
{{ $infrWindowSecs := 15552000 }}{{/* 180 days */}}
{{ $advertBanSecs := 1209600 }}{{/* 14 days */}}

{{/* Approved roleplay reactions. For custom emojis use the emoji NAME with
     no colons; names are CASE-SENSITIVE and must match the server emojis. */}}
{{ $approved := cslice
  "❤️"
  "18_minor" "18_plus" "21_plus" "25_plus"
  "sfw" "platonic" "nsfw" "romantic" "gm_style"
  "pairing_mxf" "pairing_mxm" "pairing_fxf" "pairing_MxO" "pairing_FxO" "pairing_OxO" "pairing_axa"
  "nonbinary_friendly" "trans_friendly" "nonhuman_friendly"
  "length_oneliner" "length_1para" "length_twotofive" "length_fivetoten" "length_10plus"
  "genre_crime" "genre_cyberpunk" "genre_fantasy" "genre_historical" "genre_horror"
  "genre_modern" "genre_postapoc" "genre_sciencefiction" "genre_sliceoflife" "genre_supernatural"
  "speed_rapidfire" "speed_daily" "speed_weekly" "speed_monthly"
  "original_chars" "canon_chars"
}}

{{/* Re-fetch the post so we read its CURRENT reactions. If it was deleted, stop. */}}
{{ $msg := getMessage $data.channelID $data.msgID }}
{{ if not $msg }}{{ return }}{{ end }}

{{/* Each entry in .Reactions is one unique emoji, so counting the approved
     ones gives the number of unique approved reactions. */}}
{{ $count := 0 }}
{{ range $msg.Reactions }}
  {{ if in $approved .Emoji.Name }}
    {{ $count = add $count 1 }}
  {{ end }}
{{ end }}

{{ if lt $count 3 }}
  {{/* Use a clickable channel link if an ID was set, otherwise plain text. */}}
  {{ $advertRulesText := "#advert_rules" }}
  {{ if $advertRules }}{{ $advertRulesText = printf "<#%d>" $advertRules }}{{ end }}

  {{- /* infraction count — prune to 6-month window, append now (post author) */ -}}
  {{ $uid := $msg.Author.ID }}
  {{ $cutoff := (add (toInt currentTime.Unix) (mult $infrWindowSecs -1)) }}
  {{ $dates := cslice }}
  {{ $prev := (dbGet $uid "infractionDates").Value }}
  {{ if $prev }}{{ range $prev }}{{ if ge (toInt .) $cutoff }}{{ $dates = $dates.Append (toInt .) }}{{ end }}{{ end }}{{ end }}
  {{ $dates = $dates.Append (toInt currentTime.Unix) }}
  {{ $infrCount := len $dates }}
  {{ dbSet $uid "infractionDates" $dates }}

  {{ $suffix := "" }}
  {{ if eq $infrCount 3 }}
    {{ $suffix = "\n\n⚠️ **This is your 3rd infraction within six months.** Further infractions will result in a temporary suspension of your advertising privileges." }}
  {{ else if ge $infrCount 4 }}
    {{ $suffix = "\n\n⛔ **This is your 4th infraction within six months.** Your advertising privileges have been suspended for 14 days." }}
  {{ end }}

  {{ $pingID := sendMessageRetID $infractionsChannel (printf "Hello %s !  Your post in %s needs at least three unique reactions from the list of roleplay reactions viewable in %s.\n\nPlease add these or your post may be removed. Thank you!%s" $data.userMention $data.channelMention $advertRulesText $suffix) }}

  {{- /* 4th: 14-day advert ban + wipe history + bot-spam alert */ -}}
  {{ if ge $infrCount 4 }}
    {{ dbSetExpire $uid "advertBan" (toInt currentTime.Unix) $advertBanSecs }}
    {{ dbDel $uid "infractionDates" }}
    {{ if $botSpam }}{{ sendMessage $botSpam (printf "⛔ <@%d> just hit their **4th infraction in 6 months** (latest in %s) and has been suspended from posting adverts for 14 days." $uid $data.channelMention) }}{{ end }}
  {{ end }}

  {{/* Flag the post for staff follow-up, same emoji the advert commands use. */}}
  {{ if $staffPending }}{{ addMessageReactions $data.channelID $data.msgID $staffPending }}{{ end }}

  {{/* Re-stick the #rule_infractions sticky beneath the ping we just posted
       (its own `.*` trigger never fires on this bot message), AND — by passing
       the recheck params — hand the post off to start the :staffpending:
       re-check chain, exactly like the advert commands do. The sticky spends
       its own budget on the scheduleUniqueCC, so we stay within our one
       execCC. The re-check now understands the reaction floor, so it will
       auto-clear :staffpending: and add :staffapproved: once the post reaches
       three reactions (or delete it at the 8h terminal stage if it never does). */}}
  {{ execCC $infractionsSticky $infractionsChannel 1 (sdict
    "stickyChannel"   $infractionsChannel
    "recheckMsgID"    $data.msgID
    "recheckChannel"  $data.channelID
    "recheckType"     "quick"
    "infractionMsgID" $pingID
  ) }}
{{ end }}
