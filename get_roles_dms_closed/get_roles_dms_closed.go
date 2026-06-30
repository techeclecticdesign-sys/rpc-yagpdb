{{/* =====================================================================
     GET-ROLES "DMs CLOSED" ENFORCEMENT

     Trigger type: Reaction (added)   — restrict to the #get_roles channel.

     Watches ONE status message in #get_roles. The three valid status
     reactions on it are:
        :mailbox_open:           → "looking | dms open"
        :mailbox_with_no_mail:   → "neutral | advert only"
        :mailbox_closed:         → "not looking | dms closed"
     Role assignment for all three is handled by YAGPDB's native reaction
     roles. This command only enforces the :mailbox_closed: ("not looking")
     case — it does nothing for the other two.

     STAGE 1 (reaction add):
       When a member reacts :mailbox_closed: on the watched message AND they
       did NOT already hold the "not looking | dms closed" role AND they still
       have at least one advert live in an advert channel (same check the
       advert commands use: lastMsg_<channel> -> getMessage), DM them a 30-min
       warning and schedule STAGE 2 for that member.

     STAGE 2 (scheduled recheck, runs via scheduleUniqueCC ~30 min later):
       Re-read the member's CURRENT roles. If they are STILL "not looking",
       delete every advert of theirs that is still posted. If they switched
       back to neutral / looking, do nothing.

     Whether this runs as STAGE 1 or STAGE 2 is decided by .ExecData: it is
     set only on the scheduled call. The scheduleUniqueCC key "rolegrace_<uid>"
     means one pending recheck per member (re-reacting just re-uses it).
     See setup.txt for dashboard instructions.
     ===================================================================== */}}

{{- /* ===== CONFIG ===== */ -}}
{{/* ▼▼ The status message ID in #get_roles that carries the mailbox reactions. ▼▼ */}}
{{ $watchMessageID := "1514068735172939858" }}

{{/* ▼▼ Role ID of "not looking | dms closed". ▼▼ */}}
{{ $notLookingRole := "400451648817987584" }}

{{/* ▼▼ THIS command's OWN id (from the page URL after you first save it).
       Required — without it STAGE 2 never gets scheduled, so nothing is
       ever deleted. ▼▼ */}}
{{ $selfCC := 62 }}

{{/* ▼▼ #get_roles channel ID — only used to make a clickable link in the DM.
       0 falls back to plain "#get_roles" text. ▼▼ */}}
{{ $getRolesChannel := 670079832746622996 }}

{{/* ▼▼ The :mailbox_closed: emoji. For a standard (unicode) emoji this is the
       glyph itself; for a custom server emoji use its NAME with no colons.
       Type \:mailbox_closed: in Discord if you need to confirm which glyph. ▼▼ */}}
{{ $mailboxClosed := "📪" }}

{{/* ▼▼ Channel members can ask staff in — used in the DM footer. ▼▼ */}}
{{ $askTheStaff := 324571668569915393 }}

{{/* Grace period before the recheck, in seconds. */}}
{{ $graceSeconds := 1800 }}{{/* 30 minutes */}}

{{- /* Shared embed boilerplate (matches the advert command DMs). */ -}}
{{ $icon := "https://i.ibb.co/mt5sNFb/Main.png" }}
{{ $author := (sdict "name" "Roleplay Central Database" "icon_url" $icon) }}
{{ $thumb := (sdict "url" $icon) }}
{{ $footer := (joinStr "" "If you have any further questions please feel free to ask on " (printf "<#%d>" $askTheStaff) ".") }}

{{- /* =====================================================================
       STAGE 2 — scheduled recheck (.ExecData is set only on this path)
       ===================================================================== */ -}}
{{ $d := .ExecData }}
{{ if $d }}
  {{ $m := getMember $d.userID }}
  {{ if not $m }}{{ return }}{{ end }}

  {{/* still "not looking"? */}}
  {{ $stillClosed := false }}
  {{ range $m.Roles }}{{ if eq (str .) $notLookingRole }}{{ $stillClosed = true }}{{ end }}{{ end }}
  {{ if not $stillClosed }}{{ return }}{{ end }}

  {{/* delete every advert that is still posted (DB left untouched, exactly
       like the advert dupe-check: a deleted post just stops resolving). */}}
  {{ range (dbGetPattern $d.userID "lastMsg_%" 100 0) }}
    {{ if hasPrefix .Key "lastMsg_" }}
      {{ $cid := slice .Key 8 }}
      {{ if getMessage $cid .Value }}
        {{ deleteMessage $cid .Value 0 }}
      {{ end }}
    {{ end }}
  {{ end }}
  {{ return }}
{{ end }}

{{- /* =====================================================================
       STAGE 1 — reaction added on the watched #get_roles message
       ===================================================================== */ -}}

{{- /* only reaction ADDS, only the watched message, only :mailbox_closed: */ -}}
{{ if not .ReactionAdded }}{{ return }}{{ end }}
{{ if ne (str .Reaction.MessageID) $watchMessageID }}{{ return }}{{ end }}
{{ if ne .Reaction.Emoji.Name $mailboxClosed }}{{ return }}{{ end }}

{{- /* did they ALREADY hold "not looking"? if so this isn't a new transition. */ -}}
{{ $hadRole := false }}
{{ range .Member.Roles }}{{ if eq (str .) $notLookingRole }}{{ $hadRole = true }}{{ end }}{{ end }}
{{ if $hadRole }}{{ return }}{{ end }}

{{- /* do they still have an advert live anywhere? (same source of truth the
       advert commands use) */ -}}
{{ $hasAd := false }}
{{ range (dbGetPattern .User.ID "lastMsg_%" 100 0) }}
  {{ if and (not $hasAd) (hasPrefix .Key "lastMsg_") }}
    {{ $cid := slice .Key 8 }}
    {{ if getMessage $cid .Value }}{{ $hasAd = true }}{{ end }}
  {{ end }}
{{ end }}
{{ if not $hasAd }}{{ return }}{{ end }}

{{- /* one warning + one pending recheck per grace window (dedupes a member who
       toggles the reaction off/on a few times). */ -}}
{{ if (dbGet .User.ID "roleGracePending").Value }}{{ return }}{{ end }}
{{ dbSetExpire .User.ID "roleGracePending" 1 $graceSeconds }}

{{ $name := .User.Username }}
{{ if .Member.Nick }}{{ $name = .Member.Nick }}{{ end }}
{{ $getRolesText := "#get_roles" }}
{{ if $getRolesChannel }}{{ $getRolesText = printf "<#%d>" $getRolesChannel }}{{ end }}

{{ sendDM (cembed
  "title" (joinStr "" "Hello " $name "!\n\n" "You just set your status to **not looking | dms closed**, but you still have one or more advertisements posted in our advert channels.")
  "fields" (cslice (sdict
    "name" "**What happens now?**"
    "value" (joinStr ""
      "**You have 30 minutes to decide. If you'd like to keep your advert(s) up, change your status in " $getRolesText " to either **neutral | advert only** or **looking | dms open**.\n\n"
      "If your status is still **not looking | dms closed** after 30 minutes, your active advert(s) will be removed automatically.\n\n"
      $footer "**")
    "inline" false))
  "color" 14905344
  "author" $author
  "thumbnail" $thumb
) }}

{{- /* schedule the recheck (this command's one execCC/scheduleUniqueCC per run). */ -}}
{{ if $selfCC }}
  {{ scheduleUniqueCC $selfCC nil $graceSeconds (joinStr "" "rolegrace_" .User.ID) (sdict
    "userID" .User.ID
  ) }}
{{ end }}
