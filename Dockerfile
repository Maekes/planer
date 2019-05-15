FROM golang:1.12.5-stretch as builder
WORKDIR /go/src/github.com/Maekes/planer
COPY . .
RUN go get -u -v
EXPOSE 8080
CMD ["planer"]