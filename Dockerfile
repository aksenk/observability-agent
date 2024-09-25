FROM golang:1.23-alpine as build
COPY . /build/
RUN cd /build/ \
 && go build

FROM alpine:3
COPY --from=build /build/observability-agent /app/
WORKDIR /app/
ENTRYPOINT ["./observability-agent"]
# TODO add a user
