{{/* =====================================================================
     ban -- prefix alias for YAGPDB's built-in Ban command.

     Trigger type: Command.   Name: ban

       >>ban 1444117608000520232 Ban evasion, _mrenigma
       >>ban @member 7d Spamming -ddays 1

     Since 15 Sep 2026 YAGPDB ignores prefixed BUILT-IN commands (slash or
     @mention only), but custom commands still run and can call built-ins via
     exec. This hands everything after "ban" to the real Ban command unchanged,
     so user/duration/reason/-ddays parse exactly as the old >>ban did, and the
     modlog, DM, and reason rules all come from the Moderation settings.

     exec (not execAdmin) runs it AS THE CALLER, so YAGPDB's own ban-permission
     check still applies and the modlog names the real moderator.
     ===================================================================== */}}
{{ $args := trimSpace .StrippedMsg }}
{{ if not $args }}
	Usage: `ban <user ID or @mention> [duration] [reason] [-ddays 0-7]`
{{ else }}
	{{ try }}
		{{ with exec (print "ban " $args) }}{{ sendMessage nil . }}{{ end }}
	{{ catch }}
		Ban failed: {{ .Error }}
	{{ end }}
{{ end }}
