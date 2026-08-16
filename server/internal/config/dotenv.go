package config

import "github.com/joho/godotenv"

// LoadDotEnv はカレントまたは親ディレクトリの .env を読み込む。
// ファイルが無くてもエラーにしない。既に設定済みの環境変数は上書きしない。
func LoadDotEnv() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Load("../../../../.env")
}
