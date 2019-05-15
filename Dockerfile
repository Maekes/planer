FROM golang:1.12.5-stretch as builder
WORKDIR /go/src/github.com/Maekes/planer
#COPY . .

ENV GO111MODULE=on
RUN go get http://github.com/Maekes/planer
# COPY go.mod and go.sum files to the workspace
COPY go.mod . 
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .
#RUN go get ./...
WORKDIR /go/src/github.com/Maekes/planer
RUN go install
# Build the binary
CMD ["planer"]

#FROM scratch 
#COPY /go/src/github.com/Maekes/planer/planer /go/bin/planer
#EXPOSE 80
#ENTRYPOINT ["planer"]