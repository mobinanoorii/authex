############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
RUN go install github.com/goreleaser/goreleaser@latest
RUN apk add --no-cache git
# checkout the project
WORKDIR /builder
COPY . .
# Fetch dependencies.
RUN go get -d -v
# Build the binary.
RUN goreleaser build --single-target --config .github/.goreleaser.yaml --clean  --single-target --output /authex
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /authex /
# Copy the temlates folder
# COPY templates /templates
# Run the hello binary.
ENTRYPOINT [ "/authex" ]
CMD [ "server", "start" ]

