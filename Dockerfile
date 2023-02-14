# setup project and deps
FROM golang:1.20-bullseye AS init

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
FROM scratch
# Copy our static executable.
COPY --from=build /go/dinnerclub/dinnerclub /go/bin/dinnerclub
# Run the binary.
ENTRYPOINT ["/go/bin/dinnerclub"]
