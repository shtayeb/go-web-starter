package config

type Database struct {
	DBUrl    string
	Database string
	Password string
	Username string
	Port     string
	Host     string
	Schema   string
}

type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

type SocialLogins struct {
	GoogleClientID     string
	GoogleClientSecret string
}

type Config struct {
	AppName      string
	AppEnv       string
	AppURL       string
	Debug        bool
	Port         int
	Database     Database
	Mailer       SMTP
	SocialLogins SocialLogins
}

func LoadConfigFromEnv() Config {
	return Config{
		AppName: GetEnv("APP_NAME", "Go Web Starter"),
		Port:    GetEnvAsInt("PORT", 8080),
		AppEnv:  GetEnv("APP_ENV", "local"),
		Debug:   GetEnvAsBool("DEBUG", true),
		AppURL:  GetEnv("APP_URL", "http://localhost:8080"),
		Database: Database{
			DBUrl:    GetEnv("BLUEPRINT_DB_URL", "./database.db"),
			Database: GetEnv("BLUEPRINT_DB_DATABASE", "blueprint"),
			Password: GetEnv("BLUEPRINT_DB_PASSWORD", "password1234"),
			Username: GetEnv("BLUEPRINT_DB_USERNAME", "shtb"),
			Port:     GetEnv("BLUEPRINT_DB_PORT", "5432"),
			Host:     GetEnv("BLUEPRINT_DB_HOST", "psql_bp"),
			Schema:   GetEnv("BLUEPRINT_DB_SCHEMA", "public"),
		},
		Mailer: SMTP{
			Host:     GetEnv("SMTP_HOST", "localhost"),
			Port:     GetEnvAsInt("SMTP_PORT", 587),
			Username: GetEnv("SMTP_USERNAME", "test"),
			Password: GetEnv("SMTP_PASSWORD", "test"),
			Sender:   GetEnv("SMTP_SENDER", "test@example.com"),
		},
		SocialLogins: SocialLogins{
			GoogleClientID:     GetEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: GetEnv("GOOGLE_CLIENT_SECRET", ""),
		},
	}
}
