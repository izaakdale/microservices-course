# build a tiny docker image
FROM alpine:latest 

RUN mkdir /app
# requires a built binary
COPY brokerApp /app

CMD ["/app/brokerApp"]

EXPOSE 8090