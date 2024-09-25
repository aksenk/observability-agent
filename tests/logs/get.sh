#!/bin/bash

curl -ki -u elastic:password 'http://localhost:9200/logs/_search?pretty=true&q=1It*'
