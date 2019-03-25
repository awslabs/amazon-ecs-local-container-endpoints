FROM amazonlinux:2 as certs

FROM scratch
# Add certificates to this scratch image so that we can make calls to the AWS APIs
COPY --from=certs /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem
ADD bin/linux-amd64/local-container-endpoints /
EXPOSE 80
CMD ["/local-container-endpoints"]
