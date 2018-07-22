FROM golang:latest AS build

RUN mkdir -p /go/src/github.com/zottelchin/Notenservice &&\
    curl https://glide.sh/get | sh
WORKDIR /go/src/github.com/zottelchin/Notenservice
COPY . /go/src/github.com/zottelchin/Notenservice
RUN glide install &&\
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static" -s' -o Notenservice -v .

########################################

FROM scratch

COPY frontend /var/noten/frontend
COPY example.config.yml /var/noten/config.yml
COPY --from=build /go/src/github.com/zottelchin/Notenservice/Notenservice /var/noten/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /var/noten
ENV GIN_MODE=release
EXPOSE 3412
CMD ["/var/noten/Notenservice"]
