# nmea-logger

Simple NMEA logger.

## Output format

Format: JSONL

| Name        | Type     | Description         |
|-------------|----------|---------------------|
| `timestamp` | `int64`  | Epoch milliseconds. |
| `nmea`      | `string` | NMEA string.        |
