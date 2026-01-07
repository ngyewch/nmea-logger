FROM jdxcode/mise:2025.12.13 AS builder

WORKDIR /workspace
COPY .mise.toml .
RUN mise trust && mise install

WORKDIR /workspace/ui
COPY ui/.mise.toml .
RUN mise trust && mise install

WORKDIR /workspace
COPY . .
ENV CI=true
RUN mise exec -- task build-single-target

FROM scratch

WORKDIR /app
COPY --from=builder /workspace/dist/nmea-logger .
ENTRYPOINT ["./nmea-logger"]
