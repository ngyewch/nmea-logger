# nmea-logger

Simple NMEA logger.

## Logger

### Output format

Format: JSONL

| Name        | Type     | Description         |
|-------------|----------|---------------------|
| `timestamp` | `int64`  | Epoch milliseconds. |
| `nmea`      | `string` | NMEA string.        |

### Configuration

#### Environment variables

| Name          | Type     | Required | Default  | Description                                                                |
|---------------|----------|----------|----------|----------------------------------------------------------------------------|
| `LOG_LEVEL`   | `string` | No       | `info`   | Log level. One of: `error`, `warn`, `info`, `debug`, `trace`               |
| `OUTPUT_DIR`  | `string` | No       | `./logs` | Output directory.                                                          |
| `SERIAL_PORT` | `string` | Yes      |          | Serial port.                                                               |
| `BAUD_RATE`   | `int`    | Yes      |          | Baud rate.                                                                 |
| `DATA_BITS`   | `int`    | No       | `8`      | Data bits.                                                                 |
| `PARITY`      | `string` | No       | `N`      | Parity. One of: `N` (none), `O` (odd), `E` (even), `M` (mark), `S` (space) |
| `STOP_BITS`   | `string` | No       | `1`      | Stop bits. One of: `1`, `1.5`, `2`                                         |

### systemd

* Unit name: `nmea-logger.service`
* Configuration file: `/etc/nmea-logger.env`

## AIS viewer

```
nmea-logger ais view (input-file)
```
