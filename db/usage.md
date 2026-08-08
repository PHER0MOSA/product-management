参考
[github.com/golang-migrate/migrate](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md)
[Dockerコンテナ上でgolang-migrateを使ってMongoDBのマイグレーション](https://zenn.dev/optimisuke/articles/f0db1b46567a3e)

すべて **リポジトリルート（`product-management/`）** で実行する前提。

## SQL ファイルの生成（`create`）

空の up/down SQL を `db/migrations/` に作る。DB 接続は不要。

```bash
mkdir -p db/migrations

docker run --rm \
  -u "$(id -u):$(id -g)" \
  -v ./db/migrations:/migrations \
  migrate/migrate \
  create -ext sql -dir /migrations -seq {{sql_name}}
```

| 指定 | 意味 |
|---|---|
| `--rm` | 実行後にコンテナを削除する |
| `-u "$(id -u):$(id -g)"` | ホストと同じ UID/GID で動かし、作ったファイルの所有者を自分にする（root 所有を防ぐ） |
| `-v ./db/migrations:/migrations` | ホストの `db/migrations` をコンテナの `/migrations` に載せる（相対パス可。基準はカレントディレクトリ） |
| `create ... -seq` | 連番付きの `.up.sql` / `.down.sql` を生成する |

`{{sql_name}}` の例: `create_products` → `000001_create_products.up.sql` など。

## マイグレーション実行（`up`）

SQL を Postgres に適用する。先に DB を起動しておく。

```bash
docker compose up -d db
```

```bash
docker run --rm \
  --network product-management_default \
  -v ./db/migrations:/migrations \
  migrate/migrate \
  -path=/migrations \
  -database "postgres://products:products@db:5432/products?sslmode=disable" \
  up
```

| 指定 | 意味 |
|---|---|
| `--network product-management_default` | compose と同じ Docker ネットワークに入る（下記「ネットワーク」参照） |
| `-path=/migrations` | コンテナ内のマイグレーションディレクトリ |
| `-database "..."` | DB 接続 URL（user / password / host / port / DB 名） |
| `@db:5432` | **サービス名** `db`（`docker-compose.yml` のキー）。コンテナ同士はこの名前で通信する |
| `up` | 未適用のマイグレーションを順に適用する |

戻すときは末尾を `down 1`（1つ戻す）などに変える。

### `product-management_default` とは

`docker compose up` すると、プロジェクト用ネットワークが **1つ** 自動作成される。名前はだいたい `{ディレクトリ名}_default`。

このリポジトリでは `product-management_default`。  
`products-db` / `products-client` など **複数コンテナが同じネットワークに接続**する（コンテナごとにネットワークは作られない）。

確認:

```bash
docker network ls
docker network inspect product-management_default
```

migrate コンテナもこのネットワークに入らないと、ホスト名 `db` に届かない。

### ホスト名の使い分け

| 接続元 | DB のホスト |
|---|---|
| DBeaver / ホスト上のツール | `localhost`（ポート `5432` がホストに公開されている） |
| compose と同じネットワーク上のコンテナ（migrate など） | `db`（サービス名） |

コンテナ内の `localhost` は「そのコンテナ自身」なので、migrate の `-database` に `localhost` を書くと DB に届かない。
