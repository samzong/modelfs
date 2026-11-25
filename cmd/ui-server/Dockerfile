FROM node:20 AS webbuild
WORKDIR /app/ui
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN corepack enable && corepack prepare pnpm@latest --activate
COPY ui/ .
ENV CI=true
RUN pnpm install --frozen-lockfile && pnpm build

FROM golang:1.25 AS gobuild
WORKDIR /src
COPY . .
RUN rm -rf pkg/ui/static/dist && mkdir -p pkg/ui/static/dist
COPY --from=webbuild /app/ui/dist/ pkg/ui/static/dist/
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o /ui-server ./cmd/ui-server

FROM gcr.io/distroless/static:nonroot
COPY --from=gobuild /ui-server /ui-server
EXPOSE 8080
ENTRYPOINT ["/ui-server"]
