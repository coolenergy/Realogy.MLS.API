# Localstack is used to simulate aws services in local for testing.
FROM localstack/localstack:0.12.20
COPY health.sh /bin/health.sh
RUN wget -qO/bin/wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && \
    chmod +x /bin/wait-for-it.sh
RUN wget -qO/bin/jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 && \
    chmod +x /bin/jq
HEALTHCHECK --interval=3s --timeout=60s --start-period=45s CMD [ "/bin/health.sh"]

