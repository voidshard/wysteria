# -- stage 1 --
FROM golang

WORKDIR /tmp

# whack in local wysteria files
ADD vendor vendor
ADD server server
ADD common common
ADD go.mod go.mod
Add go.sum go.sum

# build server
RUN CGO_ENABLED=0 go build -o wys-server server/*.go 

# -- stage 2 --
FROM alpine:latest  

# expose the nats port
EXPOSE 4222
# expose the grpc default port
EXPOSE 31000
# expose health port
EXPOSE 8150

# add binary
RUN mkdir -p /opt/
WORKDIR /opt/
COPY --from=0 /tmp/wys-server /opt/
RUN chmod +rx -R /opt/

# switch to new non-root user
RUN adduser -D app
USER app

ENTRYPOINT ["/opt/wys-server"]

