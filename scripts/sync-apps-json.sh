#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
UPSTREAM_REPO="${UPSTREAM_REPO:-caiyy/Agent2API_FnOS}"
TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
WORK="$(mktemp "${TMPDIR:-/tmp}/cocode-store-feed.XXXXXX")"
trap 'rm -f "${WORK}" "${WORK}.new"' EXIT

for command in curl jq; do
    if ! command -v "${command}" >/dev/null 2>&1; then
        printf 'Required command not found: %s\n' "${command}" >&2
        exit 1
    fi
done

CURL_ARGS=(
    -fsSL
    --retry 5
    --retry-delay 3
    --retry-all-errors
    --connect-timeout 30
    -H "Accept: application/vnd.github+json"
    -H "X-GitHub-Api-Version: 2022-11-28"
)
if [ -n "${TOKEN}" ]; then
    CURL_ARGS+=(-H "Authorization: Bearer ${TOKEN}")
fi

curl "${CURL_ARGS[@]}" \
    "https://api.github.com/repos/${UPSTREAM_REPO}/releases?per_page=30" \
    -o "${WORK}"

release="$(jq -c '
    def has_asset($name): any(.assets[]?; .name == $name);
    [
        .[]
        | select(.draft == false)
        | select(.tag_name | startswith("fnos-v"))
        | select(
            has_asset("SHA256SUMS")
            and any(.assets[]?; .name | test("^agent2api_.+_x86\\.fpk$"))
            and any(.assets[]?; .name | test("^agent2api_.+_arm\\.fpk$"))
        )
    ][0] // empty
' "${WORK}")"

if [ -z "${release}" ]; then
    printf 'No compatible Agent2API fnOS release found\n' >&2
    exit 1
fi

TAG="$(jq -r '.tag_name' <<<"${release}")"
VERSION="${TAG#fnos-v}"
UPDATED_AT="$(jq -r '.published_at' <<<"${release}")"

case "${VERSION}" in
    ''|*[!0-9A-Za-z.+-]*)
        printf 'Invalid package version: %s\n' "${VERSION}" >&2
        exit 1
        ;;
esac

jq -n \
    --arg version "${VERSION}" \
    --arg release_tag "${TAG}" \
    --arg updated_at "${UPDATED_AT}" \
    '{
        apps: [
            {
                appname: "agent2api",
                display_name: "Agent2API",
                description: "AI 路由网关与 Agent 客户端接管，统一模型接口、账号路由、网关密钥与请求报表。",
                homepage_url: "https://github.com/aimod-cc/Agent2API",
                updated_at: $updated_at,
                version: $version,
                fpk_version: $version,
                release_tag: $release_tag,
                file_prefix: "agent2api",
                service_port: 3065,
                icon_url: "https://raw.githubusercontent.com/caiyy/Agent2API_FnOS/main/packaging/fnos/ICON_256.PNG",
                download_count: 0,
                app_type: "native",
                category: "ai",
                platforms: ["x86", "arm"],
                post_install_note: "首次打开请在 Agent2API 面板注册管理员，然后创建网关 Key。"
            }
        ]
    }' > "${WORK}.new"

mkdir -p "${ROOT}"
mv "${WORK}.new" "${ROOT}/apps.json"
printf 'Updated apps.json: Agent2API %s (%s)\n' "${VERSION}" "${TAG}"
