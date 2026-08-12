# 開発ツール学習メモ

各ツールの概念・用語・よく使う操作をまとめた学習ノートです。
このリポジトリでの具体的な設定は、各節の「このプロジェクトでは」に短く書いています。

## 概要（このリポジトリとの対応）

| ツール | 役割（ざっくり） | このリポジトリでの接点 |
|---|---|---|
| PostgreSQL | データを永続化する RDB | `docker compose` の `db`（`:5432`） |
| Go | API サーバーの実装言語 | `server/`（compose 外でローカル起動） |
| golangci-lint | Go の静的解析（Lint）をまとめて実行 | `server/.golangci.yml` + CI の lint ジョブ |
| DBeaver | DB を GUI で触るクライアント | localhost の PostgreSQL に接続 |
| nginx | Web サーバー（静的配信など） | `web-client` が `client/` を `:3000` で配信 |

```mermaid
flowchart LR
  Browser["Browser"]
  Nginx["nginx\n:3000"]
  GoAPI["Go API\n:8080"]
  Postgres["PostgreSQL\n:5432"]
  DBeaver["DBeaver"]

  Browser --> Nginx
  Browser -->|"API"| GoAPI
  GoAPI --> Postgres
  DBeaver --> Postgres
```

---

## PostgreSQL

### なにか

PostgreSQL（略称 Postgres）は、関係データベース管理システム（RDBMS）です。
データを「表（テーブル）」に整理して保存し、SQL で問い合わせ・更新します。

### 覚えておく概念

| 用語 | 意味 |
|---|---|
| RDB / 関係データベース | 行と列の表でデータを管理する方式 |
| データベース（DB） | テーブルなどをまとめた入れ物（例: `products`） |
| スキーマ | DB 内の名前空間。テーブルの配置場所（Postgres では既定で `public`） |
| テーブル | 列定義を持つ表。1行が1レコード |
| 主キー（PK） | 行を一意に識別する列（または列の組） |
| SQL | データの定義・操作・問い合わせ用の言語 |
| マイグレーション | スキーマ変更をバージョン管理した SQL を順に適用する仕組み |
| 接続情報 | host / port / user / password / database の5点セットが基本 |

**接続の考え方**

- **host**: サーバーの場所（ローカルなら `localhost`）
- **port**: 待ち受けポート（Postgres 既定は `5432`）
- **user / password**: 認証
- **database**: どの DB に入るか

アプリからはよく次のような URL 形式で渡します。

```text
postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=disable
```

### よく使う操作

**コンテナ経由で起動・確認（Docker 利用時）**

```bash
docker compose up -d db
docker compose ps
```

**対話シェル `psql`（コンテナ内の例）**

```bash
docker compose exec db psql -U products -d products
```

`psql` 内でよく使うもの:

| 操作 | 例 |
|---|---|
| テーブル一覧 | `\dt` |
| テーブル定義 | `\d テーブル名` |
| 終了 | `\q` |
| 行を見る | `SELECT * FROM products LIMIT 10;` |
| 件数 | `SELECT COUNT(*) FROM products;` |

**SQL の入口（CRUD）**

```sql
-- Create
INSERT INTO products (name, price) VALUES ('りんご', 100);

-- Read
SELECT id, name, price FROM products WHERE id = 1;

-- Update
UPDATE products SET price = 120 WHERE id = 1;

-- Delete
DELETE FROM products WHERE id = 1;
```

### このプロジェクトでは

- イメージ: `postgres:16`（サービス名 `db`、コンテナ名 `products-db`）
- 接続: `products` / `products` @ `localhost:5432` / DB 名 `products`
- URL 例: `postgres://products:products@localhost:5432/products?sslmode=disable`
- スキーマの正は `db/migrations/`（golang-migrate）。詳細は [ROADMAP.md](ROADMAP.md)

---

## Go

### なにか

Go（Golang）は Google が開発したプログラミング言語です。
コンパイルして単一バイナリにしやすく、静的型付け・標準ライブラリが充実しているのが特徴です。Web API や CLI、インフラ周辺のツール実装によく使われます。

### 覚えておく概念

| 用語 | 意味 |
|---|---|
| モジュール | 依存関係の単位。ルートに `go.mod` がある |
| パッケージ | 同じディレクトリ内の Go ファイルのまとまり（`package` 宣言） |
| `internal/` | モジュール外から import できないディレクトリ規約（公開 API を限定できる） |
| 静的型付け | コンパイル時に型が決まる |
| コンパイル | ソースを実行ファイル（バイナリ）に変換すること |
| 標準ライブラリ | 言語同梱のパッケージ群（`net/http`, `database/sql` など） |

**オニオンアーキテクチャとの橋渡し（ざっくり）**

- 外側（HTTP・DB）と内側（ドメイン・ユースケース）を分ける
- 依存は「外側 → 内側」向きにする
- 本プロジェクトでは機能別（例: `internal/product/`）の中で層を分ける

### よく使う操作

`server/` などモジュールルートで実行する想定です。

```bash
# 依存関係の整理
go mod tidy

# そのまま実行（開発時）
go run ./cmd/api

# テスト
go test ./...

# 単体テストのみ（DB 不要）
go test ./internal/...

# インテグレーションテスト（テスト用 DB 必須。下記「インテグレーションテスト」参照）
go test ./test/integration/ -v

# バイナリを作る
go build -o bin/api ./cmd/api
```

| コマンド | 用途 |
|---|---|
| `go run` | ビルドせずにすぐ実行（開発向け） |
| `go test` | テスト実行 |
| `go build` | 実行ファイル生成 |
| `go mod tidy` | `go.mod` / `go.sum` を実態に合わせて整える |

### このプロジェクトでは

- API 実装は `server/` 配下
- Docker Compose では API を起動しない（コード変更のたびにイメージ再ビルドしたくないため）
- ローカルで `go run` 等により起動する想定（ポートは ROADMAP 上 `:8080`）
- 設計の詳細は [ROADMAP.md](ROADMAP.md)

**開発 DB とテスト DB の使い分け**

| 用途 | 接続先 | 環境変数 |
|---|---|---|
| API 開発・起動 | `localhost:5432` / `products` | `DATABASE_URL` |
| インテグレーションテスト | `localhost:5433` / `products_test` | `TEST_DATABASE_URL` |

開発中に `products` の中身が変わっても、テストは別 DB を使うため結果がぶれない。

**インテグレーションテスト**

毎回新しい DB にする（ボリュームなし。コンテナを作り直すと中身は空になる）:

```bash
# リポジトリルート
docker compose -f docker-compose.test.yml down
docker compose -f docker-compose.test.yml up -d

export TEST_DATABASE_URL="postgres://products:products@localhost:5433/products_test?sslmode=disable"
cd server
go test ./test/integration/ -v
```

`TestMain` がスキーマ作成（`CREATE TABLE IF NOT EXISTS`）とシード投入（商品A/B/C）を行う。**migrate は不要**。

同じコンテナのまま再実行する場合も、`TestMain` が毎回 TRUNCATE してシードを入れ直すため、テストデータは固定される。

---

## golangci-lint

### なにか

golangci-lint は、複数の Go 用リンター（静的解析ツール）を **1コマンドでまとめて実行**するラッパーです。
未使用変数・怪しいエラーハンドリング・スタイル違反などを、コンパイルやテストの前に見つけやすくします。

公式ドキュメント: [https://golangci-lint.run/](https://golangci-lint.run/)

### 覚えておく概念

| 用語 | 意味 |
|---|---|
| 静的解析（static analysis） | コードを実行せずにソースを読んで問題を検出すること |
| リンター（linter） | 規約違反やバグの芽を指摘するツール（`errcheck`, `staticcheck` など） |
| golangci-lint | 多数のリンターを並列実行し、結果をまとめて報告する統合ツール |
| `.golangci.yml` | 有効なリンター・除外パス・タイムアウトなどを書く設定ファイル |
| CI での lint | PR ごとに自動実行し、品質の下限をチームで揃える |

**`go test` との違い（ざっくり）**

| | golangci-lint | `go test` |
|---|---|---|
| 見るもの | 書き方・潜在バグのパターン | 実行した結果が期待どおりか |
| 実行 | コードを動かさない | テストコードを実行する |
| 役割 | 「変な書き方」を早く止める | 「振る舞い」が正しいか確認する |

両方あると、静的チェックと実行時検証でカバーが補完し合います。

### よく使う操作

**実行**

```bash
# 設定ファイルがあるディレクトリ（本リポジトリでは server/）で
golangci-lint run

# またはフルパスで
$(go env GOPATH)/bin/golangci-lint run
```

| コマンド | 用途 |
|---|---|
| `golangci-lint run` | カレント配下を解析（既定で `.golangci.yml` を読む） |
| `golangci-lint linters` | 使えるリンター一覧を見る |
| `golangci-lint version` | インストール済みバージョン確認 |

指摘が出たら、まずメッセージを読んで該当箇所を直す → 再実行、が基本サイクルです。

### このプロジェクトでは

- 設定: [`server/.golangci.yml`](../server/.golangci.yml)（`version: "2"`、`linters.default: standard`）
- ローカル実行: `server/` で `$(go env GOPATH)/bin/golangci-lint run`
- CI: [`.github/workflows/golangci-lint.yml`](../.github/workflows/golangci-lint.yml) が PR 時に `working-directory: server` で実行

---

## DBeaver

### なにか

DBeaver は、データベースを GUI で操作するクライアントです。
SQL を書いて実行したり、テーブルの中身や定義を一覧したりできます。CLI（`psql`）の代わりに「見て触る」学習に向きます。
[【DBクライアントツール論争】それでも私はDBeaverを使う](https://product.plex.co.jp/entry/dbeaver)

### 覚えておく概念

| 用語 | 意味 |
|---|---|
| DB クライアント | DB サーバーに接続して操作するアプリ（DBeaver, psql など） |
| ドライバ | DB 製品ごとの接続用ライブラリ（PostgreSQL 用ドライバなど） |
| コネクション | host / port / user / password / database を保存した接続設定 |
| 結果セット | `SELECT` の結果として返ってくる表形式のデータ |
| ER 図 | テーブル同士の関係を図で見る機能（学習・把握に便利） |

**GUI と CLI の使い分け（目安）**

- **DBeaver**: テーブル閲覧、結果の表確認、接続設定の保存、ER 図
- **psql / migrate**: スクリプト化、CI、マイグレーション適用

### よく使う操作

1. **新規接続**
   Database → New Database Connection → PostgreSQL を選ぶ
2. **接続パラメータを入れる**
   Host / Port / Database / Username / Password
3. **Test Connection**
   ドライバが未取得なら案内に従ってダウンロード
4. **SQL エディタ**
   接続を選んで SQL を書き、実行（▶）する
5. **テーブルを開く**
   左のツリーから schema → Tables → テーブルをダブルクリック

学習に効く機能の例:

- データタブで行を直接確認・編集（演習用。本番データでは慎重に）
- ER Diagram でテーブル関係を可視化
- 実行した SQL の履歴

### このプロジェクトでは

ローカルで `docker compose up -d db` したあと、次の設定で接続できます。

| 項目 | 値 |
|---|---|
| Host | `localhost` |
| Port | `5432` |
| Database | `products` |
| Username | `products` |
| Password | `products` |

マイグレーション適用後でないとテーブルが見えない点に注意（手順は [ROADMAP.md](ROADMAP.md)）。

---

## nginx

### なにか

nginx（エンジンエックス）は、高速な Web サーバー / リバースプロキシです。
静的ファイル（HTML / CSS / JS）の配信や、後ろのアプリへのリクエスト転送によく使われます。

### 覚えておく概念

| 用語 | 意味 |
|---|---|
| 静的ファイル配信 | すでに用意されたファイルをそのまま返す（本プロジェクトの用途） |
| リバースプロキシ | 受け取ったリクエストを裏側の別サーバー（API など）に転送する |
| `root` | 静的ファイルの置き場所（ドキュメントルート） |
| `location` | URL パスごとの扱いを定義するブロック |
| `proxy_pass` | リバースプロキシで転送先を指定するディレクティブ |
| alpine イメージ | 小さい Linux ベースの Docker イメージ（`nginx:alpine` など） |

**静的配信 vs リバースプロキシ**

```text
静的配信:  Browser → nginx → /usr/share/nginx/html のファイル
リバースプロキシ: Browser → nginx → 別プロセス（例: Go API :8080）
```

本プロジェクトの `web-client` は前者です。`proxy_pass` はまだ使いませんが、Web 構成の話で頻出するため概念だけ押さえておくとよいです。

### よく使う操作

**Compose でフロント用 nginx を起動**

```bash
docker compose up -d web-client
```

ブラウザで `http://localhost:3000` を開く。

**コンテナ内の既定ドキュメントルート**

- よくある配置: `/usr/share/nginx/html`
- ホストのディレクトリを volume マウントすると、コンテナを作り直さずに HTML を差し替えられる

設定ファイルの場所（イメージによるが典型例）:

- `/etc/nginx/nginx.conf`
- `/etc/nginx/conf.d/default.conf`

### このプロジェクトでは

- サービス: `web-client`（イメージ `nginx:alpine`、コンテナ名 `products-client`）
- ポート: ホスト `3000` → コンテナ `80`
- volume: `./client` → `/usr/share/nginx/html`
- ブラウザは nginx 経由で静的 UI を見つつ、API は Go（`:8080`）へ直接叩く構成（当面）
