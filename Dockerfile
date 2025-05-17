FROM alpine

COPY dns-failover /

CMD ["/dns-failover"]
