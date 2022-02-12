#!/bin/bash
set -e

if [ -n "$EXT_IP" ] && [ -n "$SOURCE_IP" ]; then
    iptables -t nat -A POSTROUTING -o `ip r get ${EXT_IP} | awk '{ print $3 }'` -j MASQUERADE --random
fi

sleep infinity