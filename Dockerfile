# Use amazonlinux as the base image so that:
# - we have certificates to make calls to the AWS APIs (/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem)
# - it provides 'sh' excutable that is required by aws-sdk-go credential_process
# NOTE: the amazonlinux:2 base image is multi-arch, so docker should be
# able to detect the correct one to use when the image is run
FROM amazonlinux:2

COPY ["LICENSE", "NOTICE", "THIRD-PARTY", "/"]

ARG ARCH_DIR
ADD bin/$ARCH_DIR/local-container-endpoints /

EXPOSE 80

ENV HOME /home

CMD ["/local-container-endpoints"]
