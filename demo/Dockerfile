ARG PLAYWRIGHT_BASE_IMAGE=mcr.microsoft.com/playwright:v1.56.1-noble
FROM ${PLAYWRIGHT_BASE_IMAGE}

ARG PLAYWRIGHT_VERSION=1.56.1

RUN apt-get update \
  && apt-get install -y --no-install-recommends ffmpeg \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /runner
RUN npm init -y >/dev/null \
  && npm install --silent "playwright@${PLAYWRIGHT_VERSION}" >/dev/null

COPY record_screen.sh /runner/record_screen.sh
