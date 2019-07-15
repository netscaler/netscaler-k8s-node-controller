FROM golang:alpine AS builder
#RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/
COPY cmd/citrix-node-controller/k8sInterface.go .
COPY cmd/citrix-node-controller/netScalerInterface.go .
COPY cmd/citrix-node-controller/inputFeeder.go .
COPY cmd/citrix-node-controller/flannel.go .
COPY cmd/citrix-node-controller/server.go .
COPY cmd/citrix-node-controller/main.go .
COPY version/VERSION .
COPY vendor .
#RUN go get -d -v 
RUN go build -o /go/bin/citrix-node-controller

FROM alpine
COPY --from=builder /go/bin/citrix-node-controller  /go/bin/citrix-node-controller 
EXPOSE 8080
ENTRYPOINT ["/go/bin/citrix-node-controller"]

