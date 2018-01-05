FROM golang:1.9 AS builder

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep
COPY . /go/src/github.com/aitva/gryzzly-builder
WORKDIR /go/src/github.com/aitva/gryzzly-builder
RUN dep ensure -vendor-only
RUN cd cmd/builder && go build

FROM debian:stretch

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
WORKDIR /root
COPY --from=builder /go/src/github.com/aitva/gryzzly-builder/cmd/builder/builder .
EXPOSE 8080
CMD [ "./builder" ]