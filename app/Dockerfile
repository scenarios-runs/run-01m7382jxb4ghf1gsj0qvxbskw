FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /build/app .

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build --chown=65532:65532 /build/app /app/app
USER 65532:65532
ENV PORT=8000
EXPOSE 8000
CMD ["/app/app"]
