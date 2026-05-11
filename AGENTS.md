# Fleet Maintenance Platform — Build Spec

## What we're building

An internal fleet maintenance platform for a trucking operation. Replaces a Google Sheets
fleet board. Single fleet, no auth, no multi-tenancy. Validate with one team, scale later.

Stack: **Go** (API) · **Next.js / TypeScript** (frontend) · **PostgreSQL** (database) ·
**DigitalOcean Spaces** (file storage) · **Claude API** (invoice parsing)

---

## Design language

- **Pitch black background** — `#0a0a0a` base, `#111111` cards, `#1a1a1a` borders
- **No color except status** — white text, gray secondary text, no decorative color
- **Status colors only** — green (ENROUTE), amber (needs attention), red (SHOP/overdue), muted (STOP/UNAVAILABLE)
- **Minimal chrome** — no shadows, no gradients, no rounded corners everywhere, flat surfaces
- **Dense tables** — this is a ops tool, not a consumer app. Information density matters
- **Monospace for numbers** — odometer, costs, unit numbers use monospace font
- Font: Inter or Geist, 13–14px base size

---

## File architecture

### Backend (Go)

```
/cmd
  main.go                  # entry point, starts server

/internal
  /db
    db.go                  # connection pool, migrations runner
    queries.go             # raw SQL query strings as constants

  /models
    truck.go               # Truck struct + related types
    trailer.go
    transponder.go
    tablet.go
    maintenance.go         # MaintenanceLog struct
    invoice.go             # parsed invoice struct (AI output)

  /handlers
    trucks.go              # CRUD handlers for trucks
    trailers.go
    transponders.go
    tablets.go
    maintenance.go         # maintenance log handlers
    invoice.go             # invoice upload + AI parsing handler
    dashboard.go           # dashboard stats handler

  /services
    invoice_parser.go      # calls Claude API, returns structured invoice data
    storage.go             # DigitalOcean Spaces upload/download
    samsara.go             # Samsara API client (odometer sync)

  /middleware
    cors.go
    logger.go

  /router
    router.go              # all routes registered here

/migrations
  001_init.sql             # base schema (init.sql)
  002_*.sql                # future migrations

.env
go.mod
go.sum
```

### Frontend (Next.js)

```
/app
  layout.tsx               # root layout, pitch black bg, font
  page.tsx                 # redirects to /dashboard

  /dashboard
    page.tsx               # fleet overview

  /trucks
    page.tsx               # truck list
    [id]/page.tsx          # single truck detail

  /trailers
    page.tsx

  /maintenance
    page.tsx               # full expense log
    new/page.tsx           # add expense manually or via invoice

  /settings
    page.tsx               # thresholds, config

/components
  /ui
    table.tsx              # base table component
    badge.tsx              # status badge
    stat-card.tsx          # dashboard stat card
    modal.tsx              # modal wrapper
    button.tsx
    input.tsx

  /trucks
    truck-table.tsx        # trucks list with status, DOT expiry, PM due
    truck-status-badge.tsx
    truck-detail.tsx       # full truck info card

  /maintenance
    maintenance-table.tsx  # paginated expense log
    maintenance-form.tsx   # manual log entry form
    invoice-upload.tsx     # drag-drop upload → AI parse → confirm → save

  /dashboard
    fleet-summary.tsx      # totals: active, in shop, overdue PM
    expense-chart.tsx      # monthly spend by category
    alerts-panel.tsx       # DOT expiring, PM overdue

/lib
  api.ts                   # typed fetch wrappers for every endpoint
  types.ts                 # TypeScript types mirroring Go models
  utils.ts                 # formatters: currency, dates, odometer
  constants.ts             # PM thresholds, status colors

/styles
  globals.css              # base styles, CSS variables
```

---

## Database schema

Five tables. No extra complexity.

| Table              | Purpose                                                                           |
| ------------------ | --------------------------------------------------------------------------------- |
| `trucks`           | Core fleet registry. Includes DOT expiry, next PM/oil change odometer, Samsara ID |
| `trailers`         | Trailer registry, separate from trucks                                            |
| `transponders`     | Transponder assignments per truck                                                 |
| `tablets`          | AT&T tablet inventory per truck                                                   |
| `maintenance_logs` | Every expense ever. Linked to truck or trailer. Includes invoice file URL         |

Key design decisions:

- `next_pm_odometer` and `next_oil_change_odometer` live directly on `trucks` — updated when a PM/oil change is logged
- No separate PM schedule table — PM history lives in `maintenance_logs` filtered by category
- `maintenance_logs` supports both truck and trailer entries via nullable FKs with a CHECK constraint

---

## API endpoints

### Trucks

```
GET    /api/trucks              list all trucks (with status, DOT expiry, PM due calc)
GET    /api/trucks/:id          single truck + recent maintenance
POST   /api/trucks              create truck
PATCH  /api/trucks/:id          update truck (status, notes, next PM odometer, etc.)
DELETE /api/trucks/:id          soft delete (sets active = false)
```

### Trailers

```
GET    /api/trailers
GET    /api/trailers/:id
POST   /api/trailers
PATCH  /api/trailers/:id
```

### Transponders

```
GET    /api/transponders
POST   /api/transponders
PATCH  /api/transponders/:id
```

### Tablets

```
GET    /api/tablets
POST   /api/tablets
PATCH  /api/tablets/:id
```

### Maintenance

```
GET    /api/maintenance         paginated, filterable by truck/category/date range
GET    /api/maintenance/:id
POST   /api/maintenance         create log entry
PATCH  /api/maintenance/:id
DELETE /api/maintenance/:id
```

### Invoice

```
POST   /api/invoice/parse       upload file → Claude extracts fields → returns JSON preview
POST   /api/invoice/confirm     user confirms parsed data → saves to maintenance_logs
```

### Dashboard

```
GET    /api/dashboard/summary   fleet counts, total spend YTD, trucks in shop
GET    /api/dashboard/alerts    DOT expiring <30 days, PM overdue trucks
GET    /api/dashboard/expenses  weekly/monthly spend grouped by category
```

---

## Invoice parsing flow

This is the flagship feature. Flow:

```
1. User drags invoice file (PDF or image) onto upload area
2. Frontend sends file to POST /api/invoice/parse
3. Go handler uploads file to DO Spaces, gets a URL
4. Go handler sends file + URL to Claude API with a structured prompt
5. Claude returns JSON: { date, vendor, truck_unit, driver, category, amount, description, reference_number }
6. Frontend shows parsed fields in an editable confirmation form
7. User reviews, corrects if needed, hits confirm
8. POST /api/invoice/confirm saves the maintenance_log row with invoice_file_url attached
```

Claude prompt for invoice parsing (in `invoice_parser.go`):

```
You are parsing a trucking fleet maintenance invoice.
Extract the following fields and return ONLY valid JSON, no prose:
{
  "date": "YYYY-MM-DD or null",
  "vendor": "string or null",
  "truck_unit": "unit number like 071 or null",
  "driver_name": "string or null",
  "amount": number or null,
  "category": one of [PM Service, Oil change, Tire issue, Engine issue, Towing,
                      Road Service, Body work, Leakage, Electrical issue,
                      Fluids/Truck Parts, Brakes/Drums/Rotors, Other],
  "description": "brief description of work done",
  "reference_number": "invoice/transaction number or null"
}
If a field cannot be determined, use null.
```

---

## Dashboard — key views

### Fleet overview (main page)

- Total trucks active / in shop / unavailable
- Trucks with DOT inspection expiring in <30 days (sorted by days remaining)
- Trucks overdue for PM (current odometer > next_pm_odometer)
- Trucks overdue for oil change
- Total spend this month vs last month

### Truck list

- Table: unit number · make/year · status badge · DOT expiry · PM due (miles remaining) · last expense date
- Click row → truck detail page
- Filter by status, company, ownership

### Truck detail

- Header: unit, VIN, make, year, plate, company, ownership, Samsara ID
- Status block: current status + note + date changed
- PM block: next PM odometer, next oil change odometer (editable inline)
- DOT inspection: expiry date + link to form
- Maintenance history: paginated table of all expenses for this truck
- Transponder + tablet info

### Maintenance log

- Full paginated table of all expenses across all trucks
- Filters: truck, category, date range, who_covers, verified status
- Each row: date · unit · driver · category · amount · description · invoice link
- Add new entry button → form or invoice upload

---

## Build order

Build in this sequence — each step is usable before the next:

```
1. DB setup          init.sql, connection pool, basic health check endpoint
2. Trucks API        full CRUD, all fields
3. Truck list UI     table, status badges, DOT expiry column
4. Maintenance API   full CRUD, filters
5. Maintenance UI    paginated log, add manual entry form
6. Dashboard API     summary + alerts endpoints
7. Dashboard UI      fleet counts, alerts panel
8. Invoice upload    DO Spaces integration + Claude parsing + confirm flow
9. Trailers          API + UI (simpler, same pattern as trucks)
10. Transponders/    API + UI (simple assignment views)
    Tablets
11. Samsara sync     pull current odometer per truck, update next PM calc
```

---

## Environment variables

```env
DATABASE_URL=postgres://...
DO_SPACES_KEY=
DO_SPACES_SECRET=
DO_SPACES_BUCKET=
DO_SPACES_REGION=
DO_SPACES_ENDPOINT=
ANTHROPIC_API_KEY=
SAMSARA_API_TOKEN=
PORT=8080
```

---

## Code style rules

- **Go**: one responsibility per file, no god files, errors always handled explicitly, no `panic` in handlers
- **TypeScript**: strict mode, no `any`, all API responses typed via `lib/types.ts`
- **Components**: one component per file, props typed inline, no prop drilling past 2 levels
- **SQL**: raw queries in `db/queries.go` as constants — no ORM, readable SQL
- **No premature abstraction**: if something is only used once, don't abstract it yet
