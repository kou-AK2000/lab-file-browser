let refreshInterval = null;
let processData = [];

/* ==============================
   システム情報読み込み
============================== */
function loadSystemInfo() {
  fetch("/api/system")
    .then((res) => res.json())
    .then((data) => {
      document.getElementById("systemInfo").innerHTML = `
                <b>Hostname:</b> ${data.hostname} |
                <b>Distribution:</b> ${data.distribution} |
                <b>Kernel:</b> ${data.kernel} |
                <b>Arch:</b> ${data.arch}
            `;
    })
    .catch(() => {
      document.getElementById("systemInfo").innerText =
        "Failed to load system information";
    });
}

/**
 * モニターデータ取得
 */
function loadMonitor() {
  fetch("/api/monitor")
    .then((res) => {
      if (!res.ok) throw new Error("Monitor API error");
      return res.json();
    })
    .then((data) => {
      renderSystem(data.system);
      renderProcesses(data.processes);
    })
    .catch((err) => {
      console.error(err);
      document.getElementById("systemArea").innerText =
        "Failed to load monitor data";
    });
}

/**
 * システム情報描画
 */
function renderSystem(sys) {
  const area = document.getElementById("systemArea");

  area.innerHTML = `
        <h3>System Overview</h3>
        <b>CPU Usage:</b> ${sys.cpu_usage.toFixed(2)} % <br>
        <b>Load Avg:</b> ${sys.load_avg.join(" , ")} <br><br>

        <b>Memory Total:</b> ${formatMB(sys.mem_total)} MB <br>
        <b>Memory Used:</b> ${formatMB(sys.mem_used)} MB <br>
        <b>Memory Free:</b> ${formatMB(sys.mem_free)} MB <br>
        <b>RSS Total:</b> ${formatMB(sys.rss_total)} MB
    `;

  // コア別CPU
  if (sys.cpu_cores) {
    let coreHTML = "<h4>CPU Cores</h4>";
    sys.cpu_cores.forEach((c, i) => {
      coreHTML += `Core ${i}: ${c.toFixed(2)} %<br>`;
    });
    area.innerHTML += coreHTML;
  }
}

/**
 * プロセス一覧描画
 */
function renderProcesses(list) {
  processData = list;

  const tbody = document.getElementById("processTable");
  tbody.innerHTML = "";

  list.forEach((p) => {
    const row = document.createElement("tr");

    row.innerHTML = `
            <td>${p.pid}</td>
            <td>${p.user}</td>
            <td>${p.cpu.toFixed(2)}%</td>
            <td>${p.mem.toFixed(2)}%</td>
            <td>${formatMB(p.rss)} MB</td>
            <td>${p.command}</td>
        `;

    tbody.appendChild(row);
  });
}

/**
 * 並び替え
 */
function sortProcess(key) {
  processData.sort((a, b) => {
    if (a[key] < b[key]) return 1;
    if (a[key] > b[key]) return -1;
    return 0;
  });
  renderProcesses(processData);
}

/**
 * MB表示変換
 */
function formatMB(bytes) {
  return (bytes / 1024 / 1024).toFixed(2);
}

/**
 * 自動更新ON/OFF
 */
function toggleAutoRefresh() {
  const btn = document.getElementById("autoBtn");

  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
    btn.innerText = "Auto Refresh: OFF";
  } else {
    refreshInterval = setInterval(loadMonitor, 2000);
    btn.innerText = "Auto Refresh: ON";
  }
}

window.addEventListener("DOMContentLoaded", () => {
  document.getElementById("autoBtn").onclick = toggleAutoRefresh;
  loadMonitor();
});
