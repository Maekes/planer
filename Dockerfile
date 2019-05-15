FROM golang:1.12.5-stretch as builder
WORKDIR $GOPATH/go/src/planer
COPY . .
RUN go get -u
EXPOSE 8080
CMD ["planer"]