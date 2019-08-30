FROM golang:1.12.5-stretch as build-env

# COPY . .

WORKDIR /go/src/github.com/Maekes

ADD https://api.github.com/repos/Maekes/planer/git/refs/heads/master version.json
RUN git clone -b master https://github.com/Maekes/planer.git

RUN apt-get install tzdata -y

WORKDIR /go/src/github.com/Maekes/planer
RUN git pull
# COPY go.mod and go.sum files to the workspace
# COPY go.mod . 
# COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
ENV GO111MODULE=on
RUN go mod download
# COPY the source code as the last step
# COPY . .
# RUN go get ./...
WORKDIR /go/src/github.com/Maekes/planer
RUN go install

# Build the binary
ENTRYPOINT planer -notls -pw=ADD_IMAP_PW_HERE
