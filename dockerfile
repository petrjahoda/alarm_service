FROM alpine:latest
RUN apk update && apk upgrade && apk add bash && apk add procps && apk add nano && apk add tzdata
RUN apk add chromium
WORKDIR /bin
COPY /linux /bin
ENTRYPOINT alarm_service_linux