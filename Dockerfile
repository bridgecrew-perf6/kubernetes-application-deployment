# build stage
FROM golang:1.11.3  AS build-env

# Set our workdir to our current service in the gopath
WORKDIR /go/src/kubernetes-services-deployment/
# Copy the current code into our workdir
COPY . .
ENV GOPATH /go/
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN cd controllers/ && swag init
RUN go build -o service-engine main/main.go

# final stage
FROM ubuntu:bionic

WORKDIR /app
COPY --from=build-env /go/src/kubernetes-services-deployment/service-engine /app/

EXPOSE 8089

CMD ./service-engine
