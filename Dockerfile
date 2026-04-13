FROM arm64v8/golang:1.24-alpine AS builder

RUN apk add --no-cache build-base

COPY . /build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=arm64

RUN cd /build && go build -o ./mongo-backup .

FROM arm64v8/alpine:3.23

WORKDIR /root

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /build/mongo-backup ./
RUN chmod +x ./mongo-backup
RUN mkdir ./backup

RUN wget https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem -O global-bundle.pem

# Command to run the executable
ENTRYPOINT ["./mongo-backup"]
