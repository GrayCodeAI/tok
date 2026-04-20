FROM alpine:3.21
RUN apk add --no-cache git bash && \
    git config --global user.email t@e && \
    git config --global user.name t
WORKDIR /w
COPY case.sh /w/case.sh
RUN chmod +x /w/case.sh
CMD ["/w/case.sh"]
