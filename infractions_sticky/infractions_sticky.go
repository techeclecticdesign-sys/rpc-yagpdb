{{/* #rule_infractions sticky — keeps the sticky at the BOTTOM of the channel,
     even under the bot's auto-posted infractions.
     Adapted from the BlackWolf sticky (MIT, https://github.com/BlackWolfWoof/yagpdb-cc/).

     YAGPDB triggers don't fire on the bot's own messages, so the `.*` trigger
     can't re-stick under a bot ping; the advert commands + reaction_check call
     this via execCC after they ping. It still runs on `.*` for human posts.
     Trigger: Regex `.*`, restricted to #rule_infractions. See setup.txt.

     EXPIRY LEDGER: because this command runs after EVERY #rule_infractions
     message (`.*` for humans, execCC for bot pings — which hand over the ping
     ID as .ExecData.infractionMsgID), it is the one choke point that can
     record the channel's messages for the 180-day expiry sweep
     (post_expiry/infraction_expiry.go). IDs collected during a run are
     appended to a 30-day bucket entry, infrLedger_<unixTime/2592000>, owned by
     the channel ID — a fixed ~7 entries total, never one per infraction. The
     sticky message itself is never ledgered (it deletes + reposts itself). */}}

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

{{/* Message IDs to append to the expiry ledger; written in one dbGet+dbSet at
     the bottom so ledgering costs 2 DB ops no matter how many IDs a run adds. */}}
{{ $ledger := cslice }}

{{/* Bot ping (execCC path): ledger the ping the calling command just posted. */}}
{{ if .ExecData }}{{ if .ExecData.infractionMsgID }}
  {{ $ledger = $ledger.Append (str .ExecData.infractionMsgID) }}
{{ end }}{{ end }}

{{/* Manual infraction counting (`.*` trigger, .ExecData nil). Bot pings don't
     fire this, so the advert commands count those — no double-count. Can't edit
     a human's message, so escalation posts as a follow-up line. */}}
{{ if not .ExecData }}
  {{/* every human message in the channel expires — ledger it, mentions or not */}}
  {{ $ledger = $ledger.Append (str .Message.ID) }}
{{ end }}
{{ if not .ExecData }}{{ if gt (len .Message.Mentions) 0 }}
  {{ $u := index .Message.Mentions 0 }}
  {{- /* infraction log — prune to the 6-month window, migrate any legacy
         plain-timestamp entries, then append this manual record. Reason is
         "manual" and the link points at the staff note itself. History is never
         wiped, so every infraction from the 4th on re-applies the ban. */ -}}
  {{ $cutoff := (add (toInt currentTime.Unix) (mult $infrWindowSecs -1)) }}
  {{ $log := cslice }}
  {{ $legacy := (dbGet $u.ID "infractionDates").Value }}
  {{ if $legacy }}{{ range $legacy }}{{ if ge (toInt .) $cutoff }}{{ $log = $log.Append (sdict "t" (toInt .) "r" "" "c" "" "m" "") }}{{ end }}{{ end }}{{ end }}
  {{ $prevLog := (dbGet $u.ID "infractionLog").Value }}
  {{ if $prevLog }}{{ range $prevLog }}{{ if ge (toInt .t) $cutoff }}{{ $log = $log.Append . }}{{ end }}{{ end }}{{ end }}
  {{ $log = $log.Append (sdict "t" (toInt currentTime.Unix) "r" "manual" "c" (str $ch) "m" (str .Message.ID)) }}
  {{ $count := len $log }}
  {{ dbSet $u.ID "infractionLog" $log }}
  {{ if $legacy }}{{ dbDel $u.ID "infractionDates" }}{{ end }}

  {{ if eq $count 3 }}
    {{ $wid := sendMessageRetID $ch (printf "⚠️ <@%d> — **this is your 3rd infraction within six months.** Further infractions will result in a temporary suspension of your advertising privileges." $u.ID) }}
    {{ $ledger = $ledger.Append (str $wid) }}
  {{ else if ge $count 4 }}
    {{/* 4th and every one after: notice + (re)apply 14-day ban + bot-spam alert. History kept. */}}
    {{ $wid := sendMessageRetID $ch (printf "⛔ <@%d> — **this is infraction #%d within six months.** Your advertising privileges have been suspended for 14 days." $u.ID $count) }}
    {{ $ledger = $ledger.Append (str $wid) }}
    {{ dbSetExpire $u.ID "advertBan" (toInt currentTime.Unix) $advertBanSecs }}
    {{ if $botSpam }}{{ sendMessage $botSpam (printf "⛔ <@%d> now has **%d infractions in 6 months** and has been suspended from posting adverts for 14 days." $u.ID $count) }}{{ end }}
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

{{/* Flush the expiry ledger: append this run's IDs to the current 30-day
     bucket (2592000 must match post_expiry/infraction_expiry.go). Worst-path
     DB ops stay within 10: infractionDates get + legacy del, infractionLog
     get+set, advertBan setExpire, stickymessage get+set, ledger get+set. */}}
{{ if len $ledger }}
  {{ $bkey := printf "infrLedger_%d" (div (toInt currentTime.Unix) 2592000) }}
  {{ $cur := (dbGet $ch $bkey).Value }}
  {{ $list := cslice }}
  {{ if $cur }}{{ $list = $list.AppendSlice $cur }}{{ end }}
  {{ $list = $list.AppendSlice $ledger }}
  {{ dbSet $ch $bkey $list }}
{{ end }}
