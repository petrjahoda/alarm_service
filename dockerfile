# FROM alpine:latest
# RUN apk update && apk upgrade && apk add bash && apk add procps && apk add nano
# RUN apk add chromium
# WORKDIR /bin
# COPY /linux /bin
# ENTRYPOINT alarm_service_linux
# HEALTHCHECK CMD ps axo command | grep dll
FROM scratch
ADD /linux /
CMD ["/alarm_service_linux"]