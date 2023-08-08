# setup project and deps
FROM golang:1.21-bullseye AS init

WORKDIR /go/dinnerclub/

COPY go.mod* go.sum* ./
RUN go mod download

COPY . ./

FROM init as vet
RUN go vet ./...

# run tests
FROM init as test
RUN go test -coverprofile c.out -v ./...

# build binary
FROM init as build
ARG LDFLAGS

RUN CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" ./cmd/dinnerclub/

# runtime image
FROM scratch as runtime
# Copy our static executable.
COPY --from=build /go/dinnerclub/dinnerclub /go/bin/dinnerclub
# Expose port for publishing as web service
EXPOSE 8080
# Run the binary.
ENTRYPOINT ["/go/bin/dinnerclub"]
