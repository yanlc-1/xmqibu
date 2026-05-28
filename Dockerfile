FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/frontier-site .

FROM scratch

WORKDIR /

COPY --from=builder /out/frontier-site /frontier-site

EXPOSE 18092

ENTRYPOINT ["/frontier-site"]
