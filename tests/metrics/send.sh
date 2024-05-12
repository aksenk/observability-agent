#!/bin/bash

set -e

if [ ! -f _current_value ]; then
  echo 10 > _current_value
fi
rm -f _metrics

USER_ID_FROM_ARG=$1
USER_ID=${USER_ID_FROM_ARG:=6666}

CURRENT_TS=`date "+%s"`
CURRENT_TS="${CURRENT_TS}000"

INCREASE_TO=10
CURRENT_VALUE=`cat _current_value`
NEW_VALUE=$((CURRENT_VALUE + INCREASE_TO))

TPL="{\"metric\":{\"__name__\":\"mobile_requests\",\"code\":\"200\",\"endpoint\":\"mobile-api\", \"platform\":\"android\"},\"values\":[${NEW_VALUE}],\"timestamps\":[${CURRENT_TS}]}\n{\"metric\":{\"__name__\":\"mobile_requests\",\"code\":\"200\",\"endpoint\":\"mobile-api\", \"platform\":\"android\"},\"values\":[${NEW_VALUE}],\"timestamps\":[${CURRENT_TS}]}"

#TPL="{\"metric\":{\"__name__\":\"mobile_requests\",\"code\":\"200\",\"endpoint\":\"mobile-api\", \"platform\":\"android\", \"gambler_id\": \"1234\"},\"values\":[${NEW_VALUE}],\"timestamps\":[${CURRENT_TS}]}"

echo -e $TPL > _metrics
echo $NEW_VALUE > _current_value
cat _metrics | gzip > _metrics.gz
curl -i -H "x-access-token: $(bash ../jwt_generator/generate.sh $USER_ID)" -H 'Content-Encoding: gzip' -T ./_metrics.gz \
  'http://localhost:8080/api/v1/metrics/victoriametrics/import'
