############################
# STEP 1 build executable binary
############################

FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GO_PATH/src/ucuddle-crawlers/crawler/
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /go/bin/main

############################
# STEP 2 build a small image
############################

FROM scratch
# Copy our static executable.
COPY --from=builder /go/bin/main /
# Run the binary.
CMD ["/main"]