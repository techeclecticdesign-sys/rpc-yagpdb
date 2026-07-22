{{/* =====================================================================
     ONBOARDING REMINDERS — the two daily cohort pings (delete + repost both)

     Trigger type: Minute Interval (set Interval 1440 for once a day),
     run in #getting_started.

     Posts TWO stock messages, deleting the previous copy of each first so there's
     only ever one of each in the channel and every repost re-pings + drops to the
     bottom:
       1. @newbie                         — "pick your roles" (people past the
                                             gates who haven't chosen roles yet)
       2. @age-please + @rules-please      — "you're stuck at the entry gates,
                                             finish within 24h or you're kicked"
     Anyone who has moved on has already lost the pinged role, so each message only
     reaches the people it's for. Costs no message-trigger slots; ~8 DB ops/run.

     The actual kick is done by newbie_gate (24h in the server without the newbie
     role). This command only nudges.

     PREREQ: the bot must be able to mention all three roles — set each role
     "Allow anyone to @mention this role", or give the bot "Mention @everyone,
     @here, and All Roles". (allowed_mentions below already scopes each ping.)
     ===================================================================== */}}

{{- /* ===== CONFIG ===== */ -}}
{{/* ▼▼ #getting_started channel ID — where both messages post (make this the
       command's "run in" channel too). REQUIRED. ▼▼ */}}
{{ $channel := "0" }}

{{/* ▼▼ MESSAGE 1 (pick-your-roles) IDs. ▼▼ */}}
{{ $newbieRole := "456657052136243210" }}{{/* @newbie — the role this message pings */}}
{{ $getRolesChannel := "670079832746622996" }}{{/* #get_roles */}}
{{ $introChannel := "1048992435591729182" }}{{/* where they can drop a message */}}

{{/* ▼▼ MESSAGE 2 (entry-gate) IDs — the two gate tags this message pings. ▼▼ */}}
{{ $agePleaseRole := "735544386678554738" }}
{{ $rulesPleaseRole := "634585076524646401" }}

{{/* ▼▼ The two stock messages. Edit freely; keep the <@&role>/<#channel> refs. ▼▼ */}}
{{ $msg1 := joinStr ""
    "``` ```\n\nHey there, <@&" $newbieRole ">! This is just a reminder to complete the tasks above! "
    "**Most importantly,** pop on over to <#" $getRolesChannel "> to access more of the server "
    "(and stop getting pinged) or drop a message into <#" $introChannel ">!\n\n** **" }}

{{ $msg2 := joinStr ""
    "``` ```\n\nHello! Welcome to RPC, <@&" $agePleaseRole "> and <@&" $rulesPleaseRole ">! "
    "If you've been tagged, that means you're stuck in our entry gates. Please follow the steps above to be added to the server.\n\n"
    "You have 24 hours from joining to comply or you will be kicked from the server. You are welcome to return when you are able to complete our join process.\n\n"
    "If you have already completed these steps but were not pushed along in the process, please contact an admin to be pushed forward manually." }}

{{- /* delete the previous copy of each (no-ops if already gone / first run) */ -}}
{{ if $db := dbGet 0 "newbieReminderMsg" }}{{ deleteMessage $channel (toInt $db.Value) 0 }}{{ end }}
{{ if $db := dbGet 0 "gateReminderMsg" }}{{ deleteMessage $channel (toInt $db.Value) 0 }}{{ end }}

{{- /* repost message 1 (pings only @newbie) and remember its id */ -}}
{{ $id1 := sendMessageRetID $channel (complexMessage
    "content" $msg1
    "allowed_mentions" (sdict "roles" (cslice $newbieRole))) }}
{{ dbSet 0 "newbieReminderMsg" (str $id1) }}

{{- /* repost message 2 (pings only the two gate roles) and remember its id */ -}}
{{ $id2 := sendMessageRetID $channel (complexMessage
    "content" $msg2
    "allowed_mentions" (sdict "roles" (cslice $agePleaseRole $rulesPleaseRole))) }}
{{ dbSet 0 "gateReminderMsg" (str $id2) }}
