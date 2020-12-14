FROM alpine:latest
RUN apk update && apk upgrade && apk add tzdata
RUN apk add chromium
WORKDIR /bin
COPY /linux /bin
ENTRYPOINT alarm_service