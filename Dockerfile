FROM alpine
COPY drone-plugin-ssh-publisher /bin
ENTRYPOINT ["/bin/drone-plugin-ssh-publisher"]