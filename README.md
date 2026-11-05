# go-crud-demo

ตัวอย่าง CRUD API ด้วย Go + Echo + MongoDB จัดโครงแบบ Layered Architecture
ใช้สำหรับสอนเรื่อง Layer, Constructor / Factory, Dependency Injection และ Unit Test ด้วย Fake

## วิธีรัน

```bash
# 1) เปิด MongoDB (ต้องเปิด Docker Desktop ก่อน)
docker compose up -d

# 2) เตรียมไฟล์ตั้งค่า (ทำครั้งแรกครั้งเดียว)
cp .env.example .env

# 3) รัน API (รันจาก root หรือจากใน cmd/cmd-api ก็ได้)
go run ./cmd/cmd-api

# 4) รัน test
go test ./... -race -cover
```

### API

| Method | Path | หน้าที่ |
|---|---|---|
| GET | `/health` | ตรวจว่า server ทำงาน |
| POST | `/products` | สร้างสินค้า |
| GET | `/products` | ดูสินค้าทั้งหมด |
| GET | `/products/:id` | ดูสินค้าตาม id |
| PUT | `/products/:id` | แก้ไขสินค้า |
| DELETE | `/products/:id` | ลบสินค้า |

```bash
curl -X POST localhost:8080/products -H 'Content-Type: application/json' -d '{"name":"Coffee","price":65}'
curl localhost:8080/products
curl localhost:8080/products/<id>
curl -X PUT localhost:8080/products/<id> -H 'Content-Type: application/json' -d '{"name":"Latte","price":75}'
curl -X DELETE localhost:8080/products/<id>
```

---

## โครงสร้าง

```
go-crud-demo/
├── cmd/
│   └── cmd-api/
│       └── main.go                    ① จุดเริ่มต้นของแอป
├── models/                            ② รูปแบบข้อมูลที่ทุก layer ใช้ร่วมกัน
│   ├── product.go
│   ├── request.go
│   └── response.go
├── storage/                           ③ คุยกับ MongoDB
│   └── product_storage.go
├── internal/
│   ├── api/
│   │   ├── routers/                   ④ รับและตอบ HTTP
│   │   │   ├── factory.go
│   │   │   └── product_router.go
│   │   └── services/                  ⑤ business logic
│   │       ├── factory.go
│   │       ├── product_service.go
│   │       └── product_service_test.go
│   ├── middleware/                    ⑥ ตั้งค่า Echo
│   │   ├── middleware.go
│   │   └── validator.go
│   └── pkg/                           ⑦ เครื่องมือกลาง
│       ├── connector/mongodb.go
│       └── environment/
│           ├── environment.go
│           └── env_loader.go
├── .env / .env.example                ⑧ ค่าตั้งค่า
├── docker-compose.yml                 ⑨ MongoDB สำหรับเครื่อง local
└── go.mod / go.sum
```

---

## ภาพรวม: request หนึ่งครั้งวิ่งผ่านอะไรบ้าง

ตัวอย่าง `POST /products {"name":"Coffee","price":65}`

```
Client
  │  HTTP request
  ▼
⑥ middleware      Logger → Recover → RequestID → CORS
  │
  ▼
④ routers         CreateProduct(c)
  │                 c.Bind(&request)       แปลง JSON เป็น struct
  │                 c.Validate(&request)   ตรวจ tag validate:"required,gt=0"
  ▼
⑤ services        CreateProduct(ctx, request)
  │                 ตัดช่องว่างชื่อ แล้วสร้าง models.Product
  ▼
③ storage         Insert(ctx, product)
  │                 สร้าง ObjectID แล้ว InsertOne ลง collection "products"
  ▼
MongoDB
  │
  ▲ ผลลัพธ์ส่งย้อนกลับทางเดิม
④ routers         c.JSON(201, product)
  │                 ถ้า error: ErrNotFound → 404, error อื่น → 500
  ▼
Client            {"id":"6aec...","name":"Coffee","price":65}
```

> **กฎสำคัญ:** เรียกได้ทางเดียวคือ `routers → services → storage`
> ห้ามข้าม layer (เช่น router เรียก storage ตรง) และห้ามเรียกย้อนกลับ

---

## รายละเอียดแต่ละส่วน

### ① `cmd/cmd-api/main.go`: ประกอบทุกชิ้นเข้าด้วยกัน

ไม่มี logic ใด ๆ มีแค่การสร้างและต่อชิ้นส่วนตามลำดับ

```
1. LoadProjectEnv()                  โหลด .env
2. NewMongoDBClient() + Ping()       ต่อ DB ถ้าต่อไม่ได้แอปจะหยุดทันที
3. NewProductStorage(db)             สร้าง storage จาก db
4. NewProductService(storage)        ส่ง storage เข้า service (Dependency Injection)
5. EchoMiddleware() + NewAPIRouter() สร้าง Echo แล้วผูก route
6. e.Start(":8080")                  เริ่มรับ request
```

ที่ใช้ชื่อโฟลเดอร์ `cmd/cmd-api` เพราะโปรเจกต์หนึ่งมีโปรแกรมได้หลายตัว
เช่น BOTNOI มี `cmd-api` กับ `cmd-cron` ที่ใช้โค้ดใน `internal/` ชุดเดียวกัน

### ② `models/`: รูปแบบข้อมูล

| ไฟล์ | สิ่งที่อยู่ในไฟล์ | ใช้ที่ไหน |
|---|---|---|
| `product.go` | `Product` (มี tag json/bson) และ `ErrNotFound` | ทุก layer |
| `request.go` | `ProductRequest` (มี tag validate) | ข้อมูลที่ client ส่งเข้ามา |
| `response.go` | `GenericResponse{code, message}` | ตอบ error หรือข้อความทั่วไป |

แยก `ProductRequest` ออกจาก `Product` เพราะ client **ไม่ควรกำหนด `id` เองได้** id ต้องสร้างโดยระบบเท่านั้น

### ③ `storage/`: query อย่างเดียว

- `ProductStorage` เก็บ `*mongo.Collection`
- มี method `Insert`, `FindAll`, `FindByID`, `Update`, `Delete`
- แปลง `mongo.ErrNoDocuments` เป็น `models.ErrNotFound` เพื่อไม่ให้ error เฉพาะของ Mongo หลุดไปถึง layer อื่น
- **ไม่มี business rule** อยู่ใน layer นี้

### ④ `internal/api/routers/`: HTTP layer

| ไฟล์ | หน้าที่ |
|---|---|
| `factory.go` | `APIRouter` และ `NewAPIRouter()` ลงทะเบียน route ทั้งหมด (เปิดไฟล์นี้ไฟล์เดียวก็เห็นทุก endpoint) |
| `product_router.go` | handler ของแต่ละ endpoint และ `handleError()` ที่แปลง error เป็น HTTP status |

รูปแบบของทุก handler: **bind → validate → เรียก service → แปลง error → ตอบ JSON**

ไม่มี business logic และไม่ import `storage`

### ⑤ `internal/api/services/`: business logic

| ไฟล์ | หน้าที่ |
|---|---|
| `factory.go` | `IProductService` (สิ่งที่ router เห็น), `ProductStore` (สิ่งที่ service ต้องการจาก storage) และ `NewProductService(store)` |
| `product_service.go` | logic ของแต่ละ method |
| `product_service_test.go` | unit test ที่ใช้ fake storage |

**ไม่ import Echo และไม่ import mongo** จึงไม่ผูกกับ framework หรือ DB และเขียน test ได้ง่าย

ที่ test ได้เพราะ constructor รับ storage ผ่าน interface `ProductStore` จึงส่ง fake เข้าไปแทนของจริงได้

```go
store := newFakeStore(coffee)     // เก็บข้อมูลใน memory แทน Mongo
store.err = errDBDown             // (ถ้าต้องการ) จำลองว่า DB ล่ม
svc := NewProductService(store)   // ส่ง fake เข้าไปผ่าน constructor ตัวเดียวกับที่ main.go ใช้
```

### ⑥ `internal/middleware/`: ตั้งค่า Echo

- `middleware.go` → `EchoMiddleware()` สร้าง Echo พร้อม middleware และตอบ 404/405 ในรูปแบบ `GenericResponse`
- `validator.go` → `CustomValidator` ทำให้เรียก `c.Validate()` ได้ และแปลงข้อความ error ให้อ่านง่าย เช่น `name is required field`

### ⑦ `internal/pkg/`: เครื่องมือกลางที่ไม่เกี่ยวกับ business

| ไฟล์ | หน้าที่ |
|---|---|
| `connector/mongodb.go` | เปิด connection, `Ping`, `SelectDB`, `Disconnect` พร้อมตั้ง timeout |
| `environment/environment.go` | **ที่เดียวที่อ่าน env** และเลือกค่าตาม `MODE` (dev / uat / production) |
| `environment/env_loader.go` | หา `.env` จาก root ของโปรเจกต์ (โฟลเดอร์ที่มี `go.mod`) |

ตัวอย่างการเลือกค่าตาม MODE

```
MODE=dev        → ใช้ MONGODB_URI
MODE=uat        → ใช้ MONGODB_URI_UAT
MODE=production → ใช้ MONGODB_URI_PROD
```

### ⑧ `.env` / `.env.example`

- `.env.example` → ตัวอย่างค่า เอาขึ้น git ได้
- `.env` → ค่าจริง **ไม่ควรเอาขึ้น git** เพราะอาจมีรหัสผ่าน DB

### ⑨ `docker-compose.yml`

ใช้เปิด MongoDB ในเครื่องด้วย `docker compose up -d`
และเก็บข้อมูลไว้ใน volume `mongo-data` ข้อมูลจึงไม่หายตอนปิด container

---

## `factory.go` ทำอะไร

แต่ละ package มี `factory.go` ไว้ทำ 3 อย่าง

1. **ประกาศ struct:** บอกว่าชิ้นนี้ต้องใช้ dependency อะไรบ้าง
2. **ประกาศ interface:** บอกว่าชิ้นนี้ทำอะไรได้ หรือต้องการอะไรจากชิ้นอื่น
3. **มี constructor `NewXxx()`:** สร้าง struct ที่มี dependency ครบ พร้อมใช้งาน

```
main.go
  │
  ├─ connector.NewMongoDBClient()        → ได้ db
  │
  ├─ storage.NewProductStorage(db)       → ได้ storage
  │
  ├─ services.NewProductService(storage) → [services/factory.go]  รับ storage เข้ามา
  │
  └─ routers.NewAPIRouter(e, svc)        → [routers/factory.go]   รับ service เข้ามา
         └─ ผูก /products กับ handler ทั้งหมด
```

ทุก constructor **รับ dependency ที่สร้างไว้แล้ว** ไม่สร้างเอง การสร้างทั้งหมดเกิดใน `main.go` ที่เดียว

| ไฟล์ | สิ่งที่อยู่ในไฟล์ |
|---|---|
| `factory.go` | **โครงสร้าง:** struct, interface, constructor, การลงทะเบียน route |
| `product_service.go` / `product_router.go` | **การทำงาน:** logic ของแต่ละ method |

> ⚠️ ชื่อ "factory" ในที่นี้ไม่ใช่ Factory Pattern แบบตำรา (ที่เลือกว่าจะสร้าง type ไหนตามเงื่อนไข)
> แต่เป็นที่รวม Constructor + Dependency Injection ของ package

## ทำไม `internal/` แต่ `models/` และ `storage/` อยู่ข้างนอก

- **`internal/`** เป็นกฎของ Go: package ข้างในโฟลเดอร์นี้ **import จากโปรเจกต์อื่นไม่ได้** จึงเหมาะกับโค้ดที่ไม่ต้องการให้ใครเอาไปใช้

---

## เวลาเพิ่มฟีเจอร์ใหม่ (เช่น Order) ต้องแตะไฟล์ไหน

| ลำดับ | ไฟล์ | ทำอะไร |
|---|---|---|
| 1 | `models/order.go`, `models/request.go` | เพิ่ม `Order` และ `OrderRequest` |
| 2 | `storage/order_storage.go` | เพิ่ม `OrderStorage` และ `NewOrderStorage(db)` |
| 3 | `internal/api/services/factory.go` | เพิ่ม `IOrderService`, `OrderStore` และ `NewOrderService(store)` |
| 4 | `internal/api/services/order_service.go` (+ `_test.go`) | เขียน logic และ test |
| 5 | `internal/api/routers/factory.go` | เพิ่ม route `/orders` |
| 6 | `internal/api/routers/order_router.go` | เขียน handler |
| 7 | `cmd/cmd-api/main.go` | สร้าง `NewOrderStorage(db)` → `NewOrderService(orderStorage)` แล้วส่งเข้า router |

ไฟล์ของ Product ไม่ต้องแก้เลย ซึ่งเป็นประโยชน์หลักของการแบ่ง layer

---

## ปัญหาที่พบบ่อย

| Error | สาเหตุ | วิธีแก้ |
|---|---|---|
| `MongoDB URI is empty` | ไม่มีไฟล์ `.env` หรือไม่ได้ตั้ง `MONGODB_URI` | `cp .env.example .env` |
| `connection refused` ที่ `localhost:27017` | MongoDB ไม่ได้เปิดอยู่ | `docker compose up -d` |
| editor ขึ้น `could not import go-crud-demo/...` | editor เปิด workspace ของโฟลเดอร์อื่นอยู่ | เปิดโฟลเดอร์ `go-crud-demo` เป็น workspace |
