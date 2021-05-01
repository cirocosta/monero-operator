ARG BUILDER_IMAGE=index.docker.io/library/ubuntu@sha256:cf31af331f38d1d7158470e095b132acd126a7180a54f263d386da88eb681d93
ARG RUNTIME_IMAGE=gcr.io/distroless/base@sha256:bc84925113289d139a9ef2f309f0dd7ac46ea7b786f172ba9084ffdb4cbd9490


FROM $BUILDER_IMAGE AS builder

	ARG MONERO_VERSION=0.17.2.0
	ARG MONERO_SHA256=59e16c53b2aff8d9ab7a8ba3279ee826ac1f2480fbb98e79a149e6be23dd9086

	RUN set -ex && \
		apt update && \
		apt install -y curl bzip2

	RUN set -ex && \
		curl -SOL https://downloads.getmonero.org/cli/monero-linux-x64-v${MONERO_VERSION}.tar.bz2 && \
		echo "${MONERO_SHA256} monero-linux-x64-v${MONERO_VERSION}.tar.bz2" | sha256sum -c && \
		tar xf monero-linux-x64-v${MONERO_VERSION}.tar.bz2 --strip-components=1 && \
		mv ./monerod /usr/local/bin/monerod


FROM $RUNTIME_IMAGE

	COPY --from=builder /usr/local/bin/monerod /usr/local/bin/monerod
	ENTRYPOINT [ "monerod", "--non-interactive" ]
