package database

import (
    "database/sql"
    "fmt"
    "time"

    _ "github.com/go-sql-driver/mysql"

    "temp-recorder/internal/config"
)

// DB ist ein Wrapper für die Datenbankverbindung
type DB struct {
    conn   *sql.DB
    config *config.Config
}

// TemperatureReading repräsentiert eine Temperaturmessung
type TemperatureReading struct {
    ID          int64
    Temperature float64
    RecordedAt  time.Time
}

// New erstellt eine neue Datenbankverbindung
func New(cfg *config.Config) (*DB, error) {
    // Verbindung zur Datenbank herstellen
    conn, err := sql.Open("mysql", cfg.DSN())
    if err != nil {
        return nil, fmt.Errorf("fehler beim Öffnen der Datenbankverbindung: %w", err)
    }

    // Verbindung testen
    if err := conn.Ping(); err != nil {
        conn.Close()
        return nil, fmt.Errorf("fehler beim Ping der Datenbank: %w", err)
    }

    // Connection Pool konfigurieren
    conn.SetMaxOpenConns(10)
    conn.SetMaxIdleConns(5)
    conn.SetConnMaxLifetime(time.Minute * 5)

    return &DB{
        conn:   conn,
        config: cfg,
    }, nil
}

// Close schließt die Datenbankverbindung
func (db *DB) Close() error {
    if db.conn != nil {
        return db.conn.Close()
    }
    return nil
}

// InitSchema erstellt die benötigten Tabellen
func (db *DB) InitSchema() error {
    schema := `
        CREATE TABLE IF NOT EXISTS temperature_readings (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            temperature DECIMAL(5,2) NOT NULL,
            recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            INDEX idx_recorded_at (recorded_at)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `

    _, err := db.conn.Exec(schema)
    if err != nil {
        return fmt.Errorf("fehler beim Erstellen des Schemas: %w", err)
    }

    return nil
}

// SaveTemperature speichert eine Temperaturmessung
func (db *DB) SaveTemperature(temperature float64) error {
    query := `INSERT INTO temperature_readings (temperature, recorded_at) VALUES (?, ?)`

    _, err := db.conn.Exec(query, temperature, time.Now())
    if err != nil {
        return fmt.Errorf("fehler beim Speichern der Temperatur: %w", err)
    }

    return nil
}

// GetLatestReadings gibt die letzten n Messungen zurück
func (db *DB) GetLatestReadings(limit int) ([]TemperatureReading, error) {
    query := `
        SELECT id, temperature, recorded_at 
        FROM temperature_readings 
        ORDER BY recorded_at DESC 
        LIMIT ?
    `

    rows, err := db.conn.Query(query, limit)
    if err != nil {
        return nil, fmt.Errorf("fehler bei der Abfrage: %w", err)
    }
    defer rows.Close()

    var readings []TemperatureReading
    for rows.Next() {
        var r TemperatureReading
        if err := rows.Scan(&r.ID, &r.Temperature, &r.RecordedAt); err != nil {
            return nil, fmt.Errorf("fehler beim Scannen der Zeile: %w", err)
        }
        readings = append(readings, r)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("fehler beim Iterieren: %w", err)
    }

    return readings, nil
}

// GetAverageTemperature berechnet die Durchschnittstemperatur für einen Zeitraum
func (db *DB) GetAverageTemperature(since time.Duration) (float64, error) {
    query := `
        SELECT COALESCE(AVG(temperature), 0) 
        FROM temperature_readings 
        WHERE recorded_at >= ?
    `

    var avg float64
    err := db.conn.QueryRow(query, time.Now().Add(-since)).Scan(&avg)
    if err != nil {
        return 0, fmt.Errorf("fehler bei der Durchschnittsberechnung: %w", err)
    }

    return avg, nil
}
