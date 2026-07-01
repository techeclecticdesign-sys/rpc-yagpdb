{{/* =====================================================================
     QUICK ADVERT STICKY  (+ reaction timer rolled in)

     Keeps this quick channel's sticky pinned at the bottom AND arms the
     +5-minute Reaction Check for each new post — folding the old standalone
     "Post Timer" command into the sticky so the quick channel drops from 3
     message-command slots (advert cmd + sticky + Post Timer) to 2
     (advert cmd + sticky). The Reaction Check it schedules is trigger-type
     None, fired by scheduleUniqueCC, so it costs no slot.

     Trigger type: Regex
     Trigger:      (?s).*
     Channel:      restrict to ONE quick channel (e.g. #quick_fandoms).
                   Each quick channel keeps its own copy of this command — the
                   sticky message is tracked per-channel via the "stickymessage"
                   db key on .Channel.ID, so the two don't clash.

     This replaces BOTH the plain quick-channel sticky AND the Post Timer.
     Delete the Post Timer command after deploying this (frees the 3rd slot).
     Adapted from the BlackWolf sticky (MIT,
     https://github.com/BlackWolfWoof/yagpdb-cc/).
     ===================================================================== */}}

{{/* ▼▼ "Reaction Check" command ID (its trigger type is None). Set to the same
       id the old Post Timer pointed at. 0 disables the timer (sticky still
       works). ▼▼ */}}
{{ $reactionCC := 0 }}

{{/* ▼▼ Seconds to wait before the reaction check. 300 = 5 minutes — same
       default the Post Timer used. ▼▼ */}}
{{ $delaySeconds := 300 }}

{{/* ▼▼ This channel's sticky text. Keep whatever embed your current quick
       sticky uses — only this $message line is yours. ▼▼ */}}
{{ $message := cembed "description" "Quick adverts need at least 3 roleplay reaction tags. Add yours within ~5 minutes or the bot will remind you in #rule_infractions." "color" 0xF4700F }}

{{/* ===== arm the reaction timer for THIS post =====
     Only on a real post (.ExecData nil). YAGPDB's .* trigger never fires on
     the bot's own sticky re-post, so there's no loop and no self-arming. The
     unique key "rxn_<msgID>" gives one timer per post; if the post is deleted
     by a hard rule before it fires, Reaction Check just no-ops (getMessage
     returns nothing). */}}
{{ if $reactionCC }}{{ if not .ExecData }}
  {{ scheduleUniqueCC $reactionCC nil $delaySeconds (joinStr "" "rxn_" .Message.ID) (sdict
    "msgID"          .Message.ID
    "channelID"      .Message.ChannelID
    "userMention"    (printf "<@%d>" .User.ID)
    "channelMention" (printf "<#%d>" .Message.ChannelID)
  ) }}
{{ end }}{{ end }}

{{/* ===== re-stick: delete the old sticky, post a fresh one, remember it =====
     Keyed on .Channel.ID so each quick channel tracks its own sticky message. */}}
{{ $ch := .Channel.ID }}
{{ if $db := dbGet $ch "stickymessage" }}
  {{ deleteMessage $ch (toInt $db.Value) 0 }}
{{ end }}
{{ $id := sendMessageRetID $ch $message }}
{{ dbSet $ch "stickymessage" (str $id) }}
