# Schemathesis 検証用 Go API

Schemathesis の検出能力を検証するために設計された Go 製 REST API です。  
スキーマ逸脱、境界値バリデーション、認証、ステートフル遷移、意図的バグの検出をテストできます。

## 前提条件

| ツール | バージョン | 用途 |
|---|---|---|
| [Go](https://go.dev/) | 1.21+ | API サーバーのビルド・実行 |
| [Task](https://taskfile.dev/) | 3.x | タスクランナー |
| [Python](https://www.python.org/) | 3.9+ | Schemathesis の実行 |

> [!TIP]
> Task のインストール: `brew install go-task`

## クイックスタート

```bash
# 1. 依存関係のセットアップ（初回のみ）
task setup

# 2. API サーバーを起動
task run

# 3. 別ターミナルで Schemathesis テストを実行
task test
```

## タスク一覧

### `task test` — 基本テスト

認証不要の全エンドポイントに対して、スキーマ適合性・境界値・バリデーションをテストします。

```
schemathesis run openapi.yaml --url http://localhost:8080
```

**対象エンドポイント例:**
- `GET /health` — 正常レスポンスの型チェック
- `GET /users?limit=&role=invalid` — クエリパラメータの境界値・列挙型チェック
- `POST /users` — リクエストボディの必須フィールド・型チェック
- `GET /bugs/*` — 意図的バグの検出（スキーマ違反・500エラー・未定義ステータスコード）

**検出できるもの:** スキーマ逸脱、境界値違反、不正ステータスコード、サーバーエラー

---

### `task test:auth` — 認証付きテスト

`Authorization: Bearer test-token` ヘッダーを付与してテストします。  
`task test` の内容に加え、認証が必要なエンドポイントにもアクセスできるようになります。

```
schemathesis run openapi.yaml --url http://localhost:8080 \
  -H "Authorization: Bearer test-token"
```

**`task test` との違い:**
- `GET /me` — 認証必須エンドポイントが 401 ではなく 200 を返すことを検証できる
- 認証済み状態での Users/Items/Orders 操作も検証対象に含まれる

**検出できるもの:** 上記に加え、認証フローのスキーマ適合性

---

### `task test:stateful` — ステートフルテスト

OpenAPI の `links` 定義を使い、複数エンドポイントを連鎖させてテストします。  
単発リクエストではなく「操作の流れ」全体を検証します。

```
schemathesis run openapi.yaml --url http://localhost:8080 --stateful=links
```

**実行されるシナリオ例:**
1. `POST /users` でユーザーを作成 → レスポンスの `id` を取得
2. `GET /users/{userId}` に作成した `id` を使ってアクセス
3. `PUT /users/{userId}` で更新 → `DELETE /users/{userId}` で削除

**`task test` との違い:**
- `task test` は各エンドポイントを**独立して**テストする（ランダム値を使用）
- `task test:stateful` は**前のレスポンスの値を次のリクエストに引き継ぐ**（現実的なシナリオを再現）

**検出できるもの:** 状態遷移の不整合、リソースの生成→参照→削除フローの破綻

---

### その他のタスク

| コマンド | 説明 |
|---|---|
| `task test:all` | 上記3つ（test → test:auth → test:stateful）を順番に実行 |
| `task lint` | `go vet` による静的解析 |
| `task fmt` | `gofmt` によるコード整形 |
| `task clean` | `.schemathesis`, `.hypothesis` キャッシュを削除 |

## プロジェクト構成

```
.
├── Taskfile.yml              # タスクランナー定義
├── openapi.yaml              # OpenAPI 3.0.3 仕様書
├── cmd/
│   └── api/
│       └── main.go           # エントリーポイント（chi router）
└── internal/
    ├── handler/
    │   ├── health_handler.go  # GET /health
    │   ├── users_handler.go   # Users CRUD + GET /me
    │   ├── items_handler.go   # Items CRUD
    │   ├── orders_handler.go  # Orders Create/Get
    │   └── bugs_handler.go    # 意図的バグ 4 種
    ├── middleware/
    │   ├── auth.go            # Bearer トークン認証
    │   └── recover.go         # panic → JSON 500 レスポンス
    ├── model/                 # リクエスト/レスポンス構造体
    ├── response/
    │   └── json.go            # JSON レスポンスヘルパー
    └── store/
        └── memory.go          # スレッドセーフなインメモリストア
```

## API エンドポイント

### 正常エンドポイント

| メソッド | パス | 説明 |
|---|---|---|
| `GET` | `/health` | ヘルスチェック |
| `GET` | `/users` | ユーザー一覧（`limit`, `role` でフィルタ可能） |
| `POST` | `/users` | ユーザー作成 |
| `GET` | `/users/{userId}` | ユーザー取得 |
| `PUT` | `/users/{userId}` | ユーザー更新 |
| `DELETE` | `/users/{userId}` | ユーザー削除 |
| `GET` | `/me` | 認証済みユーザー情報（Bearer 必須） |
| `GET` | `/items` | アイテム一覧（`category`, `minPrice`, `maxPrice` でフィルタ可能） |
| `POST` | `/items` | アイテム作成 |
| `GET` | `/items/{itemId}` | アイテム取得 |
| `POST` | `/orders` | 注文作成 |
| `GET` | `/orders/{orderId}` | 注文取得 |

### 意図的バグエンドポイント (`/bugs/*`)

Schemathesis がこれらの不具合を検出できることを確認するためのエンドポイントです。

| メソッド | パス | 検出対象 |
|---|---|---|
| `GET` | `/bugs/schema-mismatch` | 型の不一致・必須フィールドの欠落 |
| `GET` | `/bugs/status-mismatch` | 未定義のステータスコード（418 Teapot） |
| `POST` | `/bugs/panic-on-zero` | サーバーエラー（value=0 で panic → 500） |
| `GET` | `/bugs/invalid-email` | 不正な email フォーマット |

## 検証結果の見方

Schemathesis 実行後、以下のように結果が分類されます:

| 結果カテゴリ | 期待される検出元 |
|---|---|
| **Server error** | `/bugs/panic-on-zero` |
| **Response violates schema** | `/bugs/schema-mismatch`, `/bugs/invalid-email` |
| **Undocumented HTTP status code** | `/bugs/status-mismatch`, `/bugs/panic-on-zero` |

> [!IMPORTANT]
> 正常エンドポイントに対する失敗は **0 件** であるべきです。  
> すべての失敗が `/bugs/*` 配下に起因していれば、API は正しく実装されています。

## 検証項目

このAPIは以下の Schemathesis 機能を検証します:

- **スキーマ適合性**: レスポンスが OpenAPI 仕様に準拠しているか
- **境界値テスト**: `minimum`, `maximum`, `minLength`, `maxLength` の境界
- **列挙型バリデーション**: `enum` 制約の遵守
- **認証**: Bearer トークンによる 401/403 ハンドリング
- **ステートフルテスト**: API Links (`POST /users → GET /users/{userId}`) による遷移
- **不正入力の拒否**: `additionalProperties: false`, null 値, 未知のクエリパラメータ

## ライセンス

MIT