# SplitBill — Development Plan

## 1. Project Overview

**SplitBill** adalah aplikasi untuk membantu pengguna membagi tagihan belanja/restoran berdasarkan item yang dipesan oleh masing-masing orang.

Alur utama aplikasi:

```text
Foto Struk
    ↓
Upload Image
    ↓
OCR Processing
    ↓
Extract Receipt Data
    ↓
Item List
    ↓
User Menambahkan Anggota
    ↓
User Menentukan Siapa Memesan Apa
    ↓
Calculate Split
    ↓
Bill Summary
```

Contoh:

```text
Struk:

Nasi Goreng       Rp25.000
Mie Goreng        Rp20.000
Es Teh            Rp 5.000
Ayam Goreng       Rp30.000
--------------------------
Subtotal          Rp80.000
Tax               Rp 8.000
Service           Rp 4.000
Total             Rp92.000
```

Pengguna kemudian menentukan:

```text
Agung
- Nasi Goreng
- Es Teh

Budi
- Mie Goreng

Citra
- Ayam Goreng
```

Aplikasi kemudian menghitung:

```text
Agung  → Rp34.500
Budi   → Rp23.000
Citra  → Rp34.500
----------------
Total  → Rp92.000
```

---

# 2. Goals

## Primary Goals

* Upload foto struk.
* Membaca struk menggunakan OCR.
* Mengubah hasil OCR menjadi data terstruktur.
* Menampilkan daftar item hasil OCR.
* Membuat daftar orang yang ikut split bill.
* Menentukan item yang dipesan oleh masing-masing orang.
* Mendukung satu item dipesan oleh beberapa orang.
* Menghitung subtotal masing-masing orang.
* Membagi pajak/service charge secara proporsional.
* Menampilkan hasil akhir split bill.

## Secondary Goals

* Menyimpan history transaksi selama session masih aktif.
* Share hasil split bill.
* Generate link untuk melihat hasil split.
* Koreksi hasil OCR secara manual.
* Mendukung berbagai format struk.
* Mendukung akses bersama melalui session/link.

---

# 3. Technology Stack

## Backend

* **Language:** Go
* **Framework:** Gin
* **ORM:** GORM
* **Database:** PostgreSQL
* **Migration:** golang-migrate
* **Validation:** go-playground/validator
* **Authentication:** Tidak diperlukan untuk MVP
* **Session:** Anonymous session berbasis access token/session ID
* **API Documentation:** Swagger / OpenAPI
* **Logging:** slog / Zap
* **Configuration:** Environment Variables
* **Testing:** Go testing + Testify

## OCR

OCR dibuat sebagai service terpisah sehingga backend tidak terlalu bergantung pada implementation OCR tertentu.

Initial architecture:

```text
Go API
   |
   v
OCR Service
   |
   +-- OCR Engine
   |
   +-- Receipt Parser
```

OCR engine dapat menggunakan salah satu:

* Tesseract
* PaddleOCR
* Cloud OCR
* Vision API
* AI Vision Model

Interface OCR di Go sebaiknya dibuat abstraction:

```go
type OCRService interface {
    ExtractReceipt(image []byte) (*Receipt, error)
}
```

Dengan demikian OCR engine dapat diganti tanpa mengubah business logic utama.

---

# 4. High Level Architecture

```text
                    ┌─────────────────┐
                    │     Client      │
                    │ Web / Mobile    │
                    └────────┬────────┘
                             │
                             │ HTTP/HTTPS
                             ▼
                    ┌─────────────────┐
                    │    Go + Gin     │
                    │      API        │
                    └────────┬────────┘
                             │
             ┌───────────────┼────────────────┐
             │               │                │
             ▼               ▼                ▼
       ┌──────────┐    ┌────────────┐   ┌───────────┐
       │ PostgreSQL│    │ OCR Service│   │ File      │
       │           │    │            │   │ Storage   │
       └──────────┘    └────────────┘   └───────────┘
```

---

# 5. Project Structure

Recommended structure:

```text
splitbill/
│
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   │
│   ├── config/
│   │   └── config.go
│   │
│   ├── handler/
│   │   ├── receipt_handler.go
│   │   ├── participant_handler.go
│   │   └── split_handler.go
│   │
│   ├── service/
│   │   ├── receipt_service.go
│   │   ├── ocr_service.go
│   │   ├── participant_service.go
│   │   └── split_service.go
│   │
│   ├── repository/
│   │   ├── receipt_repository.go
│   │   ├── participant_repository.go
│   │   └── split_repository.go
│   │
│   ├── model/
│   │   ├── receipt.go
│   │   ├── receipt_item.go
│   │   ├── participant.go
│   │   └── item_assignment.go
│   │
│   ├── dto/
│   │   ├── receipt.go
│   │   ├── participant.go
│   │   └── split.go
│   │
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── recovery.go
│   │
│   └── router/
│       └── router.go
│
├── migrations/
│
├── pkg/
│   ├── calculator/
│   └── response/
│
├── tests/
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── .env.example
└── PLAN.md
```

---

# 6. Core Domain Model

## Receipt

Merepresentasikan satu struk.

```text
receipts
---------
id
receipt_number
merchant_name
transaction_date
subtotal
tax
service_charge
discount
total
image_url
ocr_raw_text
created_at
updated_at
```

---

# 7. Receipt Items

```text
receipt_items
-------------
id
receipt_id
name
quantity
unit_price
total_price
confidence
created_at
updated_at
```

Contoh:

```json
{
  "name": "Nasi Goreng",
  "quantity": 1,
  "unit_price": 25000,
  "total_price": 25000,
  "confidence": 0.96
}
```

`confidence` digunakan untuk memberikan indikasi apakah hasil OCR perlu diperiksa secara manual.

---

# 8. Participants

Orang yang mengikuti split bill.

```text
participants
------------
id
receipt_id
name
created_at
updated_at
```

Contoh:

```text
Agung
Budi
Citra
Doni
```

---

# 9. Item Assignment

Relasi antara participant dan receipt item.

```text
item_assignments
----------------
id
receipt_item_id
participant_id
quantity
amount
created_at
```

Relasi:

```text
Receipt
   │
   ├── Item 1 ──── Participant A
   │
   ├── Item 2 ──── Participant B
   │
   └── Item 3 ──── Participant A
                   Participant C
```

Dengan model ini, satu item dapat dibagi oleh beberapa orang.

---

# 10. Split Calculation

Calculation harus dilakukan berdasarkan **item**, bukan langsung membagi total struk.

Contoh:

```text
Item:
Pizza = Rp100.000

Agung = 1/2
Budi  = 1/2
```

Maka:

```text
Agung = Rp50.000
Budi  = Rp50.000
```

Untuk quantity:

```text
Pizza
quantity = 4
price = Rp100.000

Agung = 2
Budi  = 1
Citra  = 1
```

Maka:

```text
Agung = Rp50.000
Budi  = Rp25.000
Citra = Rp25.000
```

---

# 11. Tax & Service Charge

Tax dan service charge dibagi secara proporsional berdasarkan subtotal masing-masing participant.

Contoh:

```text
Subtotal      = Rp100.000
Tax           = Rp10.000
Service       = Rp5.000
Total         = Rp115.000
```

Jika:

```text
Agung = Rp60.000
Budi  = Rp40.000
```

Maka:

```text
Agung ratio = 60%
Budi ratio  = 40%
```

Tax:

```text
Agung = 60% × 10.000 = Rp6.000
Budi  = 40% × 10.000 = Rp4.000
```

Service:

```text
Agung = 60% × 5.000 = Rp3.000
Budi  = 40% × 5.000 = Rp2.000
```

Final:

```text
Agung = 60.000 + 6.000 + 3.000
      = Rp69.000

Budi = 40.000 + 4.000 + 2.000
     = Rp46.000
```

Total:

```text
Rp69.000 + Rp46.000 = Rp115.000
```

---

# 12. OCR Pipeline

OCR tidak hanya mengambil text mentah.

Pipeline:

```text
Image
  ↓
Image Preprocessing
  ↓
OCR
  ↓
Raw Text
  ↓
Receipt Parser
  ↓
Structured Receipt
  ↓
Validation
  ↓
User Correction
```

Contoh raw OCR:

```text
WARUNG MAKAN ABC

Nasi Goreng       25.000
Mie Goreng        20.000
Es Teh             5.000
Ayam Goreng       30.000

Subtotal          80.000
Tax                8.000
Service            4.000
Total              92.000
```

Parser menghasilkan:

```json
{
  "merchant_name": "WARUNG MAKAN ABC",
  "items": [
    {
      "name": "Nasi Goreng",
      "quantity": 1,
      "unit_price": 25000,
      "total_price": 25000
    },
    {
      "name": "Mie Goreng",
      "quantity": 1,
      "unit_price": 20000,
      "total_price": 20000
    }
  ],
  "subtotal": 80000,
  "tax": 8000,
  "service_charge": 4000,
  "total": 92000
}
```

---

# 13. OCR Correction

OCR tidak selalu benar.

Maka hasil OCR **tidak boleh langsung dianggap final**.

UI harus memberikan kemampuan:

```text
Nasi Gorenq    Rp25.000
     ↑
   Edit
```

User dapat mengubah:

```text
Nasi Goreng    Rp25.000
```

Field yang dapat diedit:

* Item name
* Quantity
* Unit price
* Total price
* Tax
* Service charge
* Discount
* Total

---

# 14. API Design

Base URL:

```text
/api/v1
```

## Receipt

### Upload Receipt

```http
POST /api/v1/receipts
Content-Type: multipart/form-data
```

Request:

```text
file: receipt.jpg
```

Response:

```json
{
  "id": 1,
  "status": "processing"
}
```

---

## Get Receipt

```http
GET /api/v1/receipts/:id
```

Response:

```json
{
  "id": 1,
  "merchant_name": "Warung ABC",
  "subtotal": 80000,
  "tax": 8000,
  "service_charge": 4000,
  "total": 92000,
  "items": []
}
```

---

## Update Receipt

```http
PUT /api/v1/receipts/:id
```

Digunakan untuk koreksi hasil OCR.

---

# 15. Participants API

## Add Participant

```http
POST /api/v1/receipts/:receiptId/participants
```

Request:

```json
{
  "name": "Agung"
}
```

---

## Get Participants

```http
GET /api/v1/receipts/:receiptId/participants
```

---

## Delete Participant

```http
DELETE /api/v1/receipts/:receiptId/participants/:participantId
```

---

# 16. Assignment API

## Assign Item

```http
POST /api/v1/receipts/:receiptId/assignments
```

Request:

```json
{
  "receipt_item_id": 1,
  "participant_id": 2,
  "quantity": 1
}
```

---

## Remove Assignment

```http
DELETE /api/v1/receipts/:receiptId/assignments/:assignmentId
```

---

# 17. Calculate Split

```http
POST /api/v1/receipts/:receiptId/calculate
```

Response:

```json
{
  "receipt_id": 1,
  "total": 92000,
  "participants": [
    {
      "participant_id": 1,
      "name": "Agung",
      "subtotal": 30000,
      "tax": 3000,
      "service_charge": 1500,
      "total": 34500
    },
    {
      "participant_id": 2,
      "name": "Budi",
      "subtotal": 20000,
      "tax": 2000,
      "service_charge": 1000,
      "total": 23000
    }
  ]
}
```

---

# 18. Validation Rules

Sebelum calculation dilakukan:

* Semua item harus memiliki harga valid.
* Quantity harus lebih besar dari 0.
* Semua item harus memiliki assignment.
* Total quantity yang di-assignment tidak boleh melebihi quantity item.
* Participant harus berasal dari receipt yang sama.
* Total hasil split harus sama dengan total receipt.
* Jika terdapat rounding difference, selisih harus ditangani secara deterministic.

Contoh:

```text
Receipt Total = Rp100.000

Split Result:
A = Rp33.333
B = Rp33.333
C = Rp33.333

Total = Rp99.999
```

Sisa:

```text
Rp1
```

Sisa rounding harus diberikan ke participant tertentu menggunakan aturan yang konsisten, misalnya participant dengan subtotal terbesar.

---

# 19. Split Modes

MVP:

### Mode 1 — Full Item

Satu item dimiliki satu orang.

```text
Nasi Goreng → Agung
Mie Goreng  → Budi
```

### Mode 2 — Shared Item

Satu item dibagi beberapa orang.

```text
Pizza → Agung 50%
         Budi 50%
```

### Mode 3 — Quantity Split

```text
4x Es Teh

Agung → 2
Budi  → 1
Citra → 1
```

---

# 20. Transaction State

Receipt memiliki state:

```text
UPLOADED
   ↓
PROCESSING
   ↓
OCR_COMPLETED
   ↓
REVIEW
   ↓
SPLITTING
   ↓
COMPLETED
```

Jika OCR gagal:

```text
PROCESSING
   ↓
OCR_FAILED
```

User dapat melakukan retry.

---

# 21. Error Handling

Standard response:

```json
{
  "success": false,
  "error": {
    "code": "RECEIPT_NOT_FOUND",
    "message": "Receipt not found"
  }
}
```

Contoh error code:

```text
INVALID_REQUEST
UNAUTHORIZED
RECEIPT_NOT_FOUND
OCR_FAILED
INVALID_RECEIPT
INVALID_ASSIGNMENT
INVALID_QUANTITY
SPLIT_CALCULATION_FAILED
INTERNAL_SERVER_ERROR
```

---

# 22. Security

MVP:

* JWT Authentication
* Request validation
* File type validation
* File size limitation
* Image extension validation
* Rate limiting
* CORS
* Secure HTTP headers
* SQL injection protection melalui ORM/parameterized query

Upload file harus dibatasi.

Contoh:

```text
Allowed:
jpg
jpeg
png
webp

Maximum:
10 MB
```

---

# 23. Database Relationship

```text
receipts
   │
   ├──────────────< receipt_items
   │
   └──────────────< participants
                         │
                         │
                         ▼
                  item_assignments
                         ▲
                         │
                  receipt_items
```

Logical relationship:

```text
Receipt
  ├── Items
  ├── Participants
  └── Assignments
```

---

# 24. Development Phases

## Phase 1 — Project Initialization

* [ ] Initialize Go project.
* [ ] Setup Gin.
* [ ] Setup PostgreSQL.
* [ ] Setup GORM.
* [ ] Setup migration.
* [ ] Setup environment configuration.
* [ ] Setup structured logging.
* [ ] Setup basic error handling.
* [ ] Setup health check endpoint.

Health check:

```http
GET /health
```

---

# Phase 2 — Receipt Management

* [ ] Create receipt model.
* [ ] Create receipt item model.
* [ ] Create migration.
* [ ] Implement receipt repository.
* [ ] Implement receipt service.
* [ ] Implement receipt handler.
* [ ] Implement receipt upload.
* [ ] Implement receipt detail.
* [ ] Implement receipt update.
* [ ] Implement receipt deletion.

---

# Phase 3 — OCR

* [ ] Define OCR interface.
* [ ] Implement OCR service.
* [ ] Implement image preprocessing.
* [ ] Implement OCR engine.
* [ ] Extract raw text.
* [ ] Parse merchant name.
* [ ] Parse items.
* [ ] Parse quantity.
* [ ] Parse price.
* [ ] Parse subtotal.
* [ ] Parse tax.
* [ ] Parse service charge.
* [ ] Parse discount.
* [ ] Parse total.
* [ ] Add OCR confidence score.
* [ ] Add OCR failure handling.

---

# Phase 4 — OCR Review

* [ ] Return OCR result to client.
* [ ] Display detected items.
* [ ] Allow item editing.
* [ ] Allow quantity editing.
* [ ] Allow price editing.
* [ ] Allow tax editing.
* [ ] Allow service charge editing.
* [ ] Validate corrected receipt.
* [ ] Confirm receipt.

---

# Phase 5 — Participant Management

* [ ] Create participant model.
* [ ] Create participant migration.
* [ ] Add participant.
* [ ] List participants.
* [ ] Update participant.
* [ ] Delete participant.

---

# Phase 6 — Item Assignment

* [ ] Create item assignment model.
* [ ] Create migration.
* [ ] Assign item to participant.
* [ ] Support shared item.
* [ ] Support quantity split.
* [ ] Validate assignment.
* [ ] Prevent over-assignment.
* [ ] Allow assignment modification.

---

# Phase 7 — Split Calculation

* [ ] Implement subtotal calculation.
* [ ] Implement shared item calculation.
* [ ] Implement quantity-based calculation.
* [ ] Implement proportional tax.
* [ ] Implement proportional service charge.
* [ ] Implement discount handling.
* [ ] Implement rounding strategy.
* [ ] Validate final total.
* [ ] Return participant summary.

---

# Phase 8 — History

* [ ] Create transaction history.
* [ ] List previous receipts.
* [ ] View completed split.
* [ ] Search receipt.
* [ ] Delete history.

---

# Phase 9 — Sharing

Optional feature.

* [ ] Generate share token.
* [ ] Generate public split URL.
* [ ] Public receipt summary.
* [ ] Share via WhatsApp.
* [ ] Share via Telegram.
* [ ] Copy share link.

Example:

```text
https://splitbill.app/s/ABC123
```

---

# 25. Testing Strategy

## Unit Test

Focus pada business logic.

```text
SplitCalculator
├── calculateSingleItem
├── calculateSharedItem
├── calculateQuantitySplit
├── calculateTax
├── calculateServiceCharge
└── handleRounding
```

Example:

```text
Input:

Item = Rp100.000
Participants:
A = 60%
B = 40%

Tax = Rp10.000

Expected:

A = Rp66.000
B = Rp44.000
```

---

## Integration Test

Test:

```text
API
 ↓
Service
 ↓
Repository
 ↓
PostgreSQL
```

Test cases:

* Upload receipt.
* Get receipt.
* Add participant.
* Assign item.
* Calculate split.
* Invalid assignment.
* Invalid quantity.

---

# 26. OCR Test Dataset

Buat dataset test berisi berbagai jenis struk:

```text
receipts/
├── restaurant_01.jpg
├── restaurant_02.jpg
├── cafe_01.jpg
├── minimarket_01.jpg
├── minimarket_02.jpg
├── thermal_receipt_01.jpg
├── thermal_receipt_02.jpg
└── handwritten_receipt_01.jpg
```

Setiap image memiliki expected result:

```text
receipt_01.expected.json
receipt_02.expected.json
```

Tujuannya untuk mengukur kualitas OCR/parser.

---

# 27. Observability

Backend harus memiliki:

```text
Request ID
Correlation ID
Structured Logs
Error Logs
OCR Processing Time
OCR Success Rate
Split Calculation Duration
```

Contoh:

```text
INFO
request_id=abc123
endpoint=/api/v1/receipts
ocr_duration=1.82s
status=success
```

---

# 28. Docker Development

Development environment:

```text
docker-compose
│
├── api
├── postgres
└── ocr
```

Optional:

```text
redis
```

Redis dapat digunakan nantinya untuk:

* OCR queue
* background job
* caching
* rate limiting

---

# 29. Background OCR

Untuk MVP sederhana:

```text
Upload
  ↓
API
  ↓
OCR
  ↓
Response
```

Untuk production:

```text
Upload
  ↓
API
  ↓
Database
  ↓
Queue
  ↓
OCR Worker
  ↓
Database
  ↓
Client polling / WebSocket
```

Status:

```json
{
  "status": "processing"
}
```

Kemudian:

```json
{
  "status": "completed"
}
```

---

# 30. Future Architecture

Jika aplikasi berkembang:

```text
                         ┌──────────────┐
                         │   Frontend   │
                         └──────┬───────┘
                                │
                                ▼
                         ┌──────────────┐
                         │  API Gateway │
                         └──────┬───────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
                ▼               ▼               ▼
          ┌──────────┐    ┌───────────┐   ┌──────────┐
          │ Receipt  │    │ Split Bill│   │   Auth   │
          │ Service  │    │  Service  │   │ Service  │
          └────┬─────┘    └─────┬─────┘   └──────────┘
               │                │
               └────────┬───────┘
                        ▼
                   PostgreSQL

                        │
                        ▼
                    Redis Queue
                        │
                        ▼
                   OCR Worker
                        │
                        ▼
                   OCR Engine
```

Namun untuk tahap awal **jangan langsung membuat microservices**.

Gunakan:

```text
Go + Gin
Modular Monolith
PostgreSQL
OCR Service
```

Terlebih dahulu.

---

# 31. MVP Definition

MVP dianggap selesai apabila user dapat melakukan:

```text
1. Upload foto struk
        ↓
2. OCR membaca struk
        ↓
3. User melihat hasil OCR
        ↓
4. User memperbaiki hasil OCR jika salah
        ↓
5. User memasukkan nama peserta
        ↓
6. User menentukan siapa memesan item apa
        ↓
7. User dapat membagi item
        ↓
8. System menghitung subtotal
        ↓
9. System membagi tax/service charge
        ↓
10. System menampilkan total masing-masing orang
```

Contoh final result:

```text
┌──────────────────────────────┐
│       SPLIT BILL             │
├──────────────────────────────┤
│ Warung ABC                   │
│                              │
│ Agung                        │
│ Nasi Goreng       Rp25.000   │
│ Es Teh             Rp5.000   │
│ Tax                Rp3.000   │
│ Service            Rp1.500   │
│ ───────────────────────────  │
│ Total             Rp34.500   │
│                              │
│ Budi                         │
│ Mie Goreng        Rp20.000   │
│ Tax                Rp2.000   │
│ Service            Rp1.000   │
│ ───────────────────────────  │
│ Total             Rp23.000   │
│                              │
│ TOTAL             Rp57.500   │
└──────────────────────────────┘
```

---

# 32. Definition of Done

Feature dianggap selesai apabila:

* [ ] Code mengikuti Go project structure.
* [ ] Business logic berada di service layer.
* [ ] Database access berada di repository layer.
* [ ] Handler hanya menangani HTTP concerns.
* [ ] Semua endpoint memiliki validation.
* [ ] Error response konsisten.
* [ ] Unit test tersedia untuk business logic.
* [ ] Integration test tersedia untuk critical flow.
* [ ] Tidak ada hardcoded configuration.
* [ ] Docker development environment tersedia.
* [ ] API terdokumentasi dengan OpenAPI/Swagger.
* [ ] Logging tersedia.
* [ ] OCR dapat di-retry apabila gagal.
* [ ] User dapat melakukan koreksi OCR.
* [ ] Hasil split selalu balance dengan total receipt.

---

# 33. Recommended Development Order

Urutan implementasi yang disarankan:

```text
1. Go + Gin setup
       ↓
2. PostgreSQL + Migration
       ↓
3. Receipt CRUD
       ↓
4. Receipt Item CRUD
       ↓
5. OCR Interface
       ↓
6. OCR Implementation
       ↓
7. OCR Parser
       ↓
8. OCR Review
       ↓
9. Participant
       ↓
10. Item Assignment
       ↓
11. Split Calculator
       ↓
12. Tax / Service / Discount
       ↓
13. Validation
       ↓
14. Unit Tests
       ↓
15. Integration Tests
       ↓
16. Authentication
       ↓
17. History
       ↓
18. Share Result
```

## Important Architectural Principle

Jangan membuat OCR menjadi bagian yang terlalu erat dengan Gin.

Pisahkan:

```text
Gin Handler
     ↓
Receipt Service
     ↓
OCR Service Interface
     ↓
OCR Implementation
```

Sehingga nantinya:

```text
Tesseract
```

dapat diganti menjadi:

```text
PaddleOCR
```

atau:

```text
Google Vision
```

atau:

```text
AI Vision
```

tanpa harus mengubah business logic split bill.

Hal yang paling penting dari project ini bukan hanya OCR, tetapi **receipt normalization + item assignment + split calculation**. OCR hanya menjadi input awal; user tetap harus dapat mengoreksi hasil OCR sebelum perhitungan dilakukan.
