# eCommerce ONE Record Converter

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black?logo=next.js)](https://nextjs.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![IATA ONE Record](https://img.shields.io/badge/IATA-ONE%20Record-004B87)](https://www.iata.org/one-record)

A tool that converts eCommerce shipment data from Excel files into
[IATA ONE Record](https://www.iata.org/one-record) format and posts the
resulting Logistics Objects to a [NE:ONE Server](https://git.openlogisticsfoundation.org/wg-digitalaircargo/ne-one).

## Background

This project was developed by **[CHI Deutschland Cargo Handling GmbH](https://chi-cargo.com/)** as part of
the **[Digitales Testfeld Air Cargo (DTAC)](https://www.digitales-testfeld-air-cargo.de/)** research project. Within
DTAC, this converter belongs to the **Teilprojekt A: NE:ONE® Ecosystem** sub-project, which demonstrates how
the IATA ONE Record standard can create end-to-end transparency for eCommerce shipments across all entities in the air
cargo supply chain — from shippers and ground handling agents to carriers and customs.

In eCommerce air cargo, large shippers (e.g. TEMU, Shaoke, Shein) typically announce shipments via Excel spreadsheets.
Rather than requiring these shippers to change their established processes, this converter bridges the gap: it accepts
the
traditional Excel-based input and transforms it into the modern ONE Record linked-data format, enabling seamless
integration with ONE Record-enabled supply chain participants.

## Architecture

```
┌─────────────┐     ┌───────────────────────────────┐     ┌────────────────┐
│   Frontend  │────▶│          Go Backend           │────▶│  NE:ONE Server │
│  (Next.js)  │     │                               │     │                │
│             │     │  ┌────────┐   ┌─────────────┐ │     │  ONE Record    │
│  Upload UI  │     │  │ Excel  │──▶│  ONE Record │─┼────▶│  Logistics     │
│  + Config   │     │  │ Parser │   │  Converter  │ │     │  Objects       │
│             │     │  └────────┘   └─────────────┘ │     │                │
└─────────────┘     └───────────────────────────────┘     └────────────────┘
```

**Frontend** — A static Next.js app where users configure the NE:ONE Server connection (base URL + auth token) and
upload an Excel file.

**Backend** — A Go HTTP server that:

1. Receives the uploaded Excel file
2. Parses and validates the spreadsheet data
3. Converts each row into ONE Record Logistics Objects (JSON-LD)
4. Posts the objects to the configured NE:ONE Server via its REST API
5. Sends notifications for each created Box-level Piece

### Data Model

The Excel data is mapped to the following ONE Record object hierarchy:

```
Piece [Box]   (1 Shipment : N Box-Level Pieces)
└── Piece [Parcel]   (1 Box : N Parcel-Level Pieces)
    └── Piece [ReferenceID]   (1 Parcel : N ReferenceID-Level Pieces)
        └── Item   (1 Piece : N Items)
            └── Product   (1 Item : 1 Product)
```

Each level is created as a separate Logistics Object on the NE:ONE Server, with references linking them together
following the ONE Record data model.

## Prerequisites

| Dependency                                                                         | Version | Purpose                                         |
|------------------------------------------------------------------------------------|---------|-------------------------------------------------|
| [Go](https://go.dev/dl/)                                                           | 1.26+   | Backend server                                  |
| [Node.js](https://nodejs.org/)                                                     | 18+     | Building the frontend                           |
| [npm](https://www.npmjs.com/)                                                      | 9+      | Frontend dependency management                  |
| [NE:ONE Server](https://git.openlogisticsfoundation.org/wg-digitalaircargo/ne-one) | —       | ONE Record server to receive the converted data |

## Getting Started

### 1. Build the Frontend

The backend serves the frontend as static files. Build them first:

```bash
cd cmd/frontend
npm install
npm run build
```

This creates a `dist/` directory with the static export.

### 2. Configure the Backend

Edit `config.yaml` in the project root to adjust settings:

```yaml
http:
  addr: ":8181"                    # Address the HTTP server listens on
  staticFilesDir: "cmd/frontend/dist"  # Path to the built frontend files
neone:
  requestTimeout: 3m               # Timeout for individual NE:ONE API requests
  rateLimiterPolicy:
    maxExecutionsPerMinute: 100    # Max requests per minute to NE:ONE Server
    maxWaitTime: 30s               # Max time to wait when rate limit is exceeded
  retryPolicy:
    maxAttempts: 10                # Max retry attempts for failed requests
    delay: 1s                      # Initial retry delay (exponential backoff)
    maxDelay: 30s                  # Maximum retry delay
```

If `config.yaml` is not found or cannot be read, the application falls back to sensible defaults.

### 3. Run the Backend

```bash
go run cmd/converter
```

The server starts at `http://localhost:8181` (or the address configured in `config.yaml`).

### 4. Use the Application

1. Open `http://localhost:8181` in your browser
2. Enter your NE:ONE Server base URL and authentication token
3. Upload an Excel file (`.xlsx`) containing your eCommerce shipment data
4. The converter validates your token, parses the file, and begins posting Logistics Objects in the background

## Excel File Format

An example file with anonymized data is provided at
[`cmd/converter/testdata/anonymized_ecommerce_data.xlsx`](cmd/converter/testdata/anonymized_ecommerce_data.xlsx).

The spreadsheet must contain one item per row with the following columns (in order):

| #  | Column               | Description                                                                              |
|----|----------------------|------------------------------------------------------------------------------------------|
| 0  | `BoxID`              | Unique box identifier                                                                    |
| 1  | `ParcelID`           | Unique parcel identifier                                                                 |
| 2  | `ReferenceID`        | Unique reference identifier for a piece                                                  |
| 3  | `ShipperName`        | Name of the shipper                                                                      |
| 4  | `ShipperStreet`      | Shipper street address                                                                   |
| 5  | `ShipperZipcode`     | Shipper postal code                                                                      |
| 6  | `ShipperCity`        | Shipper city                                                                             |
| 7  | `ShipperCountryCode` | Shipper country code (ISO)                                                               |
| 8  | `BuyerName`          | Name of the buyer/consignee                                                              |
| 9  | `BuyerStreet`        | Buyer street address                                                                     |
| 10 | `BuyerZipcode`       | Buyer postal code                                                                        |
| 11 | `BuyerCity`          | Buyer city                                                                               |
| 12 | `BuyerCountryCode`   | Buyer country code (ISO)                                                                 |
| 13 | `NumberOfPieces`     | `1` for the first item of a piece, `0` for subsequent items sharing the same ReferenceID |
| 14 | `Quantity`           | Item quantity                                                                            |
| 15 | `TotalWeight`        | Total weight in kg                                                                       |
| 16 | `ItemHSCode`         | HS code for the item                                                                     |
| 17 | `SKUNumber`          | Product SKU number                                                                       |
| 18 | `GoodsDescription`   | Description of the goods                                                                 |
| 19 | `InvoiceDate`        | Invoice date                                                                             |
| 20 | `InvoiceNumber`      | Invoice number                                                                           |
| 21 | `InvoiceCurrency`    | Currency code (e.g. `EUR`, `USD`)                                                        |
| 22 | `TotalValue`         | Total value of the item                                                                  |
| 23 | `UnitPrice`          | Unit price                                                                               |
| 24 | `ProductWeight`      | Weight of the product                                                                    |
| 25 | `CountryOfOrigin`    | Country of origin code (ISO)                                                             |

> **Note:** The `NumberOfPieces` column controls piece grouping. Set it to `1` for the first item in a piece and `0`
> for all subsequent items that share the same `ReferenceID` within the same parcel.

## Notifications

The converter sends a creation notification to the NE:ONE Server for each Box-level Piece. To forward these
notifications to third-party systems, configure Notification Forwarding on your NE:ONE Server instance.

## Running Tests

```bash
go test ./...
```

## Building for Linux

A build script for cross-compiling to Linux (amd64) is provided:

```bash
cd build
bash gobuild.sh
```

## Project Structure

```
├── cmd/
│   ├── converter/          # Go backend application
│   │   ├── main.go         # HTTP server setup and entry point
│   │   ├── converter.go    # Excel-to-ONE-Record conversion and NE:ONE posting
│   │   ├── parser.go       # Excel file parsing and data model definitions
│   │   ├── parser_test.go  # Parser tests with embedded test data
│   │   ├── config/         # YAML configuration loading
│   │   └── testdata/       # Sample Excel file for testing
│   └── frontend/           # Next.js frontend application
│       └── src/
│           ├── app/        # Pages (upload wizard + done page)
│           ├── components/ # Reusable UI components
│           ├── context/    # React context (NE:ONE server config)
│           └── lib/        # Utility functions
├── pkg/
│   ├── iata/
│   │   └── onerecord/      # ONE Record data model (JSON-LD structs)
│   └── neone/
│       └── neone.go        # NE:ONE Server API client (with retry + rate limiting)
├── build/                  # Build scripts
├── config.yaml             # Application configuration
└── README.md
```

## Troubleshooting

| Problem                         | Possible Cause                       | Solution                                                                                  |
|---------------------------------|--------------------------------------|-------------------------------------------------------------------------------------------|
| Static files not loading        | Frontend not built                   | Run `npm run build` in `cmd/frontend`                                                     |
| `401 Unauthorized` after upload | Invalid or expired NE:ONE auth token | Check your token and NE:ONE Server configuration                                          |
| `connection refused`            | NE:ONE Server not reachable          | Verify the server URL and that the NE:ONE Server is running                               |
| Excel parsing errors            | Wrong file format or missing columns | Use the provided example file as a template; ensure all columns match the expected format |
| Rate limit warnings in logs     | Too many requests to NE:ONE Server   | Adjust `maxExecutionsPerMinute` in `config.yaml`                                          |

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Acknowledgments

- **[Digitales Testfeld Air Cargo (DTAC)](https://www.digitales-testfeld-air-cargo.de/)** — The research project
  enabling this work
- **[IATA ONE Record](https://www.iata.org/one-record)** — The linked-data standard for air cargo
- **[NE:ONE Server](https://git.openlogisticsfoundation.org/wg-digitalaircargo/ne-one)** — The open-source ONE Record
  server implementation
- **[Open Logistics Foundation](https://www.openlogisticsfoundation.org/)** — Hosting the collaborative development
  environment

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

Copyright © 2025 [CHI Deutschland Cargo Handling GmbH](https://chi-cargo.com/)
