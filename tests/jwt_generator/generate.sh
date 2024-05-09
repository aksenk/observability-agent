#!/bin/bash

SECRET_KEY="your_secret_key_here"

export USER_ID=666
export SESSION_ID=kek-fek-gek
export IAT=$(date +%s)
export EXP=$(date -v +1d +%s)

HEADER='{"alg":"HS256","typ":"JWT"}'
PAYLOAD="{\"gambler_id\": $USER_ID, \"gamblerId\": $USER_ID, \"session_id\": \"$SESSION_ID\", \"iat\": $IAT, \"exp\": $EXP}"

header_base64=$(echo -n "$HEADER" | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')
payload_base64=$(echo -n "$PAYLOAD" | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')

signature=$(echo -n "$header_base64.$payload_base64" | openssl dgst -binary -sha256 -hmac "$SECRET_KEY" | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')

jwt_token="$header_base64.$payload_base64.$signature"

#echo "j:{\"token\":\"$jwt_token\",\"refresh_token\":\"$jwt_token\"}"
echo "$jwt_token"
