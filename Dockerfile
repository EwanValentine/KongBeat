FROM centurylink/ca-certs
WORKDIR /app
COPY KongBeat /app/
ENTRYPOINT ["./KongBeat"]
