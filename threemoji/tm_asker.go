{{/* Sequential /threemoji asker — posts one "describe X using three emojis"
     prompt as an EMBED. Open-ended (no vote reactions): people just reply with
     their three emojis. Walks the "threemojiquestions" list top-to-bottom via a
     persisted counter ("threemojiquestions_idx") instead of random, so nothing
     repeats until the whole list has been shown, then wraps back to the top. The
     list itself was SHUFFLED once at seed time (see tm_seeder.go), so the fixed
     walk order is already randomized. dbIncr is atomic, so simultaneous uses
     still advance cleanly. SEPARATE from the "questions", "rpquestions",
     "ocquestions", "wbquestions", "foolorcoolquestions", and
     "sinnerorsaintquestions" keys, so the lists never mix.

     The title is three emojis, the game name, then three emojis (mirrored). The
     prompt goes in the description. Works from a Command trigger (users run
     -threemoji) or an Interval trigger (the bot posts one on its own); nil = the
     current/target channel in both cases. No reactions are added, so no extra
     permissions are required beyond posting.

     DB BUDGET: dbGet(list) + dbIncr(counter) = 2 db_interactions, 0 db_multiple —
     nowhere near the free-tier caps. */}}
{{- $q := (dbGet 0 "threemojiquestions").Value -}}
{{- if $q -}}
{{- $c := toInt (dbIncr 0 "threemojiquestions_idx" 1) -}}
{{- $i := toInt (mod (sub $c 1) (len $q)) -}}
{{- $question := index $q $i -}}
{{- sendMessage nil (complexMessage "embed" (cembed "title" "🎭 🔮 ✨ Threemoji ✨ 🔮 🎭" "description" (printf "**%s**" $question) "color" 0x9B59B6)) -}}
{{- else -}}
No Threemoji questions loaded yet.
{{- end -}}
