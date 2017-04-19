# Lying SNTP Server

## Description
Simple SNTP server that adds given delta to reponse.

## Usage
### Flags
* `-p` / `--port` — to specify server port (default is 123)
* `-d` / `--delay` — to specify delta in seconds (default is 0)

### Example
* `sudo go run main.go -p 124 -d 6`

