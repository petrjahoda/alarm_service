#!/usr/bin/env bash
cd linux
upx alarm_service_linux
cd ..

docker rmi -f petrjahoda/alarm_service:latest
docker build -t petrjahoda/alarm_service:latest .
docker push petrjahoda/alarm_service:latest

docker rmi -f petrjahoda/alarm_service:2020.3.2
docker build -t petrjahoda/alarm_service:2020.3.2 .
docker push petrjahoda/alarm_service:2020.3.2
