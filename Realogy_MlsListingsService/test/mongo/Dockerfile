FROM bitnami/mongodb:4.2.3-debian-10-r21
USER root
ENV TZ UTC
COPY docker-healthcheck /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-healthcheck
HEALTHCHECK CMD ["docker-healthcheck"]