let currentPath = "/";
let currentData = [];

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

/* ==============================
   サイズ変換
============================== */
function convertSize(bytes) {
  const unit = document.getElementById("unit").value;

  switch (unit) {
    case "bit":
      return bytes * 8 + " bit";
    case "kb":
      return (bytes / 1024).toFixed(2) + " KB";
    case "mb":
      return (bytes / (1024 * 1024)).toFixed(2) + " MB";
    case "gb":
      return (bytes / (1024 * 1024 * 1024)).toFixed(2) + " GB";
    default:
      return bytes + " B";
  }
}

/* ==============================
   パーミッション色判定
============================== */
function getPermColor(item) {
  if (!item.perm) return "#ffffff";
  if (item.perm === "777") return "#ff4d4d";
  if (/[2367]/.test(item.perm)) return "#ffa500";
  if (/[1357]/.test(item.perm)) return "#00cc66";
  return "#f5c542";
}

/* ==============================
   パーミッション分解
============================== */
function explainPerm(perm) {
  if (!perm) return "Invalid permission";

  perm = perm.toString().slice(-3); // ← 後ろ3桁だけ取る

  const map = {
    0: "---",
    1: "--x",
    2: "-w-",
    3: "-wx",
    4: "r--",
    5: "r-x",
    6: "rw-",
    7: "rwx",
  };

  const u = perm[0];
  const g = perm[1];
  const o = perm[2];

  return `Owner : ${u} → ${map[u]}
Group : ${g} → ${map[g]}
Other : ${o} → ${map[o]}`;
}

/* ==============================
   ファイルを開く
============================== */
function openFile(path) {
  fetch("/api/file?path=" + encodeURIComponent(path))
    .then((res) => {
      if (!res.ok) throw new Error("Cannot open file");
      return res.text();
    })
    .then((text) => {
      document.getElementById("fileContent").textContent = text;
      document.getElementById("fileModal").style.display = "block";
    })
    .catch((err) => alert(err.message));
}

function closeModal() {
  document.getElementById("fileModal").style.display = "none";
  document.getElementById("fileContent").textContent = "";
}

/* ==============================
   描画
============================== */
function render(data, path) {
  const tbody = document.getElementById("list");
  tbody.innerHTML = "";

  const filter = document.getElementById("typeFilter").value;
  const keyword = document.getElementById("searchBox").value.toLowerCase();
  const sortBy = document.getElementById("sortBy").value;
  const sortOrder = document.getElementById("sortOrder").value;

  let filtered = data.filter((item) => {
    if (filter === "dir" && !item.is_dir) return false;
    if (filter === "file" && item.is_dir) return false;
    if (keyword && !item.name.toLowerCase().includes(keyword)) return false;
    return true;
  });

  filtered.sort((a, b) => {
    let valA, valB;

    if (sortBy === "name") {
      valA = a.name.toLowerCase();
      valB = b.name.toLowerCase();
    } else if (sortBy === "size") {
      valA = a.size;
      valB = b.size;
    } else {
      valA = new Date(a.mod_time);
      valB = new Date(b.mod_time);
    }

    if (valA < valB) return sortOrder === "asc" ? -1 : 1;
    if (valA > valB) return sortOrder === "asc" ? 1 : -1;
    return 0;
  });

  filtered.forEach((item) => {
    const row = document.createElement("tr");

    // Mode
    const modeCell = document.createElement("td");
    modeCell.textContent = item.mode;
    row.appendChild(modeCell);

    // Perm
    const permCell = document.createElement("td");
    const permSpan = document.createElement("span");
    permSpan.textContent = item.perm;
    permSpan.style.color = getPermColor(item);
    permSpan.style.fontWeight = "bold";
    permSpan.style.cursor = "pointer";

    permSpan.onclick = () => {
      alert(explainPerm(item.perm));
    };

    permCell.appendChild(permSpan);
    row.appendChild(permCell);

    // Links
    const linkCell = document.createElement("td");
    linkCell.textContent = item.nlink;
    row.appendChild(linkCell);

    // Owner
    const ownerCell = document.createElement("td");
    ownerCell.textContent = item.owner;
    row.appendChild(ownerCell);

    // Group
    const groupCell = document.createElement("td");
    groupCell.textContent = item.group;
    row.appendChild(groupCell);

    // Size
    const sizeCell = document.createElement("td");
    sizeCell.textContent = convertSize(item.size);
    row.appendChild(sizeCell);

    // Modified
    const modCell = document.createElement("td");
    modCell.textContent = item.mod_time;
    row.appendChild(modCell);

    // Name
    const nameCell = document.createElement("td");
    const span = document.createElement("span");
    span.classList.add("name-cell");
    span.style.cursor = "pointer";

    if (item.is_dir) {
      span.classList.add("dir-icon");
      span.onclick = () => {
        const newPath = path === "/" ? "/" + item.name : path + "/" + item.name;
        load(newPath);
      };
    } else {
      span.classList.add("file-icon");
      span.onclick = () => {
        const filePath =
          path === "/" ? "/" + item.name : path + "/" + item.name;
        openFile(filePath);
      };
    }

    span.textContent = item.name;
    nameCell.appendChild(span);
    row.appendChild(nameCell);

    tbody.appendChild(row);
  });
}

/* ==============================
   APIロード
============================== */
function load(path) {
  fetch("/api/list?path=" + encodeURIComponent(path))
    .then((res) => {
      if (!res.ok) throw new Error("API Error: " + res.status);
      return res.json();
    })
    .then((data) => {
      currentPath = path;
      currentData = data;
      document.getElementById("path").innerText = "Current: " + path;
      render(data, path);
    })
    .catch((err) => alert(err.message));
}

function reload() {
  render(currentData, currentPath);
}

function goUp() {
  if (currentPath === "/") return;

  let parts = currentPath.split("/");
  parts.pop();
  let newPath = parts.join("/");
  if (newPath === "") newPath = "/";
  load(newPath);
}

/* ==============================
   初期化
============================== */
window.addEventListener("DOMContentLoaded", () => {
  document.getElementById("unit").onchange = reload;
  document.getElementById("typeFilter").onchange = reload;
  document.getElementById("searchBox").oninput = reload;
  document.getElementById("sortBy").onchange = reload;
  document.getElementById("sortOrder").onchange = reload;

  document.getElementById("closeModal").onclick = closeModal;

  document.getElementById("fileModal").addEventListener("click", (e) => {
    if (e.target.id === "fileModal") closeModal();
  });

  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") closeModal();
  });

  loadSystemInfo();
  load("/");
});
