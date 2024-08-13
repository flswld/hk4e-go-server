FROM ubuntu:18.04

EXPOSE 8080/tcp
EXPOSE 22222/udp

WORKDIR /standalone
COPY ./bin/standalone ./standalone
RUN chmod +x ./standalone

ENTRYPOINT ["./standalone"]
