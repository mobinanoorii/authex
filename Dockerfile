############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
ARG DOCKER_TAG=0.0.0
# checkout the project 
WORKDIR /builder
COPY . .
# Fetch dependencies.
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /authex -ldflags="-s -w -extldflags \"-static\" -X main.Version=$DOCKER_TAG"
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
CMD [ "start" ]