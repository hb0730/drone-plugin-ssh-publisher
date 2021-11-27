FROM alpine
ADD drone-plugin-ssh-publisher /bin
ENTRYPOINT ["/bin/drone-plugin-ssh-publisher"]