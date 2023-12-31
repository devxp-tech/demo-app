# this is just a common layer to use local and builds
FROM golang:1.19-alpine as base
RUN apk --no-cache update && apk add --no-cache git ca-certificates
RUN update-ca-certificates

# this layer is reponsable to execute tests in cicle-ci
FROM base as ci
WORKDIR /app/
COPY . .

# the build layer responsable to create the entrypoint
FROM ci as builder
WORKDIR /app/
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o entrypoint

# the shipment layer
FROM scratch
WORKDIR /
COPY --from=base /usr/local/share/ca-certificates /usr/local/share/ca-certificates
COPY --from=base /etc/ssl/certs /etc/ssl/certs/
COPY --from=builder /app/entrypoint .
COPY --from=builder /app/rev.txt .

ENTRYPOINT [ "/entrypoint" ]


