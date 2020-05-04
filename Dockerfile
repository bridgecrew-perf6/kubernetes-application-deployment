# build stage
FROM golang:1.12.17  AS build-env

# Set our workdir to our current service in the gopath
WORKDIR /go/src/bitbucket.org/cloudplex-devs/kubernetes-services-deployment/
# Copy the current code into our workdir
COPY . .
ENV GOPATH /go/
RUN go build -o service-engine main/main.go

# final stage
FROM ubuntu:bionic


WORKDIR /app
COPY --from=build-env /go/src/bitbucket.org/cloudplex-devs/kubernetes-services-deployment/service-engine /app/

EXPOSE 8089

CMD ./service-engine
