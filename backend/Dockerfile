FROM golang:1.27.3-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
	&& mkdir -p /out/data /out/uploads \
	&& chown -R 65532:65532 /out/data /out/uploads

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/server /server
COPY --from=build --chown=65532:65532 /out/data /data
COPY --from=build --chown=65532:65532 /out/uploads /uploads

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
