FROM gliderlabs/alpine:3.3
ENTRYPOINT ["/bin/translator"]

COPY . /go/src/github.com/tetsusat/translator
RUN apk-install -t build-deps build-base go git mercurial \
        && cd /go/src/github.com/tetsusat/translator \
        && export GOPATH=/go \
        && go get \
        && go build -o /bin/translator \
        && cp -R playbooks / \
        && rm -rf /go \
        && apk del --purge build-deps
