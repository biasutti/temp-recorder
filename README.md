# Temperature Recorder

Eine Go-Anwendung, die Temperaturdaten über eine serielle Schnittstelle empfängt und in einer MariaDB-Datenbank speichert.

## Features

- Liest Temperaturdaten von serieller Schnittstelle (USB-zu-Seriell Adapter)
- Speichert Messwerte mit Zeitstempel in MariaDB/MySQL
- Läuft als OCI-Container
- Automatische Schema-Migration
- Graceful Shutdown
- Plausibilitätsprüfung der Messwerte

## Voraussetzungen

- Docker & Docker Compose
- USB-zu-Seriell Adapter (z.B. für Arduino, ESP32, etc.)
- Sensor, der Temperaturwerte über Serial sendet

## Schnellstart

### 1. Repository klonen und konfigurieren

```bash
# .env Datei erstellen
cp .env.example .env

# Passwörter in .env anpassen!
```

### 2. Container starten

```bash
# Mit Docker Compose

# Logs anzeigen
```

### 3. Serielle Schnittstelle anpassen

Passe in der `.env` oder `docker-compose.yml` den seriellen Port an:

```yaml
  - "/dev/ttyUSB0:/dev/ttyUSB0"  # Linux USB-Serial
  - "/dev/ttyACM0:/dev/ttyACM0"  # Linux Arduino
```

## Konfiguration

Alle Einstellungen erfolgen über Umgebungsvariablen:

| Variable | Beschreibung | Standardwert |
|----------|--------------|--------------|
| `SERIAL_PORT` | Pfad zur seriellen Schnittstelle | `/dev/ttyUSB0` |
| `BAUD_RATE` | Baudrate der Verbindung | `9600` |
| `DB_HOST` | Datenbank-Host | `localhost` |
| `DB_PORT` | Datenbank-Port | `3306` |
| `DB_USER` | Datenbank-Benutzer | `tempuser` |
| `DB_PASSWORD` | Datenbank-Passwort | *(erforderlich)* |
| `DB_NAME` | Datenbankname | `temperatures` |

## Erwartetes Datenformat

Die Anwendung erwartet einfache Temperaturwerte über die serielle Schnittstelle, einen Wert pro Zeile:

```
23.5
24.1
23.8
```

Unterstützte Formate:
- `23.5` - Einfacher Wert
- `23,5` - Europäisches Format (Komma)
- `23.5°C` - Mit Einheit
- `Temp: 23.5` - Mit Präfix

## Datenbank-Schema

```sql
CREATE TABLE temperature_readings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    temperature DECIMAL(5,2) NOT NULL,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_recorded_at (recorded_at)
);
```

## Entwicklung

### Lokal bauen

```bash
# Dependencies installieren

# Anwendung bauen

# Ausführen
DB_PASSWORD=test ./temp-recorder
```

### Docker Image bauen

```bash
```

### Nur Container starten (externe DB)

```bash
  --name temp-recorder \
  --device=/dev/ttyUSB0:/dev/ttyUSB0 \
  -e SERIAL_PORT=/dev/ttyUSB0 \
  -e BAUD_RATE=9600 \
  -e DB_HOST=192.168.1.100 \
  -e DB_PORT=3306 \
  -e DB_USER=tempuser \
  -e DB_PASSWORD=geheim \
  -e DB_NAME=temperatures \
  temp-recorder:latest
```

## Adminer (Datenbank-UI)

Nach dem Start mit Docker Compose ist Adminer unter `http://localhost:8080` erreichbar:

- **System**: MySQL
- **Server**: mariadb
- **Benutzer**: tempuser
- **Passwort**: (aus .env)
- **Datenbank**: temperatures

## Beispiel-Arduino-Sketch

```cpp
void setup() {
  Serial.begin(9600);
}

void loop() {
  float temperature = readTemperatureSensor();
  Serial.println(temperature, 2);  // 2 Dezimalstellen
  delay(5000);  // Alle 5 Sekunden
}
```

## Troubleshooting

### Keine Berechtigung für serielle Schnittstelle

```bash
# Benutzer zur dialout-Gruppe hinzufügen
sudo usermod -aG dialout $USER
# Neu einloggen erforderlich!
```

### Container hat keinen Zugriff auf /dev/ttyUSB0

```bash
# Prüfen ob das Device existiert
ls -la /dev/ttyUSB*

# Eventuell muss der Container privilegiert laufen
```

### Datenbank-Verbindungsfehler

```bash
# Prüfen ob MariaDB läuft

# Logs der Datenbank prüfen
```

## Lizenz

MIT License
