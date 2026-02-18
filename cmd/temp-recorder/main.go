package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "temp-recorder/internal/config"
    "temp-recorder/internal/database"
    "temp-recorder/internal/serial"
)

func main() {
    log.Println("Temperatur-Recorder startet...")

    // Konfiguration laden
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Fehler beim Laden der Konfiguration: %v", err)
    }

    log.Printf("Konfiguration geladen: Port=%s, Baudrate=%d, Intervall=%ds, DB=%s@%s/%s",
        cfg.SerialPort, cfg.BaudRate, cfg.ReadInterval, cfg.DBUser, cfg.DBHost, cfg.DBName)

    // Datenbankverbindung herstellen
    db, err := database.New(cfg)
    if err != nil {
        log.Fatalf("Fehler bei Datenbankverbindung: %v", err)
    }
    defer db.Close()

    log.Println("Datenbankverbindung hergestellt")

    // Datenbank-Schema initialisieren
    if err := db.InitSchema(); err != nil {
        log.Fatalf("Fehler bei Schema-Initialisierung: %v", err)
    }

    log.Println("Datenbank-Schema initialisiert")

    // Context für graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Serielle Schnittstelle öffnen und Daten lesen
    reader, err := serial.NewReader(cfg)
    if err != nil {
        log.Fatalf("Fehler beim Öffnen der seriellen Schnittstelle: %v", err)
    }
    defer reader.Close()

    log.Printf("Serielle Schnittstelle %s geöffnet", cfg.SerialPort)

    // Kanal für Temperaturdaten
    tempChan := make(chan float64, 100)

    // Goroutine zum Lesen der seriellen Daten
    go func() {
        if err := reader.ReadTemperatures(ctx, tempChan); err != nil {
            log.Printf("Fehler beim Lesen der Temperaturen: %v", err)
            cancel()
        }
    }()

    // Goroutine zum Speichern in die Datenbank
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case temp := <-tempChan:
                if err := db.SaveTemperature(temp); err != nil {
                    log.Printf("Fehler beim Speichern der Temperatur: %v", err)
                } else {
                    log.Printf("Temperatur gespeichert: %.2f°C", temp)
                }
            }
        }
    }()

    // Auf Shutdown-Signal warten
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    log.Println("Temperatur-Recorder läuft. CTRL+C zum Beenden...")

    <-sigChan
    log.Println("Shutdown-Signal empfangen, beende Anwendung...")
    cancel()

    log.Println("Temperatur-Recorder beendet")
}
