#!/bin/bash
# keys.sh: drive the ChatAskMulti demo. Each step: LABEL KIND VALUE (KIND keys|text).
P=wWD:p2; O=.hive/tmp/v4/keys; mkdir -p $O; n=0
read_stable() { prev=""; for i in $(seq 1 20); do sleep 0.3; cur=$(herdr pane read $P --source visible --format text); [ "$cur" = "$prev" ] && break; prev=$cur; done; printf '%s\n' "$cur"; }
herdr pane run $P "clear; TERM=vt100 COLORTERM=truecolor .hive/tmp/hive --demo .hive/spec/tui-v2/design/ChatAskMulti.fixture.json --size 120x40" >/dev/null
read_stable > $O/00-start.txt
step() {
  n=$((n+1)); label=$1; kind=$2; shift 2
  if [ "$kind" = keys ]; then herdr pane send-keys $P "$@" >/dev/null; else herdr pane send-text $P "$1" >/dev/null; fi
  f=$(printf '%s/%02d-%s.txt' $O $n "$label"); read_stable > $f
}
E=$'\e'
step ctrl-bracket-leave-note keys ctrl+]
step up keys up
step down keys down
step tab-open-note keys tab
step type-in-note text " now"
step ctrl-bracket-keep-note keys ctrl+]
step tab-again keys tab
step tab-drops-note keys tab
step J-next-card keys J
step K-prev-card keys K
step space-pick keys space
step digit-3-other keys 3
step type-other text "later"
step ctrl-bracket-leave-other keys ctrl+]
step wheel-up-over-card text "${E}[<64;60;30M${E}[<64;60;30M"
step wheel-down-over-card text "${E}[<65;60;30M"
step pgup text "${E}[5~"
step pgdn text "${E}[6~"
step click-option-2-q2 text "${E}[<0;20;33M${E}[<0;20;33m"
step click-feed text "${E}[<0;60;10M${E}[<0;60;10m"
step click-card text "${E}[<0;60;26M${E}[<0;60;26m"
step i-input keys i
step ctrl-bracket-back keys ctrl+]
step enter-send keys enter
step x-dismiss keys x
herdr pane send-keys $P ctrl+c ctrl+c >/dev/null; sleep 1
