let showHidden = false;
let currentPath = "/";
let currentData = [];
let currentSortKey = "name";
let currentSortOrder = "asc";

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
    .catch(() => {
      // 何も出さない（無反応設計）
    });
}

function closeModal() {
  document.getElementById("fileModal").style.display = "none";
  document.getElementById("fileContent").textContent = "";
}

/* ==============================
   描画
============================== */
function render(data, path) {
  const container = document.getElementById("list");
  container.innerHTML = "";

  const filter = document.getElementById("typeFilter").value;
  const keyword = document.getElementById("searchBox").value.toLowerCase();

  let filtered = data.filter((item) => {
    if (filter === "dir" && !item.is_dir) return false;
    if (filter === "file" && item.is_dir) return false;
    if (keyword && !item.name.toLowerCase().includes(keyword)) return false;
    return true;
  });

  filtered.sort((a, b) => {
    let valA = a[currentSortKey];
    let valB = b[currentSortKey];

    if (
      currentSortKey === "name" ||
      currentSortKey === "owner" ||
      currentSortKey === "group" ||
      currentSortKey === "mode" ||
      currentSortKey === "perm"
    ) {
      valA = valA.toString().toLowerCase();
      valB = valB.toString().toLowerCase();
    }

    if (currentSortKey === "mod_time") {
      valA = new Date(valA);
      valB = new Date(valB);
    }

    if (valA < valB) return currentSortOrder === "asc" ? -1 : 1;
    if (valA > valB) return currentSortOrder === "asc" ? 1 : -1;
    return 0;
  });

  if (filtered.length === 0) {
    const empty = document.createElement("div");
    empty.style.opacity = "0.6";
    empty.style.padding = "12px";
    empty.textContent = "This directory is empty";
    container.appendChild(empty);
    return;
  }

  filtered.forEach((item) => {
    const row = document.createElement("div");
    row.classList.add("file-row");

    if (item.is_dir) row.classList.add("dir");
    if (item.perm && item.perm.includes("x")) row.classList.add("exec");
    if (item.name.startsWith(".")) row.classList.add("hidden");
    if (item.owner === "root") row.classList.add("root-owner");

    const columns = [
      { class: "mode", value: item.mode },
      { class: "perm", value: item.perm },
      { class: "links", value: item.nlink },
      { class: "owner", value: item.owner },
      { class: "group", value: item.group },
      { class: "size", value: convertSize(item.size) },
      { class: "date", value: item.mod_time },
    ];

    columns.forEach((colData) => {
      const col = document.createElement("div");
      col.classList.add("col", colData.class);
      col.textContent = colData.value;
      row.appendChild(col);
    });

    // Name column
    const nameCol = document.createElement("div");
    nameCol.classList.add("col", "name");

    const span = document.createElement("span");
    span.textContent = item.name;

    if (!item.readable) {
      span.style.opacity = "0.4";
    } else {
      span.style.cursor = "pointer";

      if (item.is_dir) {
        span.onclick = () => {
          const newPath =
            path === "/" ? "/" + item.name : path + "/" + item.name;
          load(newPath);
        };
      } else {
        span.onclick = () => {
          const filePath =
            path === "/" ? "/" + item.name : path + "/" + item.name;
          openFile(filePath);
        };
      }
    }

    nameCol.appendChild(span);
    row.appendChild(nameCol);

    container.appendChild(row);
  });
}

/* ==============================
   APIロード
============================== */
function load(path) {
  fetch(
    "/api/list?path=" + encodeURIComponent(path) + "&show_hidden=" + showHidden,
  )
    .then((res) => {
      if (!res.ok) throw new Error("API Error: " + res.status);
      return res.json();
    })
    .then((data) => {
      currentPath = path;
      currentData = data;
      document.getElementById("path").innerText = "Current: " + path;
      document.getElementById("pathInput").value = path; // ← 追加
      render(data, path);
    })
    .catch(() => {
      document.getElementById("path").innerText = "Failed to load directory";
    });
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

  document.getElementById("closeModal").onclick = closeModal;

  document.getElementById("fileModal").addEventListener("click", (e) => {
    if (e.target.id === "fileModal") closeModal();
  });

  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") closeModal();
  });

  document.getElementById("goPathBtn").onclick = () => {
    const inputPath = document.getElementById("pathInput").value.trim();
    if (!inputPath) return;

    load(inputPath);
  };

  document.getElementById("pathInput").addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      document.getElementById("goPathBtn").click();
    }
  });

  document.querySelectorAll(".sortable").forEach((header) => {
    header.addEventListener("click", () => {
      const key = header.dataset.key;

      if (currentSortKey === key) {
        currentSortOrder = currentSortOrder === "asc" ? "desc" : "asc";
      } else {
        currentSortKey = key;
        currentSortOrder = "asc";
      }

      document.querySelectorAll(".sortable").forEach((h) => {
        h.classList.remove("asc", "desc");
      });

      header.classList.add(currentSortOrder);

      reload();
    });
  });

  document.getElementById("showHiddenToggle").onchange = (e) => {
    showHidden = e.target.checked;
    load(currentPath);
  };

  load("/");
});
