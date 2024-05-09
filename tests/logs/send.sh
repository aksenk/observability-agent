#!/bin/bash

set -e

LOG_FILE="_logs"
USER_ID="111"

JWT="$(bash ../jwt_generator/generate.sh)"

LOGS='It is first line
It is second line
{"custom_log": "third line"}'

echo "$LOGS" > $LOG_FILE
date >> $LOG_FILE

cat $LOG_FILE | gzip > ${LOG_FILE}.gz

#echo "Sending plain logs without user-id"
#curl -i -XPOST 'http://localhost:8080/api/v1/logs/elasticsearch/bulk' -T $LOG_FILE
#echo
#echo "Sending plain logs with user-id"
#curl -i -XPOST -H "user-id: $USER_ID" 'http://localhost:8080/api/v1/logs/elasticsearch/bulk' -T $LOG_FILE
#echo
#echo "Sending gzip logs without gzip header"
#curl -i -XPOST -H "user-id: $USER_ID" 'http://localhost:8080/api/v1/logs/elasticsearch/bulk' -T ${LOG_FILE}.gz
#echo
#echo "Sending plain logs with gzip"
#curl -i -XPOST -H 'Content-Encoding: gzip' -H "user-id: $USER_ID" 'http://localhost:8080/api/v1/logs/elasticsearch/bulk' -T ${LOG_FILE}
#echo

echo "Sending gzip logs with gzip header"
curl -i -H "x-access-token: $(bash ../jwt_generator/generate.sh)" -XPOST -H 'Content-Encoding: gzip' -T ${LOG_FILE}.gz \
  'http://localhost:8080/api/v1/logs/elasticsearch/bulk'
echo

rm $LOG_FILE ${LOG_FILE}.gz
