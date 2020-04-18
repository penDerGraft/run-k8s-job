FROM golang:1.14.1-alpine AS builder

ENV GO111MODULE=on CGO_ENABLED=0

WORKDIR /go/src/github.com/InVisionApp/run-k8s-job

COPY . .

RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -installsuffix cgo \
  -tags netgo \
  -o /bin/action .


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/action /bin/action

ENTRYPOINT ["/bin/action"]
