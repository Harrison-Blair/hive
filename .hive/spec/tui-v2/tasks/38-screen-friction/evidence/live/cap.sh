#!/bin/sh
# cap.sh PANE OUT: read the pane until two reads 500 ms apart are equal.
prev=""
for i in 1 2 3 4 5 6 7 8 9 10; do
  herdr pane read "$1" --source visible --format ansi > "$2.tmp" 2>/dev/null
  cur=$(md5sum < "$2.tmp")
  [ "$cur" = "$prev" ] && break
  prev=$cur
  sleep 0.5
done
mv "$2.tmp" "$2"
