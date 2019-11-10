FROM golang

RUN groupadd -r app -g 901
RUN useradd -m -u 901 -r -g 901 app
WORKDIR /home/app

# whack in local wysteria files
ADD vendor vendor
ADD server server
ADD common common
ADD go.mod go.mod
Add go.sum go.sum

RUN chown -R app:app /home/app/
USER app

# build server
RUN go build -o server server/*.go 

# expose the nats port
EXPOSE 4222

# expose the grpc default port
EXPOSE 31000

# expose health port
EXPOSE 8150

# WYSTERIA_SERVER_INI: where to find config file

ENTRYPOINT ["/home/app/server"]
