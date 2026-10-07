{{/* Advert Reaction Cleaner

    Trigger Type:  Reaction - Added Only
    Channel Restriction: Set to your advert channels

    Removes reactions from anyone who is not:
      1. The original poster of the message (except for staff-only emojis)
      2. A member with a staff/admin role
*/}}

{{$staffRoleNames := cslice
    "rpc | management"
    "rpc | operations"
    "rpc | concierge"
    "rpc | engagement"
}}

{{$staffOnlyEmojis := cslice
    "staffapproved"
    "staffpending"
    "staffcheck"
}}


{{/* Ignore bots */}}
{{if .User.Bot}} {{return}} {{end}}

{{/* Allow the original poster to react to their own message, unless it's a staff-only emoji */}}
{{if and (eq .User.ID .ReactionMessage.Author.ID) (not (in $staffOnlyEmojis .Reaction.Emoji.Name))}} {{return}} {{end}}

{{/* Allow staff */}}
{{$isStaff := false}}
{{range $staffRoleNames}}
    {{if hasRoleName .}} {{$isStaff = true}} {{end}}
{{end}}
{{if $isStaff}} {{return}} {{end}}

{{/* Remove the reaction */}}
{{deleteMessageReaction nil .ReactionMessage.ID .User.ID .Reaction.Emoji.APIName}}
