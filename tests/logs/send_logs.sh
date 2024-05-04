#!/bin/bash

set -e

LOG_FILE="_logs"
USER_ID="111"

LOGS='It is first line
It is second line
{"custom_log": "third line"}'

echo "$LOGS" > $LOG_FILE
date >> $LOG_FILE

cat $LOG_FILE | gzip > ${LOG_FILE}.gz

echo "Sending logs without user-id"
curl -i -XPOST 'http://localhost:8080/api/v1/logs/elasticsearch/bulk' -T $LOG_FILE

echo "Sending logs with user-id"
curl -i -XPOST -H "user-id: $USER_ID" 'http://localhost:8080/api/v1/logs/elasticsearch/bulk' -T $LOG_FILE

echo "Sending logs with gzip"
curl -i -XPOST -H 'Content-Encoding: gzip' -H "user-id: $USER_ID" 'http://localhost:8080/api/v1/logs/elasticsearch/bulk' -T ${LOG_FILE}.gz

#rm $LOG_FILE ${LOG_FILE}.gz