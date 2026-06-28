{{/* ===========================================================================
     NICKNAME NORMALIZER
     Runs on every message in #age_verification (Regex trigger (?s).* + a
     dashboard channel restriction). It reads the poster's CURRENT server
     nickname (or username if they have none), normalizes it, and — if the
     result differs — sets it as their nickname.

     Normalization, in order:
       1. Map "fancy font" Unicode letters/digits back to plain ASCII
          (𝕎𝕠𝕠𝕞𝕒 -> Wooma, 𝓓𝑂𝑉𝐸𝑇𝑇𝐸 -> Dovette). Covers the Mathematical
          Alphanumeric Symbols block (bold / italic / script / fraktur /
          double-struck / sans / monospace, incl. the Letterlike "holes" like
          ℋ ℝ ℤ), plus fullwidth, circled, parenthesized, squared,
          regional-indicator and small-caps styles.
       2. Drop everything that isn't a letter, digit, space, or one of the
          allowed specials !@#$%&+-  (so emojis, colons, slashes, quotes,
          apostrophes and brackets are removed).
       3. Collapse runs of whitespace to a single space.
       4. Trim leading/trailing spaces and specials, so the first and last
          characters are alphanumeric.
       5. Title-case: capitalize the first letter of each word, lowercase the
          rest; digits are left alone. A space OR an allowed special (e.g. the
          hyphen in lonely-guy -> Lonely-Guy) starts a new word; a digit does
          not (abc2def -> Abc2def).

     Real accented letters (José, Müller) are passed through untouched — they
     are treated as letters, not stripped, and their case is left as-is.

     NOTE: YAGPDB applies a nickname change on the member's NEXT message, not
     instantly. The bot needs Manage Nicknames and a role above the member.
     =========================================================================== */}}

{{ if .User.Bot }}{{ return }}{{ end }}
{{/* Never try to rename the server owner — editNickname would just error. */}}
{{ if eq (toString .User.ID) (toString .Guild.OwnerID) }}{{ return }}{{ end }}

{{ $U := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" }}
{{ $L := "abcdefghijklmnopqrstuvwxyz" }}
{{ $D := "0123456789" }}

{{/* Irregular single-character maps that don't fall in a contiguous range:
     - Math "holes": script/fraktur/double-struck letters that Unicode borrowed
       from the Letterlike Symbols block instead of the Math block.
     - The circled zero ⓪ (the other circled digits ①-⑨ are contiguous).
     - Small-caps letters (ᴀ ʙ ᴄ …), which are scattered IPA code points. */}}
{{ $map := dict
  8459 "H" 8464 "I" 8466 "L" 8475 "R" 8492 "B" 8496 "E" 8497 "F" 8499 "M"
  8458 "g" 8462 "h" 8495 "e" 8500 "o"
  8493 "C" 8460 "H" 8465 "I" 8476 "R" 8488 "Z"
  8450 "C" 8461 "H" 8469 "N" 8473 "P" 8474 "Q" 8477 "R" 8484 "Z"
  9450 "0"
  7424 "a" 665 "b" 7428 "c" 7429 "d" 7431 "e" 42800 "f" 610 "g" 668 "h"
  618 "i" 7434 "j" 7435 "k" 671 "l" 7437 "m" 628 "n" 7439 "o" 7448 "p"
  640 "r" 42801 "s" 7451 "t" 7452 "u" 7456 "v" 7457 "w" 655 "y" 7458 "z"
}}

{{/* Source string: current nickname, else username. */}}
{{ $display := .User.Username }}
{{ if .Member.Nick }}{{ $display = .Member.Nick }}{{ end }}

{{/* ----- STAGE 1: map fancy -> ASCII, drop disallowed -----
     NOTE: YAGPDB's `slice` on a string is BYTE-indexed, but ranging over
     `toRune` yields RUNE indices. So we can't slice the original string by the
     range index when it contains multi-byte characters. We track the real byte
     offset ($b) and each rune's UTF-8 width ($w) and slice by bytes. */}}
{{ $out := cslice }}
{{ $b := 0 }}
{{ range $r := toRune $display }}
  {{ $cp := toInt $r }}
  {{ $w := 1 }}{{ if ge $cp 65536 }}{{ $w = 4 }}{{ else if ge $cp 2048 }}{{ $w = 3 }}{{ else if ge $cp 128 }}{{ $w = 2 }}{{ end }}
  {{ $ch := "" }}
  {{ if and (ge $cp 32) (le $cp 126) }}
    {{/* printable ASCII: keep only letters, digits, space, allowed specials */}}
    {{ if or (and (ge $cp 65) (le $cp 90)) (and (ge $cp 97) (le $cp 122)) (and (ge $cp 48) (le $cp 57)) (eq $cp 32) (in (cslice 33 35 36 37 38 43 45 64) $cp) }}
      {{ $ch = slice $display $b (add $b $w) }}
    {{ end }}
  {{ else if and (ge $cp 119808) (le $cp 120483) }}
    {{/* Mathematical Alphanumeric letters: 13 styles, each 26 upper + 26 lower */}}
    {{ $m52 := toInt (mod (sub $cp 119808) 52) }}
    {{ if lt $m52 26 }}{{ $ch = slice $U $m52 (add $m52 1) }}{{ else }}{{ $p := sub $m52 26 }}{{ $ch = slice $L $p (add $p 1) }}{{ end }}
  {{ else if and (ge $cp 120782) (le $cp 120831) }}
    {{/* Mathematical Alphanumeric digits */}}
    {{ $p := toInt (mod (sub $cp 120782) 10) }}{{ $ch = slice $D $p (add $p 1) }}
  {{ else if and (ge $cp 65313) (le $cp 65338) }}{{ $p := sub $cp 65313 }}{{ $ch = slice $U $p (add $p 1) }}
  {{ else if and (ge $cp 65345) (le $cp 65370) }}{{ $p := sub $cp 65345 }}{{ $ch = slice $L $p (add $p 1) }}
  {{ else if and (ge $cp 65296) (le $cp 65305) }}{{ $p := sub $cp 65296 }}{{ $ch = slice $D $p (add $p 1) }}
  {{ else if and (ge $cp 9398) (le $cp 9423) }}{{ $p := sub $cp 9398 }}{{ $ch = slice $U $p (add $p 1) }}
  {{ else if and (ge $cp 9424) (le $cp 9449) }}{{ $p := sub $cp 9424 }}{{ $ch = slice $L $p (add $p 1) }}
  {{ else if and (ge $cp 9312) (le $cp 9320) }}{{ $p := sub $cp 9311 }}{{ $ch = slice $D $p (add $p 1) }}
  {{ else if and (ge $cp 9372) (le $cp 9397) }}{{ $p := sub $cp 9372 }}{{ $ch = slice $L $p (add $p 1) }}
  {{ else if and (ge $cp 9332) (le $cp 9340) }}{{ $p := sub $cp 9331 }}{{ $ch = slice $D $p (add $p 1) }}
  {{ else if and (ge $cp 127344) (le $cp 127369) }}{{ $p := sub $cp 127344 }}{{ $ch = slice $U $p (add $p 1) }}
  {{ else if and (ge $cp 127408) (le $cp 127433) }}{{ $p := sub $cp 127408 }}{{ $ch = slice $U $p (add $p 1) }}
  {{ else if and (ge $cp 127462) (le $cp 127487) }}{{ $p := sub $cp 127462 }}{{ $ch = slice $U $p (add $p 1) }}
  {{ else }}
    {{ $hit := $map.Get $cp }}
    {{ if $hit }}{{ $ch = $hit }}
    {{ else if and (ge $cp 192) (le $cp 255) (ne $cp 215) (ne $cp 247) }}{{ $ch = slice $display $b (add $b $w) }}
    {{ else if and (ge $cp 256) (le $cp 383) }}{{ $ch = slice $display $b (add $b $w) }}
    {{ end }}
  {{ end }}
  {{ if $ch }}{{ $out = $out.Append $ch }}{{ end }}
  {{ $b = add $b $w }}
{{ end }}
{{ $mapped := joinStr "" $out }}

{{/* ----- STAGE 2: collapse whitespace, trim spaces/specials off the ends ----- */}}
{{ $clean := reReplace "\\s+" $mapped " " }}
{{ $clean = reReplace "^[ !@#$%&+-]+" $clean "" }}
{{ $clean = reReplace "[ !@#$%&+-]+$" $clean "" }}

{{/* ----- STAGE 3: title-case (cap first letter of each word, lowercase rest),
        capped at Discord's 32-character nickname limit. Byte-offset slicing as
        in Stage 1, because $clean may contain multi-byte accented letters. ----- */}}
{{ $titled := cslice }}
{{ $wordStart := true }}
{{ $tb := 0 }}
{{ $count := 0 }}
{{ range $r := toRune $clean }}
  {{ $cp := toInt $r }}
  {{ $w := 1 }}{{ if ge $cp 65536 }}{{ $w = 4 }}{{ else if ge $cp 2048 }}{{ $w = 3 }}{{ else if ge $cp 128 }}{{ $w = 2 }}{{ end }}
  {{ if lt $count 32 }}
    {{ $c := slice $clean $tb (add $tb $w) }}
    {{ if eq $cp 32 }}
      {{ $titled = $titled.Append " " }}{{ $wordStart = true }}
    {{ else if or (and (ge $cp 65) (le $cp 90)) (and (ge $cp 97) (le $cp 122)) }}
      {{ if $wordStart }}{{ $titled = $titled.Append (upper $c) }}{{ else }}{{ $titled = $titled.Append (lower $c) }}{{ end }}
      {{ $wordStart = false }}
    {{ else if and (ge $cp 48) (le $cp 57) }}
      {{/* digit: left alone, does not start a new word (abc2def -> Abc2def) */}}
      {{ $titled = $titled.Append $c }}{{ $wordStart = false }}
    {{ else }}
      {{/* allowed special (e.g. hyphen) is a word boundary: lonely-guy -> Lonely-Guy */}}
      {{ $titled = $titled.Append $c }}{{ $wordStart = true }}
    {{ end }}
    {{ $count = add $count 1 }}
  {{ end }}
  {{ $tb = add $tb $w }}
{{ end }}
{{ $final := joinStr "" $titled }}
{{ $final = reReplace "[ !@#$%&+-]+$" $final "" }}

{{/* ----- APPLY ----- */}}
{{ if eq $final "" }}
  {{/* Nothing usable left (e.g. an all-emoji nick). DM them — but only once
       every 6h so we don't spam on every message they send. */}}
  {{ $guard := dbGet .User.ID "nickDMGuard" }}
  {{ if not $guard.Value }}
    {{ sendDM "Hi! We couldn't tidy your server nickname into a readable name — it doesn't contain standard letters or numbers we can use. Please change your nickname to something readable. Thanks!" }}
    {{ dbSetExpire .User.ID "nickDMGuard" 1 21600 }}
  {{ end }}
{{ else if ne $final $display }}
  {{ editNickname $final }}
{{ end }}
