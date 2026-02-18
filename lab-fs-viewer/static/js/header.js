export function loadHeader() {
  fetch("/partials/header.html")
    .then((res) => res.text())
    .then((html) => {
      document.getElementById("header").innerHTML = html;
      loadSystemInfo();
    });
}
/* ==============================
   システム情報読み込み
============================== */
function loadSystemInfo() {
  fetch("/api/system")
    .then((res) => res.json())
    .then((data) => {
      const el = document.getElementById("systemInfo");
      if (!el) return;

      el.innerHTML = `
        <b>Hostname:</b> ${data.hostname} |
        <b>Distribution:</b> ${data.distribution} |
        <b>Kernel:</b> ${data.kernel} |
        <b>Arch:</b> ${data.arch}
      `;
    });
}
