# stravaDataExporter

## アプリケーション概要

- StravaのAPIから自身のアクティビティを取得しInfluxDBに登録　

## 前提条件

- Go言語で開発
- Go言語のVersionは最新
- DockerContainerで運用
- GitHubでバージョン管理
- GitHubActionsでビルド
- 自宅サーバのUbuntu24.04LTSで運用
- 開発はVSCode + devContainer
- 成果物
    - 全機能実装済みの.goファイル軍
    - `go mod init stravaDataExporter` 実行済みの go.modファイル
    - gihubActionsでのビルド向け実装済みファイル郡
    - docker compose 向け compose.yaml と .env
    - .devcontainer, devcontainer.json
    - デバッグ起動可能な.vscode/launch.json

## 実装機能

- StravaのAPIから自身のアクティビティを取得しInfluxDBに登録
- Webアプリケーションとして動作
- StravaのAccessTokenはStravaの認証へリダイレクトしAccessTokenを取得し保持
- 一度取得したAccessTokenはRefreshTokenによって定期的にリフレッシュを実施。1日1回定期
- StravaAPIからのデータ取得は1時間に1回実施
- StravaAPIから取得したデータにFTPを付与。FTPはCSVから読み取った値を使用
- StravaAPIから取得したデータと付与したFTPからTSS,NPを算出し付与
- 月曜日から日曜日を1週間とし月曜日の0時0分0秒にその1週間の合計TSS、運動時間、走行距離、獲得標高を算出し登録。この処理は、その週のデータを取得した際、毎回算出し上書き
- 毎月1日からその月の末日までを1ヶ月として1日の0時0分0秒にその1ヶ月の合計TSS、運動時間、走行距離を算出し登録。この処理は、その月のデータを取得した際、毎回算出し上書き
‐ 毎年1月1日からその年の大晦日までを1年としてその年の1月1日の0時0分0秒にその1年の合計TSS、運動時間、走行距離を算出し登録。この処理は、その年のデータを取得した際、毎回算出し上書き
