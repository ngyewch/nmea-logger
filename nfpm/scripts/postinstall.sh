#!/bin/sh

set -e

case "$1" in
  configure|1|2)
    systemctl daemon-reload
  ;;

  *)
    echo "postinstall.sh called with unknown argument '$1'" >&2
    exit 1
  ;;
esac

exit 0
