FROM registry.access.redhat.com/ubi8/ubi-minimal@sha256:3364fa3bd7a5aea2932e352a446ec7f46378885f6a8f03847acc64dd100aa4cc

ENV OPERATOR=/usr/local/bin/ibm-cert-manager-operator \
    USER_UID=1001 \
    USER_NAME=ibm-cert-manager-operator

# Add licenses folder
RUN mkdir /licenses
COPY LICENSE /licenses

# install operator binary
COPY build/_output/bin/ibm-cert-manager-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

LABEL name="ibm-cert-manager-operator"
LABEL vendor="IBM"
LABEL version="0.0.1"
LABEL release="0.0.1"
LABEL summary="Operator for the cert-manager microservice"
LABEL description="Operator for the cert-manager-microservice"
