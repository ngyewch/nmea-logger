#!/bin/sh

set -e

case "$1" in
  remove|0)
    systemctl stop nmea-logger.service >/dev/null || true
  ;;

  upgrade|deconfigure|failed-upgrade)
  ;;

  *)
    echo "preremove.sh called with unknown argument '$1'" >&2
    exit 1
  ;;
esac

exit 0
