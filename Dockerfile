FROM alpine:3.5

ENV GIN_MODE "release"

COPY uwsgi-monitor /uwsgi-monitor
RUN mkdir /data

EXPOSE 6080

ENTRYPOINT ["/uwsgi-monitor"]
