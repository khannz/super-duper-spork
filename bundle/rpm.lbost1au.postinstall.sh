#!/usr/bin/env sh

mkdir -p /var/run/lbost1au

systemctl daemon-reload
systemctl enable --now lbost1au.service

exit 0
