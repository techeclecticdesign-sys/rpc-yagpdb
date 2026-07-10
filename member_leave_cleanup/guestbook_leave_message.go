{{- /* =====================================================================
     MEMBER LEAVE CLEANUP  —  goes in the guestbook LEAVE MESSAGE
     (Core → General settings → Leave message, with the guestbook set as the
     leave-message channel).

     There is no member-leave custom-command trigger, so the leave message is
     the hook. It runs the full custom-command template function set (YAGPDB
     docs: "custom command templates are supported for usage in all the
     notification feeds"), so it can delete messages and touch the database
     directly — no interval command or execCC needed.

     When a member leaves it:
       1. Deletes every post it has on record for them — their member intro and
          each advertisement (one per channel), all stored as lastMsg_<chID>.
       2. Wipes ALL their stored DB rows so no stale state is left behind.

     It only ever touches the leaving member's own posts/rows. This block prints
     nothing, so any farewell text you post in the guestbook is unaffected —
     keep it BELOW this block.
     ===================================================================== */ -}}

{{- /* 1. Delete each recorded post. Records are keyed lastMsg_<channelID> with
       the message ID as the value. SQL LIKE treats "_" as a single-char
       wildcard, so "lastMsg%" also returns lastMsgTime_* cooldown rows — we
       filter to the exact first segment "lastMsg" so we only ever delete real
       posts, never treat a timestamp as a message ID. deleteMessage on an
       already-removed message just no-ops, so no existence check is needed. */ -}}
{{ range (dbGetPattern .User.ID "lastMsg%" 100 0) }}
  {{ $parts := split .Key "_" }}
  {{ if eq (index $parts 0) "lastMsg" }}
    {{ deleteMessage (index $parts 1) .Value 0 }}
  {{ end }}
{{ end }}

{{- /* 2. Wipe every DB row this user owns: infractionLog, infractionDates,
       advertBan, lastMsg_*, lastMsgTime_*, and anything added later.
       Scoped to their user ID, so channel stickies (keyed by channel ID) are
       never matched. Return value discarded so nothing renders. */ -}}
{{ $_ := dbDelMultiple (sdict "userID" .User.ID "pattern" "%") 100 0 }}
