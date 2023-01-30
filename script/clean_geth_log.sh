#!/bin/sh
echo "======== start clean geth logs ========"
logs=$(find /data/ -name *.out)
for log in $logs
do
echo "clean logs : $log"
cat /dev/null > $log
done
echo "======== end clean geth logs ========"
