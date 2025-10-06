FROM --platform=$BUILDPLATFORM docker.io/tonistiigi/xx:1.7.0 AS xx

FROM --platform=$BUILDPLATFORM docker.io/golang:1.25.1-alpine3.22 AS builder

RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.cloud.tencent.com|g' /etc/apk/repositories

RUN apk update && apk upgrade

# add upx for binary compression
RUN apk add --no-cache upx || echo "upx not found"

COPY --from=xx / /

ARG TARGETARCH

ARG TARGETPLATFORM

ARG PROGRAM_NAME

RUN xx-info env

ENV CGO_ENABLED=0

ENV XX_VERIFY_STATIC=1

ENV HTTPS_PROXY=
ENV https_proxy=

WORKDIR /app

COPY . .

RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct

RUN go fmt . && go mod tidy && \
    xx-go build -buildvcs=false -gcflags="all=-N -l" \
    -ldflags='-extldflags "-static" -w -s -buildid=' \
    -trimpath -o $PROGRAM_NAME && \
    xx-verify $PROGRAM_NAME && \
    { upx --best $PROGRAM_NAME || true; }

FROM alpine:latest

ARG PROGRAM_NAME

ENV PROGRAM_NAME=${PROGRAM_NAME}

WORKDIR /bin/

COPY --from=builder /app/${PROGRAM_NAME} /bin/

CMD /bin/${PROGRAM_NAME} -h

