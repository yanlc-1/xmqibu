#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${SITE_BASE_URL:-}" ]]; then
  echo "SITE_BASE_URL is required" >&2
  exit 1
fi

if [[ -z "${INGEST_TOKEN:-}" ]]; then
  echo "INGEST_TOKEN is required" >&2
  exit 1
fi

go run ./cmd/ingest-demo \
  -mode article \
  -base-url "${SITE_BASE_URL}" \
  -token "${INGEST_TOKEN}" \
  -external-id "${EXTERNAL_ID:-demo-article-$(date +%s)}" \
  -title "${TITLE:-The Next Wave of AI Systems}" \
  -summary "${SUMMARY:-一篇通过脚本写入站点的示例文章。}" \
  -body "${BODY:-这里是示例正文，你可以在定时任务里替换成真实抓取或生成的内容。}" \
  -author "${AUTHOR:-Automation}" \
  -tags "${TAGS:-AI,Agent,Systems}" \
  -source-links "${SOURCE_LINKS:-https://example.com/reference}" \
  -published-at "${PUBLISHED_AT:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}" \
  "$@"
