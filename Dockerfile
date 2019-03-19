FROM scratch
ADD bin/linux-amd64/local-container-endpoints /
EXPOSE 80
CMD ["/local-container-endpoints"]
