// ==========================
// CPU / Memory / Disk / Uptime
// ==========================

export async function updateSystem() {
  try {
    const cpu = await fetch("/api/cpu").then((r) => r.json());
    document.getElementById("cpu").textContent = cpu.cpu.toFixed(2);

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
