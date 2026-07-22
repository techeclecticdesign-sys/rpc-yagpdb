{{- /* =====================================================================
     NEWBIE GATE — JOIN HOOK  (goes in the server JOIN message)

     YAGPDB has no member-join custom-command trigger, so — exactly like the leave
     cleanup rides the LEAVE message — the JOIN message is the hook. It runs the
     full custom-command template set, so it can write to the database. All it does
     is add the member to the sweep's pending list: one row `gatePending`, owner =
     the member's id, value = their join unix time. The interval sweep (newbie_gate)
     reads that list. This block prints nothing; keep any welcome text BELOW it.

     Install: Core / General settings -> Join message. Paste at the top, enable
     "Join message", and set a Join message CHANNEL (YAGPDB only runs the template
     when a channel is set; nothing is posted).

     Leaving/kicking clears the row automatically (the leave cleanup wipes the
     member's rows, and the sweep drops any row whose member is gone), so nothing
     else is needed here.
     ===================================================================== */ -}}

{{- /* ===== CONFIG ===== */ -}}
{{/* ▼▼ OPTIONAL — assign the age-please role on join from here. Leave
       $assignAgePlease false if YAGPDB Autorole (or anything else) already assigns
       age-please, or you'll double up. To use it, set BOTH values. ▼▼ */}}
{{ $agePleaseRole := "0" }}
{{ $assignAgePlease := false }}

{{- /* ---- silent; nothing runs for bots ---- */ -}}
{{ if not .User.Bot }}
  {{ if and $assignAgePlease (ne $agePleaseRole "0") }}
    {{ giveRoleID .User.ID $agePleaseRole }}
  {{ end }}
  {{ dbSet .User.ID "gatePending" (toInt currentTime.Unix) }}
{{ end }}
