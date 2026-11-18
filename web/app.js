
const API = "";
const WA = "https://wa.me/5350937508";
const views = ["home", "request", "provider", "abroad"];
let PROVINCES = {};

function showView(name) {
  if (name === "how") {
    showView("home");
    setTimeout(() => document.getElementById("how")?.scrollIntoView({ behavior: "smooth" }), 50);
    return;
  }
  if (!views.includes(name)) name = "home";
  document.querySelectorAll(".panel").forEach((p) => {
    p.classList.toggle("active", p.getAttribute("data-view") === name);
  });
  history.replaceState(null, "", name === "home" ? "/" : "#" + name);
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function fillSelect(id, items, placeholder) {
  const el = document.getElementById(id);
  if (!el) return;
  el.innerHTML =
    `<option value="">${placeholder || "Seleccione…"}</option>` +
    items.map((x) => `<option value="${x}">${x}</option>`).join("");
}

function bindProvince(provinceId, municipalityId) {
  const pEl = document.getElementById(provinceId);
  const mEl = document.getElementById(municipalityId);
  if (!pEl || !mEl) return;
  pEl.addEventListener("change", () => {
    const list = PROVINCES[pEl.value] || [];
    fillSelect(municipalityId, list, "Municipio…");
  });
}

async function loadMeta() {
  try {
    const r = await fetch(API + "/api/meta");
    const m = await r.json();
    fillSelect("service_type", m.service_types || []);
    fillSelect("service_type_abroad", m.service_types || []);
    PROVINCES = m.provinces || {};
    const provinceNames = Object.keys(PROVINCES).sort((a, b) => a.localeCompare(b, "es"));
    fillSelect("province", provinceNames, "Provincia…");
    fillSelect("province_abroad", provinceNames, "Provincia…");
    fillSelect("provider_province", provinceNames, "Provincia…");
    bindProvince("province", "municipality");
    bindProvince("province_abroad", "municipality_abroad");
    bindProvince("provider_province", "provider_municipality");
    const box = document.getElementById("provider_services");
    if (box) {
      box.innerHTML = (m.service_types || [])
        .map((s) => `<label class="check"><input type="checkbox" value="${s}"> ${s}</label>`)
        .join("");
    }
  } catch (e) {
    console.warn("meta", e);
  }
}

function setNotice(id, ok, text, waUrl) {
  const el = document.getElementById(id);
  if (!el) return;
  el.className = "notice show" + (ok ? "" : " error");
  el.innerHTML =
    text +
    (waUrl
      ? `<div style="margin-top:.75rem"><a class="btn btn-wa" href="${waUrl}" target="_blank" rel="noopener">Continuar por WhatsApp</a></div>`
      : "");
}

async function postJSON(path, body) {
  const r = await fetch(API + path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(data.error || "No se pudo completar la solicitud");
  return data;
}

function municipalityLabel(province, municipality) {
  if (province && municipality) return `${province} — ${municipality}`;
  return municipality || province || "";
}

async function submitRequest(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const body = Object.fromEntries(fd.entries());
  body.from_abroad = false;
  body.municipality = municipalityLabel(body.province, body.municipality);
  try {
    const data = await postJSON("/api/requests", body);
    setNotice("notice_request", true, data.message || "Solicitud enviada. Te contactaremos pronto.", data.whatsapp_url);
    e.target.reset();
  } catch (err) {
    setNotice("notice_request", false, err.message);
  }
}

async function submitAbroad(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const body = Object.fromEntries(fd.entries());
  body.from_abroad = true;
  body.contact_name = body.local_contact_name;
  body.contact_phone = body.local_contact_phone;
  body.municipality = municipalityLabel(body.province, body.municipality);
  try {
    const data = await postJSON("/api/requests", body);
    setNotice("notice_abroad", true, data.message || "Solicitud enviada.", data.whatsapp_url);
    e.target.reset();
  } catch (err) {
    setNotice("notice_abroad", false, err.message);
  }
}

async function submitProvider(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const services = [...document.querySelectorAll("#provider_services input:checked")].map((x) => x.value);
  const zoneParts = [
    municipalityLabel(fd.get("provider_province"), fd.get("provider_municipality")),
    fd.get("zone"),
  ].filter(Boolean);
  try {
    const data = await postJSON("/api/providers", {
      name: fd.get("name"),
      phone: fd.get("phone"),
      zone: zoneParts.join(" · "),
      experience: fd.get("experience"),
      services,
    });
    setNotice("notice_provider", true, data.message || "Registro enviado.", data.whatsapp_url);
    e.target.reset();
  } catch (err) {
    setNotice("notice_provider", false, err.message);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  loadMeta();
  document.querySelectorAll("[data-nav]").forEach((el) => {
    el.addEventListener("click", (ev) => {
      ev.preventDefault();
      showView(el.getAttribute("data-nav"));
    });
  });
  document.getElementById("form_request")?.addEventListener("submit", submitRequest);
  document.getElementById("form_abroad")?.addEventListener("submit", submitAbroad);
  document.getElementById("form_provider")?.addEventListener("submit", submitProvider);
  const hash = (location.hash || "").replace("#", "");
  if (hash) showView(hash);
});
