FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

# download all dependencies. They will be cached if the go.mod and go.sum files are not changed
RUN go mod download && go mod verify

# Copy src code to the image
COPY . .

# Build the Go application and place the binary in /usr/local/bin
RUN go build -v -o /usr/local/bin/app ./...

EXPOSE 9000

CMD ["app"]
