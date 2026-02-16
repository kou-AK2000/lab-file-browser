function apiGet(url) {
  return fetch(url).then((res) => {
    if (!res.ok) throw new Error("API error");
    return res.json();
  });
}

function apiText(url) {
  return fetch(url).then((res) => {
    if (!res.ok) throw new Error("API error");
    return res.text();
  });
}
