FROM alpine:3

RUN apk add --update --no-cache ca-certificates

COPY ./bin/stand-schedule-policy-controller .
COPY ./config/config.json /config/config.json

ENTRYPOINT ["./stand-schedule-policy-controller"]