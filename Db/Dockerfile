FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    g++ \
    make \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY . .

RUN ls -la && make rebuild

EXPOSE 7432

CMD ["./Database"]
