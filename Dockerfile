FROM oracle/graalvm-ce:20.1.0-java11 AS native-builder

RUN yum install -y git
RUN gu install native-image
RUN curl -o lein https://raw.githubusercontent.com/technomancy/leiningen/stable/bin/lein \
    && chmod a+x lein \
    && cp lein /usr/local/bin/lein

WORKDIR /tmp
RUN git clone https://github.com/rinx/libhalodb

WORKDIR /tmp/libhalodb
RUN make target/native/libhalodb.so

FROM vdaas/vald-base:latest AS builder

ENV ORG rinx
ENV REPO vald-meta-halodb
ENV PKG meta/halodb
ENV APP_NAME meta

WORKDIR /tmp
RUN git clone https://github.com/vdaas/vald

WORKDIR ${GOPATH}/src/github.com/${ORG}/${REPO}
RUN cp -r /tmp/vald/internal ./ \
    && find . -type f -name "*.go" | xargs sed -i "s:vdaas/vald/internal:${ORG}/${REPO}/internal:g"

WORKDIR ${GOPATH}/src/github.com/${ORG}/${REPO}/pkg/${PKG}
COPY pkg/${PKG} .

WORKDIR ${GOPATH}/src/github.com/${ORG}/${REPO}/cmd/${PKG}
COPY cmd/${PKG} .

WORKDIR ${GOPATH}/src/github.com/${ORG}/${REPO}
COPY --from=native-builder /tmp/libhalodb/target/native native
COPY go.mod .
RUN CGO_ENABLED=1 \
    GO111MODULE=on \
    go build \
    --ldflags "-s -w -linkmode 'external'" \
    -o "${APP_NAME}" \
    "cmd/${PKG}/main.go" \
    && upx -9 -o "/usr/bin/${APP_NAME}" "${APP_NAME}"

FROM oracle/graalvm-ce:20.1.0-java11

ENV APP_NAME meta

COPY --from=native-builder /tmp/libhalodb/target/native/graal_isolate_dynamic.h /usr/local/lib
COPY --from=native-builder /tmp/libhalodb/target/native/graal_isolate.h /usr/local/lib
COPY --from=native-builder /tmp/libhalodb/target/native/libhalodb_dynamic.h /usr/local/lib
COPY --from=native-builder /tmp/libhalodb/target/native/libhalodb.h /usr/local/lib
COPY --from=native-builder /tmp/libhalodb/target/native/libhalodb.so /usr/local/lib

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /usr/bin/${APP_NAME} /go/bin/${APP_NAME}

ENV LD_LIBRARY_PATH=/usr/local/lib

ENTRYPOINT ["/go/bin/meta"]
