{{/* =====================================================================
     RESET RULES-ACCEPT REACTIONS  (keep only Jarvis')

     Trigger type: Command   — restrict to staff.

     The #server_rules accept message carries the :rpc: reaction that
     drives the native reaction-role. Over time members' own :rpc:
     reactions pile up underneath the bot's. This command clears EVERY
     :rpc: reaction on that message and immediately re-adds the bot's
     own, so the only reaction left is Jarvis' — the reaction-role stays
     intact and no already-assigned roles are touched.

     Why it's safe: deleteAllMessageReactions fires a bulk REMOVE_EMOJI
     event, and YAGPDB's role system only reacts to individual
     MessageReactionAdd / MessageReactionRemove events — never to bulk
     clears. So members keep their accepted-rules role. Re-adding the
     emoji is done by the bot, restoring the clickable reaction users
     need. Verified on a mirror test server.

     This is the manual equivalent of `-rolemenu resetreactions`, kept
     as a custom command so it works whether or not the accept message
     is a native role menu, and so it can be triggered on demand.
     See setup.txt for dashboard instructions.
     ===================================================================== */}}

{{- /* ===== CONFIG ===== */ -}}
{{/* ▼▼ #server_rules channel ID. ▼▼ */}}
{{ $channelID := "367000571908980738" }}

{{/* ▼▼ The rules-accept message ID in that channel. ▼▼ */}}
{{ $messageID := "1385675552601411605" }}

{{/* ▼▼ The accept emoji in NAME:ID form — no colons or <> around it.
       Custom server emoji :rpc:. ▼▼ */}}
{{ $emoji := "rpc:658510846779195423" }}

{{- /* ===== RUN =====
       1) clear every :rpc: reaction (members' + the bot's)
       2) re-add the bot's own so the reaction-role stays usable */ -}}
{{ deleteAllMessageReactions $channelID $messageID $emoji }}
{{ addMessageReactions $channelID $messageID $emoji }}

Cleared all :rpc: reactions on the rules message and restored the bot's.
