{{/* =====================================================================
     When a post in a 1x1 advert channel has an uploaded image/file
     attachment, this pings the poster in #rule_infractions asking them to
     remove it from their ad.
     See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{/* ▼▼ Paste your #rule_infractions channel ID here (see setup.txt) ▼▼ */}}
{{ $infractionsChannel := 0 }}

{{ if gt (len .Message.Attachments) 0 }}
  {{ sendMessage $infractionsChannel (printf "Hey %s !  Images and other media aren’t allowed in the 1x1 advert channels. Please remove any from your ad in %s. Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID)) }}
{{ end }}
