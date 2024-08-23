FROM docker-na-public.artifactory.swg-devops.com/hyc-cloud-private-edge-docker-local/build-images/ubi9-minimal:latest-amd64
ARG VCS_REF

ENV OPERATOR=/usr/local/bin/ibm-cert-manager-operator \
    USER_UID=1001 \
    USER_NAME=ibm-cert-manager-operator

# Add licenses folder
RUN mkdir /licenses
COPY LICENSE /licenses

# install operator binary
COPY build/_output/bin/ibm-cert-manager-operator ${OPERATOR}

ENTRYPOINT ["ibm-cert-manager-operator"]

USER ${USER_UID}

LABEL name="ibm-cert-manager-operator"
LABEL vendor="IBM"
LABEL version="0.0.1"
LABEL release="0.0.1"
LABEL summary="Operator for the cert-manager microservice"
LABEL description="Operator for the cert-manager-microservice"
LABEL org.label-schema.vcs-ref=$VCS_REF
