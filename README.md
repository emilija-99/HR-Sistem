# HR Sistem

Ovaj projekat je razvijen u okviru predmeta **WEB2** na **Prirodno-matematičkom fakultetu u Kragujevcu**.
Cilj projekta je implementacija centralizovanog HR sistema sa koji omogućava upravljanje zaposlenima, njihovim odsustvima i evidencijom prisustva, uz praćenje svih promena u sistemu kroz audit log. 

Sistem je zasnovan na role-based pristupu, gde različiti tipovi korisnika (platform admin, HR admin, menadžer, zaposleni) imaju definisana prava pristupa i akcije koje mogu izvršavati.

Projekat je dostupan i kao produkciona verzija: https://hr-sistem.netlify.app/login
---

## Sadržaj

1. [Pregled funkcionalnosti](#1-pregled-funkcionalnosti)
2. [Tehnološki stack](#2-tehnološki-stack)
3. [Arhitektura](#3-arhitektura)
4. [Pokretanje projekta](#4-pokretanje-projekta)
5. [Kontejnerizacija i deployment](#11-kontejnerizacija-i-deployment)

---

## 1. Pregled funkcionalnosti

| Modul | Funkcionalnosti |
| --- | --- |
| **Autentifikacija** | Registracija, prijava, JWT access token, refresh token u httpOnly kolačiću, odjava, promena lozinke, tiho osvežavanje sesije |
| **Korisnici i uloge** | Lista korisnika, aktivacija/deaktivacija naloga, dodela uloge, pregled dozvola po ulozi |
| **Zaposleni** | Samostalno popunjavanje profila (onboarding), HR kreiranje naloga + profila, pregled svih zaposlenih, izmena, nadređeni, pozicija, departman, država |
| **Odsustva (time-off)** | Podnošenje zahteva, nacrti (DRAFT), izmena i podnošenje nacrta, otkazivanje, odobravanje/odbijanje, tipovi odsustva, preklapanje, radni dani |
| **Balans dana (leave balance)** | Ledger dana, dodela (grant), ručna korekcija, prenos (carry-over), istek, godišnji/mesečni obračun |
| **Prisustvo (attendance)** | Prijava/odjava (clock-in/out), status, lična istorija, admin pregled po zaposlenom/departmanu |
| **Audit log** | Evidencija svih važnih akcija u MongoDB, filteri po entitetu i akciji |

---

## 2. Tehnološki stack

| Sloj | Tehnologija |
| --- | --- |
| Backend | **Go 1.25**, `net/http`, `gorilla/mux`, `database/sql` + `lib/pq`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, `go-playground/validator` |
| Baze | **PostgreSQL 15** (primarna), **MongoDB 7** (audit log) |
| Frontend | **React 19**, **TypeScript**, **Vite**, **Chakra UI v3**, `react-router-dom v7` |
| Monitoring | **Prometheus**, **Grafana** |
| Kontejneri | **Podman** / Docker + `podman-compose` |
| Migracije | `golang-migrate` (`migrate/migrate`) |

---

## 3. Arhitektura

Backend je organizovan slojevito:

```
HTTP zahtev
   │
   ▼
[middleware] CORS → PrometheusMetrics → JWTAuth (opciono) → RequirePermission (opciono)
   │
   ▼
[services/<modul>/routes.go]  handler (HTTP sloj, validacija, status kodovi)
   │
   ▼
[services/<modul>/store.go]   pristup bazi (SQL / Mongo)
   │
   ▼
PostgreSQL / MongoDB
```

### 3.1 Struktura repozitorijuma

```
HR-Sistem/
├── backend/
│   ├── cmd/
│   │   ├── main.go                 # ulazna tačka: DB, Mongo, scheduler, server
│   │   └── api/api.go              # mux ruter, DI (store/handler), registracija ruta
│   ├── services/
│   │   ├── user/                   # auth + korisnici (routes.go, store.go)
│   │   ├── employee/               # zaposleni
│   │   ├── absence/                # odsustva, balans, politike, rollover
│   │   ├── attendance/             # prisustvo
│   │   ├── audit/                  # audit log (Mongo)
│   │   ├── auth/                   # bcrypt, JWT, refresh token, hash
│   │   └── scheduler/              # periodični poslovi
│   ├── middleware/                 # jwt.go, authorize.go, cors.go, prometheus.go
│   ├── types/                      # interfejsi i DTO-ovi po modulu
│   ├── utils/                      # ParseJSON, WriteSuccess/WriteError, validator
│   ├── internal/database/          # konekcija na Postgres (gorm → *sql.DB)
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/client.ts           # fetch wrapper + auto-refresh
│   │   ├── auth/                   # jwt.ts (decode), types.ts
│   │   ├── providers/AuthProvider   # stanje sesije
│   │   ├── theme/system.ts          # brend paleta + tipografija (Chakra system)
│   │   ├── styles/fonts.css         # lokalni Inter (@font-face)
│   │   ├── components/             # Layout, Menu, Brand, Account/ChangePasswordCard, ui/*
│   │   ├── pages/                  # stranice po modulima
│   │   └── routes/                 # ruter + ProtectedRoute
│   ├── public/                     # favicon.png, apple-touch-icon.png, logo.png, fonts/
│   ├── e2e/                        # Playwright testovi
│   ├── playwright.config.ts
│   └── nginx.conf                  # SPA + /api proxy
├── migrations/                     # 001..018 (up/down)
├── scripts/                        # seed-test-data.sh, run-go-tests.sh
├── deploy/                         # migrate.Dockerfile, prometheus.yml, grafana provisioning
├── podman-compose.yml
└── Makefile
```

### 3.2 Dva podsistema za pristup bazi

- **PostgreSQL** — relacioni podaci (`database/sql` + `lib/pq`). GORM se koristi isključivo za dobijanje `*sql.DB` handle-a (`gormDB.DB()`), dok svi upiti idu kroz sirovi SQL.
- **MongoDB** — audit log (kolekcija `audit_logs`)


### 3.3 Tok zahteva 

```
                            ┌───────────────────────────────┐
                            │      FRONTEND (React SPA)     │
                            │  :8080 (nginx)  ili  :5173    │
                            │   api/client.ts → fetch       │
                            └───────────────┬───────────────┘
                                            │  HTTP/1.1
              Authorization: Bearer <JWT>   │  Content-Type: application/json
              cookie: refreshToken          │  /api/v1/...
                                            ▼
        ┌─────────────────────────────────────────────────────────────────┐
        │                    GO SERVER  (:8034)                           │
        │                                                                 │
        │  [1] CORS                middleware/cors.go                     │
        │         ↓                (Origin allow-lista, preflight OPTIONS)│
        │  [2] PrometheusMetrics   middleware/prometheus.go               │
        │         ↓                (broj zahteva + trajanje, /metrics)    │
        │                                                                 │
        │  ┌───────────────────────────────────────────────────────────┐  │
        │  │ GORILLA MUX ROUTER            cmd/api/api.go              │  │
        │  └──────────────┬────────────────────────────┬───────────────┘  │
        │                 │                            │                  │
        │        /api/v1  (JAVNO)            /api/v1  (ZAŠTIĆENO)         │
        │                 │                            │                  │
        │                 │                    [3] JWTAuth                │
        │                 │                 middleware/jwt.go             │
        │                 │       (potpis/alg, is_active, sveža uloga)    │
        │                 │                            │                  │
        │                 │                    [4] RequirePermission       │
        │                 │                middleware/authorize.go         │
        │                 │             (po ruti: provera role_permissions)│
        │                 │                            │                  │
        │                 ▼                            ▼                  │
        │         ┌──────────────────────────────────────────┐           │
        │         │            HANDLER (HTTP sloj)           │           │
        │         │      services/<modul>/routes.go          │           │
        │         │  ParseJSON → validator → logika → Write  │           │
        │         └──────────────────┬───────────────────────┘           │
        │                            │                                   │
        │         ┌──────────────────┴───────────────────────┐           │
        │         │        STORE (repository sloj)           │           │
        │         │      services/<modul>/store.go           │           │
        │         │        (sirovi SQL preko *sql.DB)        │           │
        │         └────────────┬───────────────────┬─────────┘           │
        │                      │                   │                     │
        └──────────────────────┼───────────────────┼─────────────────────┘
                               ▼                   ▼
                     ┌──────────────────┐  ┌────────────────────────┐
                     │   PostgreSQL     │  │  MongoDB               │
                     │  hr_db (5432)    │  │  hrsystem_audit        │
                     │  korisnici,      │  │  kolekcija:            │
                     │  zaposleni,      │  │  audit_logs            │
                     │  odsustva,       │  └────────────────────────┘
                     │  balans, ...     │
                     └──────────────────┘
                               ▲
                               │  (paralelno, van HTTP toka)
                     ┌─────────┴───────────────────────────────┐
                     │  SCHEDULER (gorutina)                    │
                     │  services/scheduler                      │
                     │  start + svakih 1h → RolloverYear,        │
                     │  RunMonthlyAccrual, RunExpiration         │
                     │  → koristi absence store → PostgreSQL     │
                     └───────────────────────────────────────────┘
```

## 4. Pokretanje projekta

### 4.1 Preduslovi

- Podman (ili Docker) i `podman-compose`
- Go 1.25 (za lokalni razvoj bez kontejnera)
- Node.js 22+ (frontend)
- Slobodni portovi: **5433, 27017, 8034, 8080, 9090, 3000**

> Na Fedori: sistemski `mongod` mora biti zaustavljen jer zauzima port 27017:
> `sudo systemctl stop mongod && sudo systemctl disable mongod`
> Rootless Podman ne može da veže port 80, zato je frontend na **8080:80**.

### 4.2 Konfiguracija (`.env`)
`.env` u korenu se koristi za lokalno pokretanje (host):

```
DB_HOST=localhost
DB_PORT=5433
POSTGRES_USER=hr_user
POSTGRES_PASSWORD=<lozinka>
POSTGRES_DB=hr_db
DB_SSLMODE=disable
MONGO_URI=mongodb://localhost:27017
MONGO_DB=hrsystem_audit
```

U kontejnerima se `DB_HOST=db`, `DB_PORT=5432`, `MONGO_URI=mongodb://mongo:27017` zadaju kroz compose.

### 4.3 Produkciono pokretanje (baze, migracije, API, UI) 

```sh
# 1) izgradnja slika backend-a i frontend-a
podman compose build

# 2) podizanje svih servisa u pozadini
podman compose up -d
```

**Servisi i portovi:**

| Servis | Kontejner | Port (host → kontejner) |
| --- | --- | --- |
| PostgreSQL | `hr_db` | 5433 → 5432 |
| MongoDB | `hr_mongo` | 27017 → 27017 |
| Migracije | `hr-sistem_migrate_1` (one-shot) | — |
| Backend (Go API) | `hr_backend` | 8034 → 8034 |
| Frontend (nginx) | `hr_frontend` | 8080 → 80 |

- Aplikacija (UI): **http://localhost:8080**
- API: **http://localhost:8034/api/v1**

**Provera da je sve zdravo:**

```sh
podman ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
curl -s -o /dev/null -w "UI: %{http_code}\n"     http://localhost:8080/
curl -s -o /dev/null -w "API: %{http_code}\n"    http://localhost:8034/metrics
podman logs hr-sistem_migrate_1 | tail          
```

**Rebuild kontejnera:**
> ```sh
> podman compose build && podman compose down && podman compose up -d
> ```


**Zaustavljanje i reset:**

```sh
podman compose down       # zaustavi, ali čuva volumene (pgdata, mongodata)
podman compose down -v    # BRIŠE bazu, audite i dashboarde — čist start
podman logs -f hr_backend # logovi API-ja uživo
```

Reset baze od nule (migracije se ponovo izvršavaju):

```sh
podman compose down -v && podman compose up -d
```

### 4.4 Razvojni režim (dev)

U razvoju se zadržavaju **baze u kontejnerima**, a backend i frontend se pokreću
lokalno — tako rade hot-reload i debagovanje.

```sh
# 1) samo baze (Postgres na 5433, Mongo na 27017)
podman compose up -d db mongo

# 2) backend lokalno — sam učitava ../.env (godotenv) → http://localhost:8034
cd backend && go run ./cmd

# 3) frontend sa Vite dev serverom → http://localhost:5173
cd frontend && npm install && npm run dev


| Nalog | Uloga |
| --- | --- |
| `e2e-admin@hr-sistem.com` | PLATFORM_ADMIN |
| `e2e-hr@hr-sistem.com` | HR_ADMIN |
| `e2e-manager@hr-sistem.com` | MANAGER_PORTAL_ACCESS |
| `e2e-employee@hr-sistem.com` | EMPLOYEE (20 dana godišnjeg, automatski dodeljeno) |

Lozinka: `E2eTest1!`. Skripte isto postoje i kao `npm run e2e:seed` / `npm run e2e:clean`.

`Makefile` pomaže oko portova: `make restart` (zaustavi `mongod`, proveri portove,
očisti kontejnere i podigne stack).

```

## 5. Kontejnerizacija i deployment

`podman-compose.yml` definiše servise: `db`, `mongo`, `migrate`, `backend`, `frontend`, `prometheus`, `grafana` (sa healthcheck-ovima i `depends_on`).

- **Backend** (`backend/Dockerfile`): multi-stage — `golang:1.25-alpine` build → `alpine` runtime, `CGO_ENABLED=0`, izlaže `8034`.
- **Migracije** (`deploy/migrate.Dockerfile`): `migrate/migrate`, pokreće `-path=/migrations ... up` jednom po podizanju.
- **Frontend** (`frontend/Dockerfile`): `node:20` build (`npm run build`) → `nginx:alpine`, kopira `dist` i `nginx.conf`.
- **nginx** (`frontend/nginx.conf`): SPA fallback + reverse proxy `/api/` → `backend:8034`.

---

| Nalog | Uloga |
| --- | --- |
| `e2e-admin@hr-sistem.com` | PLATFORM_ADMIN |
| `e2e-hr@hr-sistem.com` | HR_ADMIN |
| `e2e-manager@hr-sistem.com` | MANAGER_PORTAL_ACCESS |
| `e2e-employee@hr-sistem.com` | EMPLOYEE (20 dana godišnjeg, automatski dodeljeno) |

Lozinka: `E2eTest1!` (`E2E_PASSWORD`). Testovi pokrivaju: prijavu (uspeh/neuspeh), role-based navigaciju i `/unauthorized`, HR kreiranje zaposlenog + nadređeni u tabeli, tok nacrt → podnošenje, pravila datuma (jedan dan prolazi, vikend odbijen), ograničenje po raspoloživim danima, pristup audit logu, rollover prozor (van decembra/januara dugme je onemogućeno) i brendiranje (Inter + brend boje). Ukupno **15 testova**.

---

| **Rollover** | Godišnji obračun: dodela + prenos + istek |
| **Scheduler** | Pozadinski poslovi (godišnji/mesečni obračun, istek) |
| **Audit log** | Nepromenljiva evidencija akcija (MongoDB) |
| **Idempotencija** | Višestruko pokretanje istog posla ne menja rezultat |
