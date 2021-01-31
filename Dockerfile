FROM scratch
COPY knx-exporter /
COPY pkg/.knx-exporter.yaml /etc/.knx-exporter.yaml
EXPOSE 8080
ENTRYPOINT ["/knx-exporter"]
CMD ["run", "--config","/etc/.knx-exporter.yaml"]