{{/* =====================================================================
     When someone posts a link in one of the quick channels, this pings the
     poster in #rule_infractions asking them to remove it from their ad.
     See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{/* ▼▼ Paste your #rule_infractions channel ID here (see setup.txt) ▼▼ */}}
{{ $infractionsChannel := 0 }}

{{ sendMessage $infractionsChannel (printf "Hey %s !  Links aren’t allowed in the quick search channels. Please remove it from your ad in %s. Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID)) }}
