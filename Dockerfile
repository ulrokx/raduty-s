# syntax=docker/dockerfile:1
# get the base image
FROM golang:1.17-alpine
# create and set the working directory
WORKDIR /app
# copy the necessary package files
COPY go.mod ./
COPY go.sum ./
# install the packages
RUN go mod download
# put the src code files
COPY . ./
# build
RUN go build -o /raduty-s
# command that is run every time
CMD ["/raduty-s"]


