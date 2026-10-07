{{/* Sequential /foolorcool asker — the wouldyourather-style vote poll, but a
     SINGLE question with two fixed answers: 🤡 (fool) vs 😎 (cool). Walks the
     "foolorcoolquestions" list top-to-bottom via a persisted counter
     ("foolorcoolquestions_idx") instead of random, so nothing repeats until
     the whole list has been shown, then it wraps back to the top. dbIncr is
     atomic, so simultaneous uses still advance cleanly. SEPARATE from the
     "questions", "rpquestions", "ocquestions", "wbquestions", and
     "sinnerorsaintquestions" keys, so the lists never mix.

     Unlike the plain topic askers, this one POSTS the question itself as an
     EMBED (via sendMessageRetID + complexMessage/cembed, the same look as the
     built-in wouldyourather) so it can grab the new message's ID and slap the
     two vote reactions on it with addMessageReactions — exactly how
     wouldyourather lets people vote. Works from a Command trigger (users run
     -foolorcool) or an Interval trigger (the bot posts one on its own); nil =
     the current/target channel in both cases. The bot needs the "Add Reactions"
     permission here.

     DB BUDGET: dbGet(list) + dbIncr(counter) = 2 db_interactions, 0 db_multiple —
     nowhere near the free-tier caps. */}}
{{- $q := (dbGet 0 "foolorcoolquestions").Value -}}
{{- if $q -}}
{{- $c := toInt (dbIncr 0 "foolorcoolquestions_idx" 1) -}}
{{- $i := toInt (mod (sub $c 1) (len $q)) -}}
{{- $question := index $q $i -}}
{{- $desc := printf "**%s**\n\n🤡 — yep, and I'll never live it down\n😎 — nope, still untouchable" $question -}}
{{- $id := sendMessageRetID nil (complexMessage "embed" (cembed "title" "🤡 Fool or Cool 😎" "description" $desc "color" 0xF1C40F)) -}}
{{- addMessageReactions nil $id "🤡" "😎" -}}
{{- else -}}
No Fool or Cool questions loaded yet.
{{- end -}}