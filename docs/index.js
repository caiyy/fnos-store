const $ = (id) => document.getElementById(id);

async function loadStatus() {
  const appVersion = $("app-version");
  const storeVersion = $("store-version");
  const appUpdated = $("app-updated");
  const releaseLink = $("release-link");
  const panel = $("status-panel");

  try {
    const appsUrl = "https://raw.githubusercontent.com/caiyy/fnos-store/main/apps.json?ts=" + Date.now();
    const apps = await fetch(appsUrl).then((r) => r.json());
    const app = (apps.apps || []).find((item) => item.appname === "agent2api");
    if (!app) throw new Error("apps.json does not contain agent2api");

    appVersion.textContent = app.fpk_version || app.version || "-";
    appUpdated.textContent = app.updated_at
      ? new Date(app.updated_at).toLocaleString("zh-CN", { hour12: false })
      : "-";
    releaseLink.href = "https://github.com/caiyy/Agent2API_FnOS/releases/tag/" + app.release_tag;
    releaseLink.textContent = "查看 " + (app.release_tag || "最新发布");
  } catch (error) {
    appVersion.textContent = "需要手动确认";
    appUpdated.textContent = "GitHub 数据暂不可用";
  }

  try {
    const manifest = await fetch(
      "https://raw.githubusercontent.com/caiyy/fnos-store/main/fnos/manifest?ts=" + Date.now()
    ).then((r) => r.text());
    const match = manifest.match(/^version\s*=\s*(.+)$/m);
    storeVersion.textContent = match ? match[1].trim() : "查看 Release";
  } catch (error) {
    storeVersion.textContent = "查看 Release";
  }

  panel.classList.add("loaded");
}

loadStatus();
