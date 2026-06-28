{{/* ===========================================================================
     NAMETEST  —  dry-run tester for the Nickname Normalizer
     Type:   nametest <name>
     The bot replies with what that name normalizes to. It does NOT touch
     anyone's nickname — it's purely for previewing/QAing the normalization.

     The normalization logic below is IDENTICAL to nickname_normalize.go; only
     the INPUT (the typed argument instead of the member's nickname) and the
     OUTPUT (a chat reply instead of editNickname) differ. If you change the
     rules in one file, mirror the change in the other.

     Suggested trigger: Regex,  (?i)^nametest\b   (prefix-independent).
     Or a Command trigger named "nametest" — then swap the $display line below
     to: {{ $display := .StrippedMsg }}
     =========================================================================== */}}

{{/* Grab everything after the "nametest" keyword as the name to test. */}}
{{ $display := trimSpace (reReplace "(?i)^\\s*nametest\\s*" .Message.Content "") }}

{{ if eq $display "" }}
  {{- "Usage: `nametest <name>` — I'll show you what that name normalizes to." -}}
{{ else }}

{{ $U := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" }}
{{ $L := "abcdefghijklmnopqrstuvwxyz" }}
{{ $D := "0123456789" }}

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

{{/* ----- STAGE 1: map fancy -> ASCII, drop disallowed -----
     YAGPDB's `slice` on a string is BYTE-indexed while `toRune` gives RUNE
     indices, so we track the real byte offset ($b) and each rune's UTF-8 width
     ($w) and slice by bytes. */}}
{{ $out := cslice }}
{{ $b := 0 }}
{{ range $r := toRune $display }}
  {{ $cp := toInt $r }}
  {{ $w := 1 }}{{ if ge $cp 65536 }}{{ $w = 4 }}{{ else if ge $cp 2048 }}{{ $w = 3 }}{{ else if ge $cp 128 }}{{ $w = 2 }}{{ end }}
  {{ $ch := "" }}
  {{ if and (ge $cp 32) (le $cp 126) }}
    {{ if or (and (ge $cp 65) (le $cp 90)) (and (ge $cp 97) (le $cp 122)) (and (ge $cp 48) (le $cp 57)) (eq $cp 32) (in (cslice 33 35 36 37 38 43 45 64) $cp) }}
      {{ $ch = slice $display $b (add $b $w) }}
    {{ end }}
  {{ else if and (ge $cp 119808) (le $cp 120483) }}
    {{ $m52 := toInt (mod (sub $cp 119808) 52) }}
    {{ if lt $m52 26 }}{{ $ch = slice $U $m52 (add $m52 1) }}{{ else }}{{ $p := sub $m52 26 }}{{ $ch = slice $L $p (add $p 1) }}{{ end }}
  {{ else if and (ge $cp 120782) (le $cp 120831) }}
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

{{/* ----- STAGE 3: title-case, capped at 32 chars (byte-offset slicing) ----- */}}
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
    {{ else }}
      {{ $titled = $titled.Append $c }}{{ $wordStart = false }}
    {{ end }}
    {{ $count = add $count 1 }}
  {{ end }}
  {{ $tb = add $tb $w }}
{{ end }}
{{ $final := joinStr "" $titled }}
{{ $final = reReplace "[ !@#$%&+-]+$" $final "" }}

{{/* ----- REPLY (no nickname is changed) ----- */}}
{{ if eq $final "" }}
  {{- printf "**Input:** `%s`\n**Normalized:** *(empty — this name has no usable letters/numbers, so the real command would reject it and DM the member)*" $display -}}
{{ else }}
  {{- printf "**Input:** `%s`\n**Normalized:** `%s`" $display $final -}}
{{ end }}

{{ end }}
