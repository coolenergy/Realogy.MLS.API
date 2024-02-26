FROM mongo:latest
USER root
ENV TZ UTC
COPY docker-healthcheck /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-healthcheck
HEALTHCHECK CMD ["docker-healthcheck"]