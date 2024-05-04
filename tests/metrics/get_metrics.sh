#!/bin/bash

curl localhost:8428/api/v1/export -d 'match={__name__="mobile_requests",gambler_id="1234"}'; echo