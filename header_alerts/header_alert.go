{{/* =====================================================================
     Discord renders a line that starts with #, ## or ### followed by a space
     as a header. When a post contains one, this pings the poster in
     #rule_infractions asking them to remove it.
     See setup.txt for full step-by-step dashboard instructions.
     ===================================================================== */}}

{{/* ▼▼ Paste your #rule_infractions channel ID here (see setup.txt) ▼▼ */}}
{{ $infractionsChannel := 1505355022828048384 }}

{{ sendMessage $infractionsChannel (printf "Hello %s ! You are welcome to use bold, however, headers are not allowed in the one on one advert channels, Please go back to your advert in %s and remove the header. Thanks!" (printf "<@%d>" .User.ID) (printf "<#%d>" .Channel.ID)) }}
