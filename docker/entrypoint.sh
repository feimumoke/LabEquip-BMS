#!/bin/sh
set -e

cd /root

# Nginx 访问 /root/uploads 需要目录可遍历
chmod 755 /root 2>/dev/null || true
mkdir -p /root/uploads /root/logs
chmod 755 /root/uploads /root/logs 2>/dev/null || true

MYSQL_HOST="${MYSQL_HOST:-mysql}"
echo "Waiting for MySQL at ${MYSQL_HOST}:3306 ..."
i=0
while [ "$i" -lt 60 ]; do
  if nc -z "$MYSQL_HOST" 3306 2>/dev/null; then
    echo "MySQL is reachable."
    break
  fi
  i=$((i + 1))
  sleep 1
done
sleep 2

echo "Starting bms-api..."
./bms-api &

echo "Starting nginx..."
exec nginx -g "daemon off;"
