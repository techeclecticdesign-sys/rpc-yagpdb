{{/* =====================================================================
     #rule_infractions sticky — keeps the sticky pinned to the BOTTOM of
     #rule_infractions, even underneath the bot's auto-posted infractions.

     Adapted from the BlackWolf sticky (MIT, https://github.com/BlackWolfWoof/yagpdb-cc/).

     WHY THIS IS DIFFERENT FROM THE PLAIN STICKY
     -------------------------------------------
     A normal sticky re-posts itself on its `.*` regex trigger, which fires on
     every NEW message in the channel. That works while humans post the
     infractions by hand. But YAGPDB custom-command triggers NEVER fire on the
     bot's own messages, so once the auto-infraction commands started posting
     the pings, the `.*` trigger stopped firing and the sticky got buried above
     them.

     Fix: the auto-infraction senders (alert_sender, reaction_check) call THIS
     command via execCC right after they post a ping, so it re-sticks below it.
     It still also runs on the normal `.*` trigger for human posts.

     Trigger type: Regex
     Trigger:      .*
     Channel:      restrict to #rule_infractions
     See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{/* Which channel to re-stick in:
       • Normal `.*` trigger (a human posted) → .ExecData is nil, so we use the
         channel the message was posted in (this command's own channel).
       • Called via execCC by an infraction sender → it passes the channel in
         .ExecData.stickyChannel (so we don't depend on the trigger context). */}}
{{ $ch := .Channel.ID }}
{{ if .ExecData }}{{ $ch = .ExecData.stickyChannel }}{{ end }}

{{/* ▼▼ Your existing #rule_infractions sticky text goes here. Keep whatever
       embed your current sticky uses — only this $message line is yours. ▼▼ */}}
{{ $message := cembed "description" "Don't panic! If you see your name here, you're not in trouble! There's a mistake in your advert that needs updated. You have ~10 hours to edit your advert before it's deleted." "color" 0xF4700F }}

{{/* do not edit below — note we pass $ch explicitly instead of nil, so this
     works whether we were triggered normally or via execCC from another channel */}}
{{ if $db := dbGet $ch "stickymessage" }}
	{{ deleteMessage $ch (toInt $db.Value) 0 }}
{{ end }}
{{ $id := sendMessageRetID $ch $message }}
{{ dbSet $ch "stickymessage" (str $id) }}
