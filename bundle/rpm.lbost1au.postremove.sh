#!/usr/bin/env sh
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf /var/run/lbost1au
fi

exit 0
