{{/* #rule_infractions sticky — keeps the sticky at the BOTTOM of the channel,
     even under the bot's auto-posted infractions.
     Adapted from the BlackWolf sticky (MIT, https://github.com/BlackWolfWoof/yagpdb-cc/).

     YAGPDB triggers don't fire on the bot's own messages, so the `.*` trigger
     can't re-stick under a bot ping; the advert commands + reaction_check call
     this via execCC after they ping. It still runs on `.*` for human posts.
     Trigger: Regex `.*`, restricted to #rule_infractions. See setup.txt. */}}

{{/* Re-stick channel: own channel on the `.*` trigger (.ExecData nil), else
     .ExecData.stickyChannel when called via execCC. */}}
{{ $ch := .Channel.ID }}
{{ if .ExecData }}{{ $ch = .ExecData.stickyChannel }}{{ end }}

{{/* ▼▼ "Infraction Re-check" command id; 0 to disable rechecking. Uses this
       command's otherwise-free execCC/scheduleUniqueCC budget. ▼▼ */}}
{{ $recheckCC := 0 }}

{{/* ▼▼ Bot-spam channel ID (4th-infraction ban alert). 0 to skip. This `.*`
       branch counts only MANUAL (human-typed) infractions; bot pings are
       counted by the advert commands. ▼▼ */}}
{{ $botSpam := 0 }}
{{ $infrWindowSecs := 15552000 }}{{/* 180 days */}}
{{ $advertBanSecs := 1209600 }}{{/* 14 days */}}

{{/* ▼▼ Your existing #rule_infractions sticky text goes here. Keep whatever
       embed your current sticky uses — only this $message line is yours. ▼▼ */}}
{{ $message := cembed "description" "Don't panic! If you see your name here, you're not in trouble! There's a mistake in your advert that needs updated. You have ~10 hours to edit your advert before it's deleted." "color" 0xF4700F }}

{{/* Manual infraction counting (`.*` trigger, .ExecData nil). Bot pings don't
     fire this, so the advert commands count those — no double-count. Can't edit
     a human's message, so escalation posts as a follow-up line. */}}
{{ if not .ExecData }}{{ if gt (len .Message.Mentions) 0 }}
  {{ $u := index .Message.Mentions 0 }}
  {{ $cutoff := (add (toInt currentTime.Unix) (mult $infrWindowSecs -1)) }}
  {{ $dates := cslice }}
  {{ $prev := (dbGet $u.ID "infractionDates").Value }}
  {{ if $prev }}{{ range $prev }}{{ if ge (toInt .) $cutoff }}{{ $dates = $dates.Append (toInt .) }}{{ end }}{{ end }}{{ end }}
  {{ $dates = $dates.Append (toInt currentTime.Unix) }}
  {{ $count := len $dates }}
  {{ dbSet $u.ID "infractionDates" $dates }}

  {{ if eq $count 3 }}
    {{ sendMessage $ch (printf "⚠️ <@%d> — **this is your 3rd infraction within six months.** Further infractions will result in a temporary suspension of your advertising privileges." $u.ID) }}
  {{ else if ge $count 4 }}
    {{/* 4th: notice + 14-day ban + wipe history + bot-spam alert */}}
    {{ sendMessage $ch (printf "⛔ <@%d> — **this is your 4th infraction within six months.** Your advertising privileges have been suspended for 14 days." $u.ID) }}
    {{ dbSetExpire $u.ID "advertBan" (toInt currentTime.Unix) $advertBanSecs }}
    {{ dbDel $u.ID "infractionDates" }}
    {{ if $botSpam }}{{ sendMessage $botSpam (printf "⛔ <@%d> just hit their **4th infraction in 6 months** and has been suspended from posting adverts for 14 days." $u.ID) }}{{ end }}
  {{ end }}
{{ end }}{{ end }}

{{/* do not edit below — $ch passed explicitly so it works on either trigger */}}
{{ if $db := dbGet $ch "stickymessage" }}
	{{ deleteMessage $ch (toInt $db.Value) 0 }}
{{ end }}
{{ $id := sendMessageRetID $ch $message }}
{{ dbSet $ch "stickymessage" (str $id) }}

{{/* If an advert command passed an infracted post, start its re-check chain
     (stage 1, +10 min). Nested ifs: template `and` doesn't short-circuit. */}}
{{ if $recheckCC }}{{ if .ExecData }}{{ if .ExecData.recheckMsgID }}
  {{ scheduleUniqueCC $recheckCC nil 600 (joinStr "" "infr_" .ExecData.recheckMsgID) (sdict
    "msgID"             .ExecData.recheckMsgID
    "channelID"         .ExecData.recheckChannel
    "type"              .ExecData.recheckType
    "stage"             1
    "infractionChannel" $ch
    "infractionMsgID"   .ExecData.infractionMsgID
  ) }}
{{ end }}{{ end }}{{ end }}
