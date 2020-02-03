FROM registry.access.redhat.com/ubi8/ubi-minimal:8.1-328

ENV OPERATOR=/usr/local/bin/ibm-cert-manager-operator \
    USER_UID=1001 \
    USER_NAME=ibm-cert-manager-operator

# install operator binary
COPY build/_output/bin/ibm-cert-manager-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
