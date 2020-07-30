# FROM alpine-chrome:latest
# RUN apk update && apk upgrade && apk add bash && apk add procps && apk add nano && apk add tzdata
# RUN apk add chromium

# WORKDIR /bin
# COPY /linux /bin
# ENTRYPOINT alarm_service_linux

# FROM scratch
# ADD /linux /
# CMD ["/alarm_service_linux"]


FROM alpine:latest
RUN apk update && apk upgrade && apk add bash && apk add procps && apk add nano && apk add tzdata
# Installs latest Chromium package.
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" > /etc/apk/repositories \
    && echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories \
    && echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
    && echo "http://dl-cdn.alpinelinux.org/alpine/v3.11/main" >> /etc/apk/repositories \
    && apk upgrade -U -a \
    && apk add --no-cache \
    libstdc++ \
    chromium \
    harfbuzz \
    nss \
    freetype \
    ttf-freefont \
    wqy-zenhei \
    && rm -rf /var/cache/* \
    && mkdir /var/cache/apk

# Add Chrome as a user
RUN mkdir -p /usr/src/app \
    && adduser -D chrome \
    && chown -R chrome:chrome /usr/src/app
# Run Chrome as non-privileged
USER chrome
WORKDIR /usr/src/app

ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/

WORKDIR /bin
COPY /linux /bin
ENTRYPOINT alarm_service_linux