package config

import (
    "fmt"
    "os"
    "strconv"
)

// Config enthält alle Konfigurationsparameter
type Config struct {
    // Serielle Schnittstelle
    SerialPort string
    BaudRate   int

    // Datenbank
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
}

// Load lädt die Konfiguration aus Umgebungsvariablen
func Load() (*Config, error) {
    cfg := &Config{
        // Standardwerte
        SerialPort: getEnv("SERIAL_PORT", "/dev/ttyUSB0"),
        BaudRate:   getEnvAsInt("BAUD_RATE", 9600),
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "3306"),
        DBUser:     getEnv("DB_USER", "tempuser"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", "temperatures"),
    }

    // Validierung
    if cfg.DBPassword == "" {
        return nil, fmt.Errorf("DB_PASSWORD muss gesetzt sein")
    }

    return cfg, nil
}

// DSN erstellt den MySQL Data Source Name
func (c *Config) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
        c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// getEnv liest eine Umgebungsvariable mit Fallback-Wert
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

// getEnvAsInt liest eine Umgebungsvariable als Integer
func getEnvAsInt(key string, defaultValue int) int {
    strValue := getEnv(key, "")
    if strValue == "" {
        return defaultValue
    }

    value, err := strconv.Atoi(strValue)
    if err != nil {
        return defaultValue
    }
    return value
}
