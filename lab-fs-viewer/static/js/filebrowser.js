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

    row.innerHTML = `
            <td>${item.mode}</td>
            <td><span style="color:${getPermColor(item)};font-weight:bold">
                ${item.perm}
            </span></td>
            <td>${item.nlink}</td>
            <td>${item.owner}</td>
            <td>${item.group}</td>
            <td>${convertSize(item.size)}</td>
            <td>${item.mod_time}</td>
        `;

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

    if (item.is_dir) {
      span.onclick = () => {
        const newPath = path === "/" ? "/" + item.name : path + "/" + item.name;
        load(newPath);
      };
    } else {
      span.onclick = () => {
        const filePath =
          path === "/" ? "/" + item.name : path + "/" + item.name;
        openFile(filePath);
      };
    }

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
