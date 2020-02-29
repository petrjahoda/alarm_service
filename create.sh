#!/usr/bin/env bash
cd linux
upx alarm_service_linux
cd ..
cd mac
upx alarm_service_mac
cd ..
cd windows
upx alarm_service_windows.exe
cd ..
docker rmi -f petrjahoda/alarm_service:"$1"
docker build -t petrjahoda/alarm_service:"$1" .
docker push petrjahoda/alarm_service:"$1"