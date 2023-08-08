#!/bin/bash

set -e


gf docker main.go -p -t xyhelper/xyhelper-arkose-v2:latest
#  获取当前时间
now=$(date +%Y%m%d%H%M%S)
docker tag xyhelper/xyhelper-arkose-v2:latest xyhelper/xyhelper-arkose-v2:$now
docker push xyhelper/xyhelper-arkose-v2:$now