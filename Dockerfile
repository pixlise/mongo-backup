FROM golang:1.24-alpine AS builder

RUN apk add --no-cache libc6-compat

COPY . /build
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN cd /build && go build -o ./mongo-backup .

FROM alpine:latest

WORKDIR /root

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /build/mongo-backup ./

RUN chmod +x ./mongo-backup

# Command to run the executable
ENTRYPOINT ["./mongo-backup"]
