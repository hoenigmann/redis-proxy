FROM golang:onbuild
RUN mkdir /app
WORKDIR /app
ENV SRC_DIR=/go/src/github.com/hoenigmann/redis-proxy
ADD . $SRC_DIR
RUN cd $SRC_DIR; go build -o myapp; cp myapp /app/
#RUN go test -timeout 30s cmd/
