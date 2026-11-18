
const API = "";
const WA = "https://wa.me/5350937508";

async function loadMeta() {
  try {
    const r = await fetch(API + "/api/meta");
    const m = await r.json();
    fillSelect("service_type", m.service_types || []);
    fillSelect("service_type_abroad", m.service_types || []);
    fillSelect("municipality", m.municipalities || []);
    fillSelect("municipality_abroad", m.municipalities || []);
    const svcBox = document.getElementById("provider_services");
    if (svcBox) {
      svcBox.innerHTML = (m.service_types || []).map(s =>
        `<label class="check"><input type="checkbox" value="${s}"> ${s}</label>`
      ).join("");
    }
  } catch (e) {
    console.warn(e);
  }
}

function fillSelect(id, items) {
  const el = document.getElementById(id);
  if (!el) return;
  const cur = el.value;
  el.innerHTML = `<option value="">Seleccione…</option>` +
    items.map(x => `<option value="${x}">${x}</option>`).join("");
  if (cur) el.value = cur;
}

function showView(name) {
  document.querySelectorAll("[data-view]").forEach(v => {
    v.classList.toggle("hidden", v.getAttribute("data-view") !== name);
  });
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function setNotice(id, ok, text, waUrl) {
  const el = document.getElementById(id);
  if (!el) return;
  el.className = "notice show" + (ok ? "" : " error");
  el.innerHTML = text + (waUrl ? `<div style="margin-top:.75rem"><a class="btn btn-wa" href="${waUrl}" target="_blank" rel="noopener">Continuar por WhatsApp</a></div>` : "");
}

async function submitRequest(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const body = Object.fromEntries(fd.entries());
  body.from_abroad = false;
  try {
    const r = await fetch(API + "/api/requests", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await r.json();
    if (!r.ok) throw new Error(data.error || "Error");
    setNotice("notice_request", true, data.message || "Solicitud enviada.", data.whatsapp_url);
    e.target.reset();
  } catch (err) {
    setNotice("notice_request", false, err.message || "No se pudo enviar");
  }
}

async function submitAbroad(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const body = Object.fromEntries(fd.entries());
  body.from_abroad = true;
  body.contact_name = body.local_contact_name;
  body.contact_phone = body.local_contact_phone;
  try {
    const r = await fetch(API + "/api/requests", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await r.json();
    if (!r.ok) throw new Error(data.error || "Error");
    setNotice("notice_abroad", true, data.message || "Solicitud enviada.", data.whatsapp_url);
    e.target.reset();
  } catch (err) {
    setNotice("notice_abroad", false, err.message || "No se pudo enviar");
  }
}

async function submitProvider(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const services = [...document.querySelectorAll("#provider_services input:checked")].map(x => x.value);
  const body = {
    name: fd.get("name"),
    phone: fd.get("phone"),
    zone: fd.get("zone"),
    experience: fd.get("experience"),
    services,
  };
  try {
    const r = await fetch(API + "/api/providers", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await r.json();
    if (!r.ok) throw new Error(data.error || "Error");
    setNotice("notice_provider", true, data.message || "Registro enviado.", data.whatsapp_url);
    e.target.reset();
  } catch (err) {
    setNotice("notice_provider", false, err.message || "No se pudo enviar");
  }
}

document.addEventListener("DOMContentLoaded", () => {
  loadMeta();
  document.getElementById("form_request")?.addEventListener("submit", submitRequest);
  document.getElementById("form_abroad")?.addEventListener("submit", submitAbroad);
  document.getElementById("form_provider")?.addEventListener("submit", submitProvider);
  document.querySelectorAll("[data-go]").forEach(btn => {
    btn.addEventListener("click", () => showView(btn.getAttribute("data-go")));
  });
});
