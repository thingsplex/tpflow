FROM alpine:edge

ARG TARGETARCH

COPY thingsplex/ /thingsplex/
COPY build/$TARGETARCH/tpflow /thingsplex/tpflow/

WORKDIR /thingsplex/tpflow/
CMD ["./tpflow","--config", "./var/config.json"]
