#!/bin/bash
# cap.sh BOARD W H OUTDIR: run the demo in the scratch pane of the exact board
# size (or a larger pane PWxPH, cropped top-left), read the pane until two reads are equal, compare with the grid, quit.
set -u
H="herdr --session t131-w2"; P=w1:p1; B=$1; W=$2; S=$3; O=$4
G=.hive/spec/tui-v2/design/$B.grid
$H pane run $P "clear; TERM=vt100 COLORTERM=truecolor .hive/tmp/hive --demo .hive/spec/tui-v2/design/$B.fixture.json --size ${W}x${S}" >/dev/null
prev=""; for i in $(seq 1 30); do
  sleep 0.5
  cur=$($H pane read $P --source visible --format ansi)
  [ -n "$prev" ] && [ "$cur" = "$prev" ] && break
  prev=$cur
done
printf '%s' "$cur" > "$O/$B-${W}x${S}.ansi"
echo "$B-${W}x${S}.ansi rows=$(printf '%s\n' "$cur" | wc -l) $(.hive/tmp/gridcmp/gridcmp "$O/$B-${W}x${S}.ansi" "$G" "$W" "$S" "$O/$B-${W}x${S}.diff.txt" ${PW:-} ${PH:-} 2>&1)"
$H pane send-keys $P ctrl+c ctrl+c >/dev/null; sleep 1
