#!/bin/bash
# cap.sh BOARD W H OUTDIR [GRID]: run the demo, read the pane until stable, compare, quit.
set -u
P=wWD:p2; B=$1; W=$2; H=$3; O=$4; G=${5:-.hive/spec/tui-v2/tasks/131-hive-ask/evidence/new-grids/$B.grid}
herdr pane run $P "clear; TERM=vt100 COLORTERM=truecolor .hive/tmp/hive --demo .hive/spec/tui-v2/design/$B.fixture.json --size ${W}x${H}" >/dev/null
prev=""; for i in $(seq 1 30); do
  sleep 0.5
  cur=$(herdr pane read $P --source visible --format ansi)
  [ -n "$prev" ] && [ "$cur" = "$prev" ] && break
  prev=$cur
done
printf '%s' "$cur" > $O/$B-${W}x${H}.ansi
echo "$B-${W}x${H}.ansi $(.hive/tmp/gridcmp/gridcmp $O/$B-${W}x${H}.ansi $G $W $H $O/$B-${W}x${H}.diff.txt 2>&1)"
herdr pane send-keys $P ctrl+c ctrl+c >/dev/null; sleep 1
