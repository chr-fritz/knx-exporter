FROM scratch
COPY knx-exporter /
COPY pkg/.knx-exporter.yaml /etc/.knx-exporter.yaml
EXPOSE 8080/tcp
EXPOSE 3671/udp
ENTRYPOINT ["/knx-exporter"]
CMD ["run", "--config","/etc/.knx-exporter.yaml"]