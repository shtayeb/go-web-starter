package config

type Database struct {
	DBUrl string
}

type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

type Config struct {
	AppEnv   string
	Port     int
	Database Database
	Mailer   SMTP
}

func LoadConfigFromEnv() Config {
	appEnv := GetEnv("APP_ENV", "local")
	port := GetEnvAsInt("PORT", 8080)
	dbUrl := GetEnv("BLUEPRINT_DB_URL", "./database.db")

	smtpHost := GetEnv("SMTP_HOST", "localhost")
	smtpPort := GetEnvAsInt("SMTP_PORT", 587)
	smtpUsername := GetEnv("SMTP_USERNAME", "test")
	smtpPassword := GetEnv("SMTP_PASSWORD", "test")
	smtpSender := GetEnv("SMTP_SENDER", "test@example.com")

	return Config{
		Port:   port,
		AppEnv: appEnv,
		Database: Database{
			DBUrl: dbUrl,
		},
		Mailer: SMTP{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
			Sender:   smtpSender,
		},
	}
}
