ARG BUILDER_IMAGE=index.docker.io/library/ubuntu@sha256:cf31af331f38d1d7158470e095b132acd126a7180a54f263d386da88eb681d93
ARG RUNTIME_IMAGE=gcr.io/distroless/base@sha256:bc84925113289d139a9ef2f309f0dd7ac46ea7b786f172ba9084ffdb4cbd9490


FROM $BUILDER_IMAGE AS builder

	ARG XMRIG_VERSION=6.12.1
	ARG XMRIG_SHA256=2a2c80bbe62349c64fd171b1f43f8b07ccb0a1bc6f1d2a57ae182e96aeb71f72

	RUN set -ex && \
		apt update && \
		apt install -y curl

	RUN set -ex && \
		curl -SOL https://github.com/xmrig/xmrig/releases/download/v${XMRIG_VERSION}/xmrig-${XMRIG_VERSION}-linux-static-x64.tar.gz && \
		echo "${XMRIG_SHA256} xmrig-${XMRIG_VERSION}-linux-static-x64.tar.gz" | sha256sum -c && \
		tar xf xmrig-${XMRIG_VERSION}-linux-static-x64.tar.gz --strip-components=1 && \
		mv ./xmrig /usr/local/bin/xmrig


FROM $RUNTIME_IMAGE

	COPY --from=builder /usr/local/bin/xmrig /usr/local/bin/xmrig
	ENTRYPOINT [ "xmrig" ]
