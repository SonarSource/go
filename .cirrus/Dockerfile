ARG CIRRUS_AWS_ACCOUNT
FROM ${CIRRUS_AWS_ACCOUNT}.dkr.ecr.eu-central-1.amazonaws.com/base:j17-latest

ARG GO_VERSION

USER root

ENV SONAR_SCANNER_VERSION="7.2.0.5079"

RUN wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz >/dev/null 2>&1 && \
    tar xf go${GO_VERSION}.linux-amd64.tar.gz --directory=/opt


RUN curl "https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-${SONAR_SCANNER_VERSION}.zip" -o /tmp/sonar-scanner.zip \
 && unzip -d /opt /tmp/sonar-scanner.zip \
 && mv /opt/sonar-scanner-${SONAR_SCANNER_VERSION} /opt/sonar-scanner \
 && rm /tmp/sonar-scanner.zip

USER sonarsource

ENV PATH="/opt/go/bin:/home/sonarsource/go/bin:/opt/sonar-scanner/bin:${PATH}"
