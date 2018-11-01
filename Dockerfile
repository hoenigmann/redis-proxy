FROM golang
ENV SRC_DIR=/go/src/github.com/hoenigmann/redis-proxy/
COPY . $SRC_DIR
RUN cd $SRC_DIR; go get ./...
WORKDIR /go/src/github.com/hoenigmann/redis-proxy/cmd
RUN mkdir /app
RUN go build -o myapp; cp myapp /app/
#RUN go test -timeout 30s cmd/
