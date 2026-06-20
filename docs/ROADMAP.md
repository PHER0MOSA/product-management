# 商品管理アプリ開発ロードマップ

## 背景と目標

Swagger 学習用の最小構成（Codegen 生成 Go API + JSON デモ UI）から、**CI 学習とオニオンアーキテクチャ実践**を主目的に再構築するプロジェクトです。

### 確定した方針

| 項目 | 方針 |
|---|---|
| API 実装 | Codegen 依存をやめ、手書きオニオンアーキテクチャ |
| API 契約 | `openapi.yaml` を維持、Swagger UI で閲覧（`api.md` は作らない） |
| フォルダ構成 | **機能別（垂直スライス）**（`internal/product/`）+ 各機能内で層分け + DB 接続のみ共通 |
| DB | PostgreSQL |
| クライアント | 当面はブラウザ → Go API 直（一覧・詳細画面） |
| マイグレーション | 本リポジトリ `db/migrations/`（golang-migrate） |
| ドキュメント | DB スキーマ docs（tbls → `docs/schema/`） |
| CI | `go test` + golangci-lint + `migrate up` + `tbls doc` + `docs/schema/` 差分チェック |
| BFF | Spring Boot + JUnit は**最終フェーズ** |

---

## 目標アーキテクチャ

### フェーズ 1〜5（当面）

```mermaid
flowchart LR
  Browser["Browser\n:3000"]
  GoAPI["Go API\n:8080"]
  DB["PostgreSQL\n:5432"]
  SwaggerUI["Swagger UI\n:8082"]

  Browser -->|"GET /products"| GoAPI
  Browser -->|"GET /products/id"| GoAPI
  GoAPI --> DB
  SwaggerUI -.->|"openapi.yaml"| GoAPI
```

### フェーズ 6（将来）

```mermaid
flowchart LR
  Browser["Browser"]
  Spring["Spring Boot BFF\n:8081"]
  GoAPI["Go API\n:8080"]
  DB["PostgreSQL"]

  Browser --> Spring
  Spring --> GoAPI
  GoAPI --> DB
```

---

## Go オニオンアーキテクチャ（機能別）

```
server/
  cmd/server/main.go              # DI・起動のみ
  internal/
    product/
      domain/                     # エンティティ・repository interface
      application/                # ユースケース
      handler/                    # HTTP I/O
      infrastructure/
        memory/                   # テスト・開発用
        postgres/                 # PostgreSQL 実装
    infrastructure/
      postgres/
        db.go                     # 共通 DB 接続
db/
  migrations/                     # SQL マイグレーション（スキーマの正）
  .tbls.yml                       # tbls 設定
docs/
  schema/                         # tbls 生成物
```

スキーマ（マイグレーション）の正は **`db/migrations/`**。Go API は `DATABASE_URL` でそのスキーマを参照する。

### 依存の向き

```mermaid
flowchart TB
  Main["cmd/server/main.go"]
  Handler["product/handler"]
  App["product/application"]
  Domain["product/domain"]
  ProdInfra["product/infrastructure/postgres"]
  DbInfra["infrastructure/postgres/db.go"]

  Main --> Handler
  Main --> ProdInfra
  Main --> DbInfra
  Handler --> App
  App --> Domain
  ProdInfra --> Domain
  ProdInfra --> DbInfra
```

- **domain**: 他レイヤーに依存しない。repository interface はここ
- **application**: domain の interface のみ使用
- **handler**: application を呼び、JSON / ステータスコードを担当
- **product/infrastructure**: `ProductRepository` の実装（product 専用 struct）
- **infrastructure/postgres/db.go**: 接続プール（複数 repository で共有）

### アーキテクチャの設計判断

#### 垂直スライス vs 水平スライス

| | 垂直（機能別） | 水平（層別） |
|---|---|---|
| 切り方 | `product/`, `order/` | `handler/`, `application/`, `domain/` |
| メリット | 変更が1機能に閉じる、機能追加・チーム分担に強い | 層のルールが見えやすい、横断変更（認証・ログ）に強い |
| デメリット | 横断変更が複数フォルダに及ぶ、共通化の判断が必要 | 機能が増えると1変更が各層に散らばる |

**本プロジェクトの選択**: 外側を **垂直（機能別）**、各機能の中を **水平（層別）** にする折衷形。機能追加を見据え、オニオンの層分けも各スライス内で守る。

```
internal/
  product/    ← 垂直
    domain/   ← 水平
    application/
    handler/
```

#### `postgres` フォルダの命名

- `postgres/` は PostgreSQL 向け adapter として **一般的**（`persistence/` や `database/` を使うチームもある）
- **接続プール**（共通）と **repository 実装**（機能専用）で役割を分ける:
  - `internal/infrastructure/postgres/db.go` — 共通接続
  - `internal/product/infrastructure/postgres/` — product 専用 struct + SQL
- 機能側はフォルダ名を `postgres` にせず `repository.go` / `postgres_repository.go` とする選択肢もあるが、本プロジェクトでは現状の構成を採用する

#### 機能追加時（例: `order`）

`order` が出たら `product` と **同じレイヤー構造** を `internal/order/` に追加する:

```
internal/
  product/
    domain/ application/ handler/ infrastructure/
  order/
    domain/ application/ handler/ infrastructure/
  infrastructure/postgres/db.go   # 接続は共通のまま
```

各機能に **専用 repository struct** を増やす（1つの巨大 repository にまとめない）。

#### 機能をまたぐ処理

原則: **ドメイン同士は直接依存しない**。またがるロジックは **application 層** でオーケストレーションする。

| 規模 | 置き場所 | 例 |
|---|---|---|
| 小さい | 片方の application | `order/application/create_order.go` が `ProductRepository` interface も受け取る |
| 大きい | 横断用パッケージ | `internal/checkout/application/place_order.go` |

```mermaid
flowchart LR
  CheckoutApp["checkout/application"]
  OrderDomain["order/domain"]
  ProductDomain["product/domain"]

  CheckoutApp -->|"interface のみ"| OrderDomain
  CheckoutApp -->|"interface のみ"| ProductDomain
```

- 複数 repository を同一トランザクションで扱う場合は、共通 `infrastructure/postgres/tx.go` 等で Tx を張り、application が調整する
- ドメインイベントによる非同期連携は将来の選択肢（現段階では不要）

#### 横断的な共通処理

機能フォルダの外に出すもの:

| パス | 内容 |
|---|---|
| `internal/infrastructure/postgres/` | DB 接続、トランザクション |
| `internal/middleware/`（将来） | 認証、ログ、CORS |
| `cmd/server/main.go` | DI 配線 |

handler のエラー形式統一など **全 API 横断** の変更は、共通 middleware または各 `handler/` の規約で揃える。

### ドメインモデル（Product）

現フェーズでは **`Product` エンティティのみ** を定義する。`order` 等は将来フェーズで `order/domain/` に別途追加する。

| 観点 | 現状 |
|---|---|
| API | `openapi.yaml` は `Product` の GET のみ |
| 画面 | 一覧・詳細（商品参照のみ） |
| ROADMAP フェーズ 1〜5 | 商品カタログ閲覧が中心 |

`CommonError` は **ドメインエンティティではない**（HTTP エラーレスポンス用 DTO）。`product/domain/` には置かず、`handler` 層で扱う。

#### Product エンティティ

| フィールド | 型 | ルール | 備考 |
|---|---|---|---|
| `ID` | `int64` | 1 以上 | OpenAPI `productID` |
| `Name` | `string` | 空でない、最大 20 文字 | OpenAPI `productName` |
| `Price` | `int64` | 0 以上 | 円単位の整数 |

```go
// internal/product/domain/product.go（実装時のイメージ）
type Product struct {
    ID    int64
    Name  string
    Price int64
}
```

#### ドメインエラー

| 名前 | 用途 |
|---|---|
| `ErrProductNotFound` | 存在しない ID 参照時 |

バリデーションエラー（名前が長すぎる等）は **将来 CRUD 追加時** に `domain` に足す。現フェーズは GET のみなので最小限。

#### ProductRepository interface

```go
type ProductRepository interface {
    List(ctx context.Context) ([]Product, error)
    GetByID(ctx context.Context, id int64) (Product, error)
}
```

現 API に合わせて **読み取りのみ**。`Create` / `Update` / `Delete` は CRUD フェーズまで interface に含めない。

#### レイヤーごとの型の扱い

- **domain の `Product`** と **API JSON** はフィールドが同じなので、現フェーズでは変換なしでよい
- 将来フィールドが分かれたら `handler` で DTO を導入する

#### 作らないもの（現フェーズ）

- `Order` エンティティ
- `ProductID` 等の Value Object 型（学習が進んだら検討可）
- `description`, `stock`, `createdAt` 等の拡張フィールド

### API エンドポイント

- `GET /products` → 200 + `[]Product`
- `GET /products/{id}` → 200 + `Product` / 404 + `CommonError`

---

## ドキュメント構成

```
product-management/
  db/
    migrations/       # SQL マイグレーション（スキーマの正）
    .tbls.yml         # tbls 設定
  docs/
    ROADMAP.md        # 本ファイル（方針・ロードマップ）
    schema/           # tbls 生成物（CI で検証）
  server/             # Go API
  openapi.yaml        # API 契約（Swagger UI で閲覧）
```

API 仕様は `openapi.yaml` + Swagger UI（http://localhost:8082）で確認します。`api.md` は作りません。

```mermaid
flowchart LR
  CI["GitHub Actions"]
  Migrations["db/migrations/"]
  Tbls["tbls doc"]
  DbDocs["docs/schema/"]
  GoAPI["Go API"]
  Postgres["PostgreSQL"]

  CI -->|"migrate up"| Migrations
  CI -->|"tbls doc"| DbDocs
  Migrations --> Postgres
  GoAPI --> Postgres
```

### DB マイグレーションとドキュメント

**マイグレーション実行**（[golang-migrate](https://github.com/golang-migrate/migrate) CLI）:

```bash
export DATABASE_URL="postgres://products:products@localhost:5432/products?sslmode=disable"
migrate -path db/migrations -database "$DATABASE_URL" up
```

**DB ドキュメント生成**（[tbls](https://github.com/k1LoW/tbls)）:

`db/.tbls.yml` の `docPath` は `docs/schema` とする。

```bash
tbls doc -c db/.tbls.yml
```

ローカル開発時は `docker compose up -d db` で PostgreSQL を起動し、`migrate up` を実行してから Go API を起動する。

**CI の学習ポイント**: マイグレーションを変更したのに `docs/schema/` を再生成していないと、本リポジトリ CI の db-docs ジョブが失敗する。

### CI ジョブ

| ジョブ | 内容 |
|---|---|
| go-test | `go test ./...` |
| go-lint | golangci-lint |
| db-docs | PostgreSQL 起動 → migrate up → tbls doc → `git diff docs/schema/` |

---

## 実装フェーズ

| フェーズ | 内容 | 状態 |
|---|---|---|
| 0 | ドキュメント（ROADMAP・設計） | 完了 |
| 1 | Go オニオンアーキテクチャ骨格 | 未着手 |
| 2 | テスト | 未着手 |
| 3 | CI（go test + lint + db-docs） | 未着手 |
| 4 | PostgreSQL（docker-compose db + `db/migrations` + Go 接続） | 未着手 |
| 5 | クライアント UI（一覧・詳細） | 未着手 |
| 6 | Spring Boot BFF + JUnit | 未着手（将来） |

### フェーズ 6（将来）の概要

- `web/` ディレクトリ新設
- Thymeleaf で一覧・詳細
- `RestClient` で Go API 呼び出し
- JUnit 5 + `@WebMvcTest` + `@MockBean`
- ブラウザは Spring のみ、8080 は内部向け

---

## 成功基準

- [ ] `go test ./...` がローカル・CI で通る
- [ ] golangci-lint が CI で通る
- [ ] PostgreSQL から商品一覧・詳細が取得できる
- [ ] `http://localhost:3000` で一覧 → 詳細遷移ができる
- [ ] 本リポジトリ CI で migrate + tbls doc が通る
- [ ] マイグレーション変更時に `docs/schema/` 差分で CI が失敗する
- [x] `docs/ROADMAP.md` に方針・構成が一通り読める
