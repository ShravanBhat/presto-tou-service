# Presto Time-of-Use (TOU) Pricing Service

A robust, production-ready microservice built in Go for managing and retrieving Time-of-Use (TOU) electricity pricing for individual Electric Vehicle (EV) chargers.

## Objective

Energy providers use TOU pricing to vary electricity costs based on the time of day, incentivizing off-peak usage. This service provides the backend infrastructure to store, update, and accurately query these pricing schedules for any specific EV charger, natively handling complexities like local time zones and concurrent administrative updates.

---

## Architecture & Key Design Choices

The service is built with reliability, consistency, and high concurrency in mind. Below are the key design decisions:

### 1. Layered (Clean) Architecture
The codebase strictly adheres to a domain-driven, layered architecture to separate concerns and improve testability:
*   **Domain**: Defines core business models and interfaces (`Charger`, `TOUSchedule`). Completely decoupled from HTTP or Database logic.
*   **Repository (`postgres_repo.go`)**: Handles all SQL interactions.
*   **Service (`pricing_service.go`)**: Contains pure business logic and complex validations.
*   **Handler (`api.go`)**: Responsible solely for HTTP routing, JSON marshaling, and error translation.
*   **Router (`router.go`)**: Wires all routes to handlers and mounts the Swagger UI.

### 2. Timezone-Aware Querying
EV chargers exist across different geographical locations.
*   **Design**: Timezones are stored directly on the `Chargers` relational model (e.g., `America/Los_Angeles`).
*   When a client requests the price for a specific global UTC timestamp, the service actively translates that UTC time into the specific charger's local timezone before querying the database. This guarantees that "peak hours" always map correctly to the charger's physical location.

### 3. Database Integrity & Exclusion Constraints
Preventing overlapping schedules at the database level is notoriously difficult with standard time structures.
*   **Design**: The PostgreSQL schema utilizes the `btree_gist` extension to enforce an `EXCLUDE` constraint using `tsrange`.
*   Overnight periods are represented by setting `end_time` to `00:00:00` (midnight), which the schema maps to `24:00:00` when computing the exclusion range. This prevents overlapping pricing periods natively in the database, eliminating bugs where a charger might have two active prices at the same moment.

### 4. Advanced Concurrency Control
Updating pricing schedules, especially in bulk, is highly susceptible to race conditions.
*   **Pessimistic Locking**: When modifying a charger's schedule, the repository uses `SELECT ... FOR UPDATE` to lock rows, guaranteeing that parallel requests to update the same charger do not result in fragmented or duplicate schedules.
*   **Deadlock Prevention**: In the bulk updater, charger IDs are sorted deterministically before acquiring locks (`ORDER BY id FOR UPDATE`). This prevents database deadlocks when two concurrent bulk-updates target overlapping sets of chargers.

### 5. Strict Data Validation
*   The service layer implements a mathematical validation (`validateSchedules`) that converts all time blocks into minutes across a 24-hour continuum (0 to 1440 minutes).
*   It asserts that every schedule covers the entire day without a single minute missing and without any overlapping boundaries, guaranteeing 100% price coverage at all times.

---

## API Reference

### Manage Pricing Schedules
*   `GET /chargers/{id}/schedules` - Retrieve the 24-hour pricing schedule for a charger.
*   `PUT /chargers/{id}/schedules` - Completely replace the schedules for a specific charger.
*   `PATCH /chargers/{id}/schedules` - Update a specific time block on a charger's schedule.
*   `POST /chargers/bulk/schedules` - Perform an atomic, transacted bulk update across multiple chargers simultaneously.

### Query Real-Time Price
*   `GET /chargers/{id}/price?timestamp={ISO8601_UTC}` - Fetch the precise `$/kWh` cost for a charger at a specific point in time, evaluated in its local timezone.

### Utility
*   `GET /health` - Health check endpoint. Returns `{"status":"ok"}` when the service is running.

---

## API Documentation (Swagger)

This service includes interactive API documentation via Swagger UI, generated using [swaggo/swag](https://github.com/swaggo/swag).

Once the service is running, the Swagger UI is accessible at:

```
http://localhost:8080/swagger/index.html
```

To regenerate the Swagger docs after making changes to handler annotations, run:

```bash
swag init
```

> Note: `swag` must be installed. Install it with:
> ```bash
> go install github.com/swaggo/swag/cmd/swag@latest
> ```

---

## Testing

The service has two independent test suites, both using Go's standard `testing` package and `net/http/httptest` — no external test frameworks required.

| Package   | Coverage                                                                                     |
|-----------|----------------------------------------------------------------------------------------------|
| `handler` | HTTP layer: input validation, correct status codes, JSON encoding, URL-encoded timestamps    |
| `service` | Business logic: schedule validation, timezone conversion, error propagation, edge cases      |

Run all tests:

```bash
go test ./... -v
```

---

## Getting Started

### Prerequisites
*   Go 1.25+
*   PostgreSQL 14+ (ensure the `btree_gist` extension is enabled)

### Environment Configuration

Copy `.env.example` to `.env` and fill in your values:

```bash
# Linux / macOS
cp .env.example .env

# Windows
copy .env.example .env
```

| Variable       | Description                  | Example                                                          |
|----------------|------------------------------|------------------------------------------------------------------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:password@localhost:5432/dbname?sslmode=disable` |

### Running the Service

1. Clone the repository.
2. Apply the database schema:
   ```bash
   psql -d your_database -f schema.sql
   ```
3. Configure your environment variables (see above).
4. Run the API:
   ```bash
   go run main.go
   ```

Dependencies are vendored in the `vendor/` directory, so no internet access is required to build.
