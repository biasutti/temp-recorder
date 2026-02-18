package serial

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "strconv"
    "strings"
    "time"

    "temp-recorder/internal/config"

    "github.com/tarm/serial"
)

// Reader liest Temperaturdaten von der seriellen Schnittstelle
type Reader struct {
    port   *serial.Port
    config *config.Config
}

// NewReader erstellt einen neuen Serial Reader
func NewReader(cfg *config.Config) (*Reader, error) {
    c := &serial.Config{
        Name: cfg.SerialPort,
        Baud: cfg.BaudRate,
    }

    port, err := serial.OpenPort(c)
    if err != nil {
        return nil, fmt.Errorf("konnte serielle Schnittstelle nicht öffnen: %w", err)
    }

    return &Reader{
        port:   port,
        config: cfg,
    }, nil
}

// Close schließt die serielle Verbindung
func (r *Reader) Close() error {
    if r.port != nil {
        return r.port.Close()
    }
    return nil
}

// ReadTemperatures liest kontinuierlich Temperaturwerte und sendet den jeweils
// letzten gültigen Wert im konfigurierten Intervall an den Kanal (Entprellung)
func (r *Reader) ReadTemperatures(ctx context.Context, tempChan chan<- float64) error {
    scanner := bufio.NewScanner(r.port)

    // Kanal für geparste Temperaturwerte aus der Lese-Goroutine
    rawChan := make(chan float64, 100)

    // Fehlerkanal für Scanner-Fehler
    errChan := make(chan error, 1)

    // Goroutine zum kontinuierlichen Lesen der seriellen Daten
    go func() {
        defer close(rawChan)
        for scanner.Scan() {
            line := strings.TrimSpace(scanner.Text())

            // Leere Zeilen überspringen
            if line == "" {
                continue
            }

            // Temperaturwert parsen
            temp, err := parseTemperature(line)
            if err != nil {
                // Ungültige Daten überspringen, aber weiter lesen
                continue
            }

            // Plausibilitätsprüfung (-50°C bis +150°C)
            if temp < -50 || temp > 150 {
                continue
            }

            select {
            case rawChan <- temp:
            case <-ctx.Done():
                return
            }
        }
        if err := scanner.Err(); err != nil {
            errChan <- fmt.Errorf("fehler beim Lesen: %w", err)
        }
    }()

    // Entprellung: nur den letzten Wert pro Intervall weiterleiten
    ticker := time.NewTicker(time.Duration(r.config.ReadInterval) * time.Second)
    defer ticker.Stop()

    log.Printf("Leseintervall: %d Sekunden (Entprellung aktiv)", r.config.ReadInterval)

    var latestTemp *float64

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case err := <-errChan:
            return err
        case temp, ok := <-rawChan:
            if !ok {
                // Scanner hat aufgehört zu lesen
                return fmt.Errorf("serielle Verbindung geschlossen")
            }
            latestTemp = &temp
        case <-ticker.C:
            if latestTemp != nil {
                select {
                case tempChan <- *latestTemp:
                case <-ctx.Done():
                    return ctx.Err()
                }
                latestTemp = nil
            }
        }
    }
}

// parseTemperature parst einen Temperaturwert aus einem String
// Unterstützt Formate wie "23.5", "23,5", "23.5°C", "Temp: 23.5"
func parseTemperature(line string) (float64, error) {
    // Komma durch Punkt ersetzen (europäisches Format)
    line = strings.ReplaceAll(line, ",", ".")

    // Grad-Zeichen und Text entfernen
    line = strings.ReplaceAll(line, "°C", "")
    line = strings.ReplaceAll(line, "°", "")

    // Versuche Zahl am Ende oder als ganze Zeile zu finden
    parts := strings.Fields(line)
    if len(parts) == 0 {
        return 0, fmt.Errorf("leere Eingabe")
    }

    // Nimm den letzten Teil (oft ist das der Wert)
    valueStr := parts[len(parts)-1]

    // Wenn der letzte Teil keine Zahl ist, probiere alle
    temp, err := strconv.ParseFloat(valueStr, 64)
    if err != nil {
        // Versuche den ersten Teil
        temp, err = strconv.ParseFloat(parts[0], 64)
        if err != nil {
            return 0, fmt.Errorf("konnte Temperatur nicht parsen: %s", line)
        }
    }

    return temp, nil
}
