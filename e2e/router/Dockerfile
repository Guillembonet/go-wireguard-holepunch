FROM alpine:3.15

# Install packages
RUN apk add --no-cache bash gcc musl-dev make iptables conntrack-tools tcpdump dnsmasq \
    && rm -rf /var/cache/apk/*

COPY ./docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]