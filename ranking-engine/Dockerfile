FROM debian:bookworm-slim AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake git build-essential ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .

RUN cmake -S . -B build -DCMAKE_BUILD_TYPE=Release \
 && cmake --build build --target ranking-server -j$(nproc)

# ── runtime image ────────────────────────────────────────────────────────────
FROM debian:bookworm-slim
COPY --from=builder /app/build/ranking-server /usr/local/bin/ranking-server
EXPOSE 8082
CMD ["ranking-server"]
