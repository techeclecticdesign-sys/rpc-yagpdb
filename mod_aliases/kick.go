{{/* =====================================================================
     kick -- prefix alias for YAGPDB's built-in Kick command.

     Trigger type: Command.   Name: kick

       >>kick 1444117608000520232 Alt account
       >>kick @member Spamming -cl 50

     Same approach as ban.go: everything after "kick" goes to the real Kick
     command unchanged (user, reason, -cl message cleanup), run AS THE CALLER
     via exec so YAGPDB's kick-permission check and modlog still apply.
     ===================================================================== */}}
{{ $args := trimSpace .StrippedMsg }}
{{ if not $args }}
	Usage: `kick <user ID or @mention> [reason] [-cl 1-100]`
{{ else }}
	{{ try }}
		{{ with exec (print "kick " $args) }}{{ sendMessage nil . }}{{ end }}
	{{ catch }}
		Kick failed: {{ .Error }}
	{{ end }}
{{ end }}
