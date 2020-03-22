FROM golang AS builder

ENV GO111MODULE=on

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go test -v ./...

WORKDIR /app/cmd/api-snowdepth

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build

FROM alpine

RUN apk add ca-certificates

COPY --from=builder /app/cmd/api-snowdepth/api-snowdepth /app/

ENTRYPOINT ["/app/api-snowdepth"]
