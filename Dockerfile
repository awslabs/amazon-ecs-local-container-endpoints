# Use amazonlinux as the base image so that:
# - we have certificates to make calls to the AWS APIs (/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem)
# - it provides 'sh' excutable that is required by aws-sdk-go credential_process
FROM amazonlinux:2

COPY ["LICENSE", "NOTICE", "THIRD-PARTY", "/"]

ADD bin/linux-amd64/local-container-endpoints /

EXPOSE 80

ENV HOME /home

CMD ["/local-container-endpoints"]
