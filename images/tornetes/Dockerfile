FROM golang:alpine AS builder

        WORKDIR /workspace

        COPY go.mod   go.mod
        COPY go.sum   go.sum
        COPY pkg/     pkg/
        COPY cmd/     cmd/

        RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
                go build -a -v -o tornetes ./cmd/tornetes


FROM alpine AS runtime

	RUN apk add --update tor
	COPY --from=builder /workspace/tornetes /usr/local/bin/tornetes
