FROM amazonlinux:2 as certs

# Use alpine which provides 'sh' excutable that is required by aws-sdk-go credential_process
FROM alpine
# Add certificates to this scratch image so that we can make calls to the AWS APIs
COPY --from=certs /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem

COPY ["LICENSE", "NOTICE", "THIRD-PARTY", "/"]

ADD bin/linux-amd64/local-container-endpoints /

EXPOSE 80

ENV HOME /home

CMD ["/local-container-endpoints"]
