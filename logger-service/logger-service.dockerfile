# build a tiny docker image
FROM alpine:latest 

RUN mkdir /app
# requires a built binary
COPY loggerApp /app

CMD ["/app/loggerApp"]