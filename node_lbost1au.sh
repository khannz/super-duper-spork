#!/usr/bin/env bash

set -u -o pipefail

count_utilisation() {
    ls /var/run/lbost1au/ |
    awk -F '.' '{sum += $3} END {print "node_lbos_bandwidth_utilisation_reserved " sum}'
}

utilisation=$(count_utilisation)

echo '# HELP node_lbos_bandwidth_utilisation_reserved Shows what is current level of reserved bandwidth'
echo '# TYPE node_lbos_bandwidth_utilisation_reserved gauge'

if [[ -n "${utilisation}" ]]; then
    echo "${utilisation}"
else
    echo 'node_lbos_bandwidth_utilisation_capacity 0'
fi

echo '# HELP node_lbos_bandwidth_utilisation_capacity Shows what is root limit for bandwidth'
echo '# TYPE node_lbos_bandwidth_utilisation_capacity gauge'
echo 'node_lbos_bandwidth_utilisation_capacity 1000'