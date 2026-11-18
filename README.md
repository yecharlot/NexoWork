# NEXO (repo NexoWork)

**NEXO** conecta personas que **necesitan** servicios del hogar en Cuba con personas que **pueden** realizarlos.  
También acepta solicitudes desde el **exterior** para familiares residentes en Cuba.

> *Conectamos lo que se necesita con quien lo puede hacer*  
> *El puente que une trabajo y hogar*

## Contacto

| Canal | Valor |
|-------|--------|
| Teléfono / WhatsApp | +53 50937508 |
| WhatsApp directo | https://wa.me/5350937508 |
| Email | lhmolam@gmail.com |
| Pagos | CUP · USD · Zelle · MLC |

## Arquitectura

```
┌─────────────────────────────────────────────┐
│  Frontend móvil-first (HTML/CSS/JS)         │
│  /  formularios  ·  /admin panel            │
└──────────────────┬──────────────────────────┘
                   │ HTTP /api/*
┌──────────────────▼──────────────────────────┐
│  NEXO API (Go)                              │
│  requests · providers · matches · admin     │
└──────────────────┬──────────────────────────┘
                   │ Organisms + memoria
┌──────────────────▼──────────────────────────┐
│  PrismaTec-Core                             │
│  Node · Organism Manager · RootCID · Pulse  │
│  AIP :8081 · Studio                         │
└─────────────────────────────────────────────┘
```

Cada entidad de dominio es un **Digital Organism**:

| Dominio | Nombre organismo | Contenido |
|---------|------------------|-----------|
| Solicitud | `nexo.request.<id>` | ServiceRequest + episodios |
| Proveedor | `nexo.provider.<id>` | ServiceProvider |
| Match | `nexo.match.<id>` | vínculo request↔provider |

Patrón de integración alineado con el gemelo **smart-wms / wms-twin** (`core.NewNode`, `organism.Manager`, `PutMemory` + `AppendEpisode`).

## Requisitos

- Go **1.25+**
- Red para descargar módulos (`GOPROXY=https://proxy.golang.org,direct`)

## Arranque local

```bash
git clone https://github.com/yecharlot/NexoWork.git
cd NexoWork

export NEXO_ADMIN_TOKEN="elige-un-token-seguro"
./scripts/start.sh
```

Abre:

- App NEXO: http://127.0.0.1:8080  
- Admin: http://127.0.0.1:8080/admin/  
- AIP / Studio PrismaTec: http://127.0.0.1:8081/studio/  

### Variables de entorno

| Variable | Default | Uso |
|----------|---------|-----|
| `NEXO_DATA_DIR` | `~/.local/share/nexo` | Estado PrismaTec + organismos |
| `NEXO_HTTP_ADDR` | `:8080` | API + frontend |
| `NEXO_AIP_ADDR` | `:8081` | AIP / Studio |
| `NEXO_ADMIN_TOKEN` | `nexo-admin-change-me` | Token panel admin |
| `NEXO_NODE_NAME` | `nexo-node` | Nombre del nodo Core |
| `NEXO_WEB_DIR` | `web` | Archivos estáticos |

## API (resumen)

| Método | Ruta | Auth | Descripción |
|--------|------|------|-------------|
| GET | `/api/health` | no | Salud |
| GET | `/api/meta` | no | Tipos de servicio, municipios, contacto |
| POST | `/api/requests` | no | Alta solicitud (Cuba o exterior) |
| POST | `/api/providers` | no | Alta proveedor |
| GET | `/api/requests` | admin | Listar solicitudes |
| PATCH | `/api/requests/{id}` | admin | Cambiar estado |
| GET | `/api/providers` | admin | Listar proveedores |
| GET/POST | `/api/matches` | admin | Listar / crear match |
| GET | `/api/admin/summary` | admin | Contadores |

Auth admin: cabecera `X-NEXO-Admin-Token: <token>` o `Authorization: Bearer <token>`.

### Ejemplo solicitud

```bash
curl -s -X POST http://127.0.0.1:8080/api/requests \
  -H 'Content-Type: application/json' \
  -d '{
    "service_type":"Limpieza de casa",
    "description":"Apartamento 2h",
    "municipality":"Playa",
    "contact_name":"Ana",
    "contact_phone":"+5355512345"
  }'
```


## Notificaciones (email / SMS)

Cada solicitud o registro de proveedor se guarda en el store y además:

1. Se escribe en `data/notifications.log` (siempre).
2. Se envía **email** a `lhmolam@gmail.com` si configuras SMTP.
3. Se envía **SMS** si configuras un gateway HTTP.

### Email con Gmail (recomendado)

1. Activa verificación en 2 pasos en la cuenta Gmail.
2. Crea una [contraseña de aplicación](https://myaccount.google.com/apppasswords).
3. Exporta variables antes de arrancar:

```bash
export NEXO_SMTP_HOST=smtp.gmail.com
export NEXO_SMTP_PORT=587
export NEXO_SMTP_USER=lhmolam@gmail.com
export NEXO_SMTP_PASS='tu-app-password'
export NEXO_NOTIFY_EMAIL=lhmolam@gmail.com
```

### SMS (opcional)

Si tienes un proveedor SMS con API HTTP JSON `{ "to", "message" }`:

```bash
export NEXO_SMS_URL='https://tu-gateway.example/send'
export NEXO_SMS_TOKEN='opcional'
export NEXO_NOTIFY_PHONE='+5350937508'
```

Sin SMTP/SMS configurado el sistema sigue funcionando: los datos quedan en el admin y en el log local, y el usuario puede continuar por WhatsApp.


## Despliegue en VPS simple

```bash
# 1) Instalar Go 1.25+
# 2) Clonar y compilar
git clone https://github.com/yecharlot/NexoWork.git /opt/nexo
cd /opt/nexo
go build -o bin/nexo ./cmd/nexo

# 3) Servicio systemd (ejemplo)
sudo tee /etc/systemd/system/nexo.service >/dev/null <<'UNIT'
[Unit]
Description=NEXO (PrismaTec-Core)
After=network.target

[Service]
Environment=NEXO_DATA_DIR=/var/lib/nexo
Environment=NEXO_HTTP_ADDR=:8080
Environment=NEXO_AIP_ADDR=:8081
Environment=NEXO_ADMIN_TOKEN=cambia-este-token
Environment=NEXO_WEB_DIR=/opt/nexo/web
WorkingDirectory=/opt/nexo
ExecStart=/opt/nexo/bin/nexo
Restart=on-failure
User=www-data

[Install]
WantedBy=multi-user.target
UNIT

sudo mkdir -p /var/lib/nexo
sudo systemctl daemon-reload
sudo systemctl enable --now nexo
```

Pon un reverse proxy (Caddy/Nginx) con TLS delante de `:8080` si expones a Internet.

## Seguridad

- Cambia **siempre** `NEXO_ADMIN_TOKEN` en producción.
- El panel admin no es multi-usuario; es operativo y deliberadamente simple.
- Los formularios públicos no listan datos de terceros.
- Datos persistidos bajo `NEXO_DATA_DIR` con permisos de directorio del servicio.

## Estructura del repo

```
NexoWork/
  cmd/nexo/           # binario principal
  internal/
    domain/           # modelos de negocio
    store/            # Organisms PrismaTec-Core
    httpapi/          # HTTP API + estáticos
    runtime/          # arranque nodo + servidores
  web/                # frontend NEXO + admin
  scripts/start.sh
  README.md
```

## Licencia / uso

Proyecto orientado a operación real de intermediación de servicios en Cuba.  
Motor de infraestructura: [PrismaTec-Core](https://github.com/yecharlot/PrismaTec-Core).
