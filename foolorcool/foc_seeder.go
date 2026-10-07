{{/* Fool or Cool seeder — run ONCE to load the questions into the
     "foolorcoolquestions" DB entry (owner 0). Like rp_seeder / oc_seeder it
     OVERWRITES the key in one shot, so it's safe to re-run: it always leaves
     exactly this list, never duplicates, and resets the walk counter to the top.
     Add more backtick lines below, one per line, then re-run this command.
     Delete this command after seeding — the data stays in the DB. */}}
{{ $slice := cslice
  `Have you ever spilled a drink on yourself at a public gathering?`
  `Have you ever walked into a glass door?`
  `Have you ever tripped going up the stairs in a public space?`
  `Have you ever waved back at someone who wasn't waving at you?`
  `Have you ever called a teacher "mom" or "dad"?`
  `Have you ever left the bathroom with toilet paper stuck to your shoe, pants or etc?`
  `Have you ever accidentally mistexted to the exact person you were talking about?`
  `Have you ever accidentally liked a post from three years ago?`
  `Have you ever gone in for a hug when they went in for a handshake?`
  `Have you ever thought headphones were on when the speakers were blaring?`
  `Have you ever answered a question that was meant for someone else?`
  `Have you ever had your fly down all day without noticing?`
  `Have you ever unmuted yourself on a call at the worst possible time?`
  `Have you ever had your video on when you thought it was off?`
  `Have you ever walked into the wrong bathroom, locker room or etc?`
  `Have you ever sent a voice message you wished you could unsend?`
  `Have you ever tried to sit on a chair that wasn't there?`
  `Have you ever clapped when nobody else was clapping?`
  `Have you ever gotten a haircut so bad you avoided people?`
  `Have you ever run into a pole while looking at your phone?`
  `Have you ever gotten a nosebleed at the worst possible time?`
  `Have you ever been the only one dressed up at a casual event or vice versa?`
  `Have you ever shown up on the wrong day for something?`
  `Have you ever ripped your pants in public?`
  `Have you ever tried to high five someone who left you hanging?`
  `Have you ever accidentally screamed at a jump scare in a theater?`
  `Have you ever gotten stuck talking to someone whose name you forgot?`
  `Have you ever posted a something you had to delete within the minute?`
  `Have you ever texted "I love you" to the wrong person?`
  `Have you ever pushed onto a door that was pull only and walked into it after?`
  `Have you ever been recorded doing something you begged them to delete?`
  `Have you ever stared off into space and realized you were staring at someone?`
  `Have you ever tried to open someone else's car that looked like yours?`
  `Have you ever leaned back so far in a chair you fell backwards?`
  `Have you ever gone to the wrong house and knocked?`
  `Have you ever autocorrected into something humiliating?`
  `Have you ever had a group photo taken mid-blink and it became the one?`
  `Have you ever missed your mouth with a drink?`
  `Have you ever done a big presentation with a typo on the first slide?`
  `Have you ever been the only person who laughed?`
}}
{{ dbSet 0 "foolorcoolquestions" $slice }}
{{ dbSet 0 "foolorcoolquestions_idx" 0 }}
Loaded Fool or Cool questions. Total: {{ len $slice }} (walk reset to the top)