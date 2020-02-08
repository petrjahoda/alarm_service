#!/usr/bin/env bash
cd linux
upx alarm_service_linux
cd ..
docker rmi -f petrjahoda/alarm_service:"$1"
docker build -t petrjahoda/alarm_service:"$1" .
docker push petrjahoda/alarm_service:"$1"