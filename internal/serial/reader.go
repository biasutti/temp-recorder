package serial

import (
    "bufio"
    "context"
    "fmt"
    "strconv"
    "strings"

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

// ReadTemperatures liest kontinuierlich Temperaturwerte und sendet sie an den Kanal
func (r *Reader) ReadTemperatures(ctx context.Context, tempChan chan<- float64) error {
    scanner := bufio.NewScanner(r.port)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if scanner.Scan() {
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

                // Temperatur an Kanal senden
                select {
                case tempChan <- temp:
                case <-ctx.Done():
                    return ctx.Err()
                }
            }

            if err := scanner.Err(); err != nil {
                return fmt.Errorf("fehler beim Lesen: %w", err)
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
