#!/bin/sh

set -e

SERVICE_USER=nmea-logger
SERVICE_GROUP=nmea-logger

case "$1" in
  install|upgrade|1|2)
    useradd --system --user-group --home-dir /var/lib/${SERVICE_USER} --create-home --shell /sbin/nologin ${SERVICE_USER} \
        || EXIT_CODE=$? \
        && [ -n "${EXIT_CODE+set}" ] && [ "${EXIT_CODE}" -ne 9 ] \
        && exit ${EXIT_CODE}
    usermod -a -G dialout ${SERVICE_USER}
  ;;

  abort-upgrade)
  ;;

  *)
    echo "preinstall.sh called with unknown argument '$1'" >&2
    exit 1
  ;;
esac

exit 0
