// ==========================
// Chart用変数
// ==========================

let cpuChart;
let cpuData = [];
let labels = [];

let processData = [];
let currentPage = 1;
const rowsPerPage = 20;

// ==========================
// Chart初期化
// ==========================

export function initChart() {
  const canvas = document.getElementById("cpuChart");
  if (!canvas) return;

  const ctx = canvas.getContext("2d");

  cpuChart = new Chart(ctx, {
    type: "line",
    data: {
      labels: labels,
      datasets: [
        {
          label: "CPU %",
          data: cpuData,
          tension: 0.3,
        },
      ],
    },
    options: {
      responsive: true,
      animation: false,
      scales: {
        y: {
          min: 0,
          max: 100,
        },
      },
    },
  });
}

// ==========================
// CPU / Memory / Disk / Uptime
// ==========================

export async function updateSystem() {
  try {
    const cpu = await fetch("/api/cpu").then((r) => r.json());
    document.getElementById("cpu").textContent = cpu.cpu.toFixed(2);

    // ===== グラフ更新処理 =====
    if (cpuChart) {
      const now = new Date().toLocaleTimeString();

      cpuData.push(cpu.cpu);
      labels.push(now);

      // 最大30件まで保持（約60秒）
      if (cpuData.length > 30) {
        cpuData.shift();
        labels.shift();
      }

      cpuChart.update();
    }

    const mem = await fetch("/api/memory").then((r) => r.json());
    document.getElementById("memory").textContent = mem.memory.toFixed(2);

    const disk = await fetch("/api/disk").then((r) => r.json());
    document.getElementById("disk").textContent = disk.disk.toFixed(2);

    const uptime = await fetch("/api/uptime").then((r) => r.json());
    document.getElementById("uptime").textContent = formatUptime(uptime.uptime);
  } catch (e) {
    console.error("System update error:", e);
  }
}

function formatUptime(seconds) {
  const sec = Math.floor(seconds);
  const days = Math.floor(sec / 86400);
  const hours = Math.floor((sec % 86400) / 3600);
  const minutes = Math.floor((sec % 3600) / 60);
  return `${days}d ${hours}h ${minutes}m`;
}

// ==========================
// DF
// ==========================

export async function loadDF() {
  try {
    const response = await fetch("/api/df");
    if (!response.ok) return;

    const data = await response.json();
    const table = document.getElementById("dfTable");
    if (!table) return;

    table.innerHTML = "";

    data.forEach((item) => {
      const row = `
        <tr>
          <td>${item.filesystem}</td>
          <td>${item.size}</td>
          <td>${item.used}</td>
          <td>${item.avail}</td>
          <td>${item.usePercent}</td>
          <td>${item.mounted}</td>
        </tr>
      `;
      table.innerHTML += row;
    });
  } catch (err) {
    console.error("DF load error:", err);
  }
}

// ==========================
// プロセス取得&テーブル描画
// ==========================

export function initPagination() {
  const prevBtn = document.getElementById("prevBtn");
  const nextBtn = document.getElementById("nextBtn");

  if (!prevBtn || !nextBtn) return;

  prevBtn.addEventListener("click", () => {
    if (currentPage > 1) {
      currentPage--;
      renderProcessTable();
    }
  });

  nextBtn.addEventListener("click", () => {
    const totalPages = Math.ceil(processData.length / rowsPerPage);
    if (currentPage < totalPages) {
      currentPage++;
      renderProcessTable();
    }
  });
}

export async function loadProcesses() {
  try {
    const response = await fetch("/api/processes");
    if (!response.ok) return;

    const data = await response.json();

    // ★ 上位50件だけに制限
    const top50 = data.slice(0, 50);

    renderProcessTable(top50);
  } catch (err) {
    console.error("Process load error:", err);
  }
}

function renderProcessTable(processList) {
  const table = document.getElementById("processTable");
  if (!table) return;

  table.innerHTML = "";

  processList.forEach((proc) => {
    const row = `
      <tr>
        <td>${proc.pid}</td>
        <td>${proc.user}</td>
        <td>${proc.cpu}</td>
        <td>${proc.mem}</td>
        <td>${proc.command}</td>
      </tr>
    `;
    table.innerHTML += row;
  });
}
