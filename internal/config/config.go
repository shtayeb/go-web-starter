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

type Config struct {
	AppName  string
	AppEnv   string
	Debug    bool
	Port     int
	Database Database
	Mailer   SMTP
}

func LoadConfigFromEnv() Config {
	appName := GetEnv("APP_NAME", "Go Web Starter")
	appEnv := GetEnv("APP_ENV", "local")
	debug := GetEnvAsBool("DEBUG", true)
	port := GetEnvAsInt("PORT", 8080)

	dbUrl := GetEnv("BLUEPRINT_DB_URL", "./database.db")

	smtpHost := GetEnv("SMTP_HOST", "localhost")
	smtpPort := GetEnvAsInt("SMTP_PORT", 587)
	smtpUsername := GetEnv("SMTP_USERNAME", "test")
	smtpPassword := GetEnv("SMTP_PASSWORD", "test")
	smtpSender := GetEnv("SMTP_SENDER", "test@example.com")

	database := GetEnv("BLUEPRINT_DB_DATABASE", "blueprint")
	password := GetEnv("BLUEPRINT_DB_PASSWORD", "password1234")
	username := GetEnv("BLUEPRINT_DB_USERNAME", "shtb")
	dbPort := GetEnv("BLUEPRINT_DB_PORT", "5432")
	host := GetEnv("BLUEPRINT_DB_HOST", "psql_bp")
	schema := GetEnv("BLUEPRINT_DB_SCHEMA", "public")

	return Config{
		AppName: appName,
		Port:    port,
		AppEnv:  appEnv,
		Debug:   debug,
		Database: Database{
			DBUrl:    dbUrl,
			Database: database,
			Password: password,
			Username: username,
			Port:     dbPort,
			Host:     host,
			Schema:   schema,
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
