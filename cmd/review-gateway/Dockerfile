FROM uhub.service.ucloud.cn/entropypool/alpine:3.16

WORKDIR /app

COPY review-gateway /app

ENTRYPOINT ["/app/review-gateway"]

EXPOSE 50450 50451 50452 50453

CMD ["run"]
