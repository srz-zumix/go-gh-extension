# instructions

指示が曖昧な場合は編集せずに曖昧な箇所を指摘してください。

ソースコード中のコメントは英語で記載


## コーディング規約

* fmt.Errorf: error strings should not end with punctuation or newlines (ST1005) go-staticcheck
* ローカルパッケージは github.com/srz-zumix/gh-team-kit/<path/to/dir> で import

### ソースコード全般

* ディレクトリ・ファイル構成は以下の責務分割に従うこと
  * action/: GitHub Actions 関連のビジネスロジック層
  * cmdflag/: コマンドライン引数の定義・パース
  * gh/: GitHub APIラッパー・ビジネスロジック層。API呼び出しはgh/client/配下で行い、gh/直下はラッパー・ユーティリティ関数のみ
  * gh/client/: go-github等の外部APIクライアント呼び出し、GraphQL 呼び出し専用。API と GraphQL のレスポンス型統一以外の理由でAPIレスポンスの整形は行わない。エラーはそのまま返す
  * ioutil/: 入出力関連のユーティリティ関数（ファイル操作、アーカイブ展開等）
  * parser/: 入力値のパース・バリデーション等
  * render/: 表示用の整形・出力処理（テーブル/JSON/hovercard等）
* gh/配下のラッパー関数は必ずctx context.Context, g *GitHubClientを先頭引数に取り、repository.Repository型等を利用する
* コメントは英語で記載し、関数・構造体・パッケージの責務が明確になるよう記述する
* テストコードは*_test.goで実装し、各責務ごとに配置する
* コード重複は避け、共通処理は関数化・ユーティリティ化する
* Lint/Formatter（go fmt, go vet, staticcheck等）を通してからコミットする
* 依存関係の循環(import cycle)が発生しないよう注意する

### パッケージ詳細

#### gh

* gh/client/*.go では API 呼び出しのエラーはフォーマットせずそのまま返します
* gh/member.go, gh/organizaion.go, gh/repo.go, gh/team.go, gh/user.go には github/client/*.go の関数のラッパーを記述します
  * owner/repo などの string は使わず repository.Repository 型を引数に取ります
  * ラッパー関数は ctx context.Context を第一、 g *GitHubClient を第二引数に取ります
