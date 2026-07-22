{{/* =====================================================================
     TEMP TEST DRIVER for the newbie_gate sweep — DELETE AFTER TESTING.

     Trigger type: Command.  Command name: lifetest.

     The sweep normally runs on an interval over the `gatePending` list. This
     injects a pending row for YOU with a join time you choose, resets the sweep
     cursor to the top, and runs the sweep once — so you can watch each branch in
     seconds with the REAL timers instead of waiting 24h/7d. Which branch you hit
     depends on the roles you're CURRENTLY holding:
        • a real role (+ newbie)                     → newbie tag removed (graduated)
        • newbie only                                → nothing until the fall-off age
        • age-please / rules-please only (no newbie)  → KICKED once "age" > 24h

     Optional arg = how many HOURS ago to pretend you joined:
       -lifetest        → joined just now
       -lifetest 25     → joined 25h ago → a gate-stuck member (no newbie) is KICKED
       -lifetest 200    → joined 200h ago (> 7d) → newbie tag falls off
     Run the kick case from a low-role test account (you can't kick yourself if you
     own the server). Best done on a throwaway TEST SERVER.
     ===================================================================== */}}

{{/* ▼▼ the newbie_gate (sweep) command's id ▼▼ */}}
{{ $sweepCC := 0 }}

{{ $hours := 0 }}
{{ if ge (len .CmdArgs) 1 }}{{ $hours = toInt (index .CmdArgs 0) }}{{ end }}

{{- /* inject a pending row for the runner, and reset the cursor so the sweep
       examines from the top (guarantees it sees this row on a small test server) */ -}}
{{ dbSet .User.ID "gatePending" (sub (toInt currentTime.Unix) (mul $hours 3600)) }}
{{ dbSet 0 "gateSweepCursor" 0 }}
{{ execCC $sweepCC nil 1 nil }}

Injected a **gatePending** row for <@{{ .User.ID }}> as if joined **{{ $hours }}h** ago and ran the sweep. Watch your roles (and whether you get kicked).
