{{/* Threemoji seeder — run ONCE to load the questions into the
     "threemojiquestions" DB entry (owner 0). Like foc_seeder / rp_seeder it
     OVERWRITES the key in one shot, so it's safe to re-run: it always leaves
     exactly this list, never duplicates, and resets the walk counter to the top.

     DIFFERENCE from the other seeders: this one SHUFFLES the list before storing
     it (via the `shuffle` function), so the sequential asker walks a randomized
     order. Because the shuffle happens here, re-running this command RESHUFFLES —
     handy if you want a fresh order, but note the order will change each run.

     Add more backtick lines below, one per line, then re-run this command.
     Delete this command after seeding — the data stays in the DB. */}}
{{ $slice := shuffle (cslice
  `Describe your personality using three emojis.`
  `Describe your current mood using three emojis.`
  `Describe the state of your life right now using three emojis.`
  `Describe your childhood using three emojis.`
  `Describe your style using three emojis.`
  `Describe your energy level using three emojis.`
  `Describe your brain using three emojis.`
  `Describe your sleep schedule using three emojis.`
  `Describe your ideal day using three emojis.`
  `Describe your attention span using three emojis.`
  `Describe your cooking ability using three emojis.`
  `Describe your driving using three emojis.`
  `Describe your sense of direction using three emojis.`
  `Describe your ability to keep secrets using three emojis.`
  `Describe your ability to handle stress using three emojis.`
  `Describe your relationship with money using three emojis.`
  `Describe your relationship with technology using three emojis.`
  `Describe your relationship with caffeine using three emojis.`
  `Describe your best friend using three emojis.`
  `Describe your friend group using three emojis.`
  `Describe your family using three emojis.`
  `Describe your sibling(s) using three emojis.`
  `Describe your parents using three emojis.`
  `Describe your favorite coworker using three emojis.`
  `Describe your boss using three emojis.`
  `Describe your first romantic partner using three emojis.`
  `Describe your celebrity crush using three emojis.`
  `Describe your pet(s) using three emojis (you can do a set of three for each one).`
  `Describe your relationship with alcohol using three emojis.`
  `Describe your first day of school using three emojis.`
  `Describe your first concert using three emojis.`
  `Describe your car using three emojis.`
  `Describe your teenage personality using three emojis.`
  `Describe your teenage fashion choices using three emojis.`
  `Describe your first year of roleplaying using three emojis.`
  `Describe your job using three emojis.`
  `Describe your coworkers using three emojis.`
  `Describe your work ethic using three emojis.`
  `Describe your relationship with Mondays using three emojis.`
  `Describe your relationship with Weekends using three emojis.`
  `Describe your relationship with meetings using three emojis.`
  `Describe your favorite movie using three emojis.`
  `Describe your favorite TV show using three emojis.`
  `Describe your favorite book using three emojis.`
  `Describe your favorite video game using three emojis.`
  `Describe your favorite song using three emojis.`
  `Describe your favorite musician using three emojis.`
  `Describe your favorite actor using three emojis.`
  `Describe your favorite fictional character using three emojis.`
  `Describe your favorite villain using three emojis.`
  `Describe your favorite superhero using three emojis.`
  `Describe your YouTube history using three emojis.`
  `Describe your music taste using three emojis.`
  `Describe your TV-watching habits using three emojis.`
  `Describe your most embarrassing moment using three emojis.`
  `Describe your dumbest decision using three emojis.`
  `Describe your worst habit using three emojis.`
  `Describe your most irrational fear using three emojis.`
  `Describe your biggest pet peeve using three emojis.`
  `Describe your most questionable purchase using three emojis.`
  `Describe the weirdest thing you've done while bored using three emojis.`
  `Describe your life after winning the lottery using three emojis.`
  `Describe yourself during a zombie apocalypse using three emojis.`
  `Describe yourself stranded on a desert island using three emojis.`
  `Describe your life as a billionaire using three emojis.`
  `Describe your life in a medieval kingdom using three emojis.`
  `Describe your life as a superhero using three emojis.`
  `Describe your life as a villain using three emojis.`
  `You have one hour to spend $10,000. Describe your choices using three emojis.`
) }}
{{ dbSet 0 "threemojiquestions" $slice }}
{{ dbSet 0 "threemojiquestions_idx" 0 }}
Loaded Threemoji questions. Total: {{ len $slice }} (shuffled, walk reset to the top)
