FROM golang:1.22.5

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod tidy

COPY . .
RUN cd beacon && go build -o /beacon

EXPOSE 48812

ENTRYPOINT ["/beacon"]