FROM alpine:3.12
COPY knx-exporter /
ENTRYPOINT ["/knx-exporter"]
CMD ["run"]