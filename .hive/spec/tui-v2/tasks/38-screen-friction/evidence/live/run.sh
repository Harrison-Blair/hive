#!/bin/sh
# run.sh BOARD W H NAME [KEY...]: start the demo of BOARD in pane w96:p3,
# press each KEY, capture the pane, and compare it with the board grid.
set -e
P=w96:p3; D=/home/penguin/source/hive/.hive/spec/tui-v2/design; T=$(dirname "$0")
B=$1; W=$2; H=$3; N=$4; shift 4
herdr pane run $P "clear; COLORTERM=truecolor $T/hive --demo $D/$B.fixture.json --size ${W}x$H" >/dev/null
sleep 1.5
for k in "$@"; do herdr pane send-keys $P "$k" >/dev/null; sleep 0.4; done
$T/cap.sh $P $T/out/$N.ansi
herdr pane get $P | grep -o '"scroll":{[^}]*}' > $T/out/$N-pane.json
$T/cmp $D/$B.grid $T/out/$N.ansi $W $H > $T/out/$N-compare.txt || true
herdr pane send-keys $P ctrl+c >/dev/null; sleep 0.3; herdr pane send-keys $P ctrl+c >/dev/null; sleep 0.5
head -1 $T/out/$N-compare.txt
