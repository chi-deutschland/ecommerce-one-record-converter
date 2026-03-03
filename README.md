# eCommerce ONE Record Converter

This project converts eCommerce data from standard Excel files into the ONE Record format and posts it to a NE:ONE
Server.

## TechStack

### Frontend

* React
* NextJS
* Tailwind

### Backend

* Go

### Required Infrastructure

* [NE\-ONE Server](https://git.openlogisticsfoundation.org/wg-digitalaircargo/ne-one/-/tree/develop?ref_type=heads)

## Usage

### Frontend

Before running the server, you need to build the static files for the frontend. To do this, use the following commands
starting a terminal on the root directory of the project:

```bash
cd cmd/frontend
npm install
npm run build
```

### Backend

To run the backend server, use the following command on the root directory of the project:

```bash
go run cmd/backend
```

This will start the backend server on `http://localhost:8181`. You can change this and other configurations in the
`config.yaml` file in the root directory of the project.

### eCommerce Data

An example Excel file with anonymized eCommerce data is provided in the `cmd/converter/testdata` directory. Make sure to
use this format for your input data, as the converter is currently hardcoded to expect this specific format.

Note that the spreadsheet contains one item per line. The column NumberOfPieces should contain the value 1 for the first
item of a piece and 0 for all other items within the same piece, which share the same piece reference ID.

The eCommerce Data is grouped into the following levels and relationships in ONE-Record format:

- Waybill (MAWB) currently only one MAWB in test data, but multiple are supported
- Shipment (1 MAWB : 1 Shipment)
- Box-Level Piece (1 Shipment : N Box-Level-Pieces)
- Parcel-Level Piece (1 Box-Level-Piece : N Parcel-Level-Pieces)
- ReferenceID-Level Piece (1 Parcel-Level-Piece : N ReferenceID-Level-Pieces)
- Item (1 ReferenceID-Level-Piece : N Items)
- Product (1 Item : 1 Product)

### Notifications

The converter sends out notifications via the same NE:ONE Server specified for logistics object creation via the 
front-end for each Box-level piece created. Set up Notification Forwarding on the NE:ONE Server to receive 
notifications on a third party server.
