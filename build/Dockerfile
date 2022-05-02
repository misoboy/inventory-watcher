FROM alpine as certs
RUN apk update && apk add ca-certificates && apk --no-cache add tzdata

#FROM busybox
#COPY --from=certs /etc/ssl/certs /etc/ssl/certs
RUN mkdir /app
WORKDIR /app
COPY ./bin/application ./application
ENTRYPOINT ["./application"]