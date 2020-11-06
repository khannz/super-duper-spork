#!/usr/bin/env sh
if [ "$1" -ge 1 ]; then
  systemctl stop lbost1au.service
fi
if [ "$1" = 0 ]; then
  systemctl disable --now lbost1au.service
fi

exit 0
