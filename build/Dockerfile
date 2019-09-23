FROM golang:alpine AS builder
WORKDIR $GOPATH/src/citrix-node-controller-v2/
COPY cmd/ cmd
COPY version/ version
COPY vendor/ vendor
RUN go build -o /go/bin/citrix-node-controller ./cmd/citrix-node-controller/ 

FROM quay.io/chorus/chorus-kube-router:latest
COPY --from=builder /go/bin/citrix-node-controller  /go/bin/citrix-node-controller
COPY build/start.sh /go/bin/start.sh
RUN ["chmod", "+x", "/go/bin/start.sh"]
ENTRYPOINT ["sh", "/go/bin/start.sh"]

