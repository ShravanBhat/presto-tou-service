# Presto Time-of-Use (TOU) Pricing Service

A robust, production-ready microservice built in Go for managing and retrieving Time-of-Use (TOU) electricity pricing for individual Electric Vehicle (EV) chargers.

## Objective

Energy providers use TOU pricing to vary electricity costs based on the time of day, incentivizing off-peak usage. This service provides the backend infrastructure to store, update, and accurately query these pricing schedules for any specific EV charger, natively handling complexities like local time zones and concurrent administrative updates.

---

## 🏗️ Architecture & Key Design Choices

The service is built with reliability, consistency, and high concurrency in mind. Below are the key design decisions:

### 1. Layered (Clean) Architecture
The codebase strictly adheres to a domain-driven, layered architecture to separate concerns and improve testability:
*   **Domain**: Defines core business models and interfaces (`Charger`, `TOUSchedule`). Completely decoupled from HTTP or Database logic.
*   **Repository (`postgres_repo.go`)**: Handles all SQL interactions.
*   **Service (`pricing_service.go`)**: Contains pure business logic and complex validations.
*   **Handler (`api.go`)**: Responsible solely for HTTP routing, JSON marshaling, and error translation. 

### 2. Timezone-Aware Querying
EV chargers exist across different geographical locations. 
*   **Design**: Timezones are stored directly on the `Chargers` relational model (e.g., `America/Los_Angeles`). 
*   When a client requests the price for a specific global UTC timestamp, the service actively translates that UTC time into the specific charger's local timezone before querying the database. This guarantees that "peak hours" always map correctly to the charger's physical location.

### 3. Database Integrity & Exclusion Constraints
Preventing overlapping schedules at the database level is notoriously difficult with standard time structures.
*   **Design**: The PostgreSQL schema utilizes `btree_gist` extension to enforce an `EXCLUDE` constraint. 
*   Schedules are required to split at midnight (`00:00:00`). This constraint prevents overlapping pricing periods natively in the database, avoiding subtle bugs where a charger might unintentionally have two different prices at the exact same minute.

### 4. Advanced Concurrency Control
Updating pricing schedules, especially bulk updates, is highly susceptible to race conditions.
*   **Pessimistic Locking**: When modifying a charger's schedule, the repository utilizes `SELECT ... FOR UPDATE` to lock the rows. This guarantees that parallel requests to update the same charger don't result in fragmented or duplicate schedules.
*   **Deadlock Prevention**: In the bulk updater, before acquiring row locks, the service explicitly sorts the Charger IDs (`ORDER BY id FOR UPDATE`). This deterministic locking order absolutely prevents database deadlocks when two concurrent bulk-updates target overlapping sets of chargers.

### 5. Strict Data Validation
*   The Service layer implements a robust mathematical validation (`validateSchedules`) that converts all time blocks into minutes spanning a 24-hour continuum (0 to 1440 minutes).
*   It asserts that every schedule strictly covers the entire day without a single minute missing, and without any overlapping boundaries, guaranteeing 100% price coverage.

---

## 🔌 API Reference

### Manage Pricing Schedules
*   `GET /chargers/{id}/schedules` - Retrieve the 24-hour pricing schedule for a charger.
*   `PUT /chargers/{id}/schedules` - Completely replace the schedules for a specific charger.
*   `PATCH /chargers/{id}/schedules` - Update a specific time block on a charger's schedule.
*   `PUT /schedules/bulk` - Perform an atomic, transacted bulk update across multiple chargers simultaneously.

### Query Real-Time Price
*   `GET /chargers/{id}/price?timestamp={ISO8601_UTC}` - Fetch the precise `$/kWh` cost for a charger at a specific point in time, evaluated in its local timezone.

---

## 🚀 Getting Started

### Prerequisites
*   Go 1.22+
*   PostgreSQL 14+ (Ensure `btree_gist` extension is enabled)

### Running the Service

1. Clone the repository
2. Apply the database schema: `psql -U your_user -d your_db -f schema.sql`
3. Set your environment variables (e.g., `DATABASE_URL`)
4. Run the API:
   ```bash
   go run main.go
   ```
5. Run the tests:
   ```bash
   go test ./... -v
   ```
