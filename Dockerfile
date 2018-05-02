FROM golang:1.10.1 as builder
WORKDIR /go/src/github.com/ksonnet/ksonnet
COPY . .
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN make ks

FROM alpine:3.6
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/ksonnet/ksonnet/ks .