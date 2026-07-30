package main

import (
	"context"
	"fmt"
	"net/http"

	"chi-deutschland.com/ecommerce-one-record-converter/pkg/iata/onerecord"
	"chi-deutschland.com/ecommerce-one-record-converter/pkg/neone"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

// NeoneDataForwarder is an HTTP handler that reads Excel files from incoming
// requests, converts the data to NE:ONE format, and posts it to the NE:ONE
// Server.
type NeoneDataForwarder struct {
	Server *neone.Client
}

// NewNeoneDataForwarder creates a new instance of NeoneDataForwarder with the
// provided NE:ONE Server.
func NewNeoneDataForwarder(server *neone.Client) *NeoneDataForwarder {
	return &NeoneDataForwarder{
		Server: server,
	}
}

// ServeHTTP implements the http.Handler interface for NeoneDataForwarder. It
// reads an Excel file from the request body, converts it to NE:ONE format, and
// posts the data to the NE:ONE Server.
func (forwarder *NeoneDataForwarder) ServeHTTP(responseWriter http.ResponseWriter, req *http.Request) {
	const maxMemory = 20 << 20 // limit max memory usage to 32MB

	err := req.ParseMultipartForm(maxMemory)
	if err != nil {
		log.Err(err).Msg("failed to parse multipart form data")
		http.Error(responseWriter, "failed to parse multipart form data: "+err.Error(), http.StatusBadRequest)

		return
	}

	neoneServerURL := req.FormValue("neoneServerBaseAddress")

	file, _, err := req.FormFile("file")
	if err != nil {
		log.Err(err).Msg("failed to get file from form data")
		http.Error(responseWriter, "failed to get file from form data: "+err.Error(), http.StatusBadRequest)

		return
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Err(err).Msg("failed to close uploaded file")
		}
	}()

	excelFile, err := excelize.OpenReader(file)
	if err != nil {
		log.Err(err).Msg("failed to read excel file")
		http.Error(responseWriter, "failed to read excel file: "+err.Error(), http.StatusBadRequest)

		return
	}

	defer func() {
		err := excelFile.Close()
		if err != nil {
			log.Err(err).Msg("failed to close excel file")
		}
	}()

	boxMap, err := ParseExcelData(excelFile)
	if err != nil {
		log.Info().Err(err).Msg("failed to parse excel data")
		http.Error(responseWriter, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	authHeaderValue := req.Header.Get("Authorization")

	err = forwarder.Server.ValidateToken(req.Context(), neoneServerURL, authHeaderValue)
	if err != nil {
		log.Info().Err(err).Msg("failed to validate token with NE:ONE Server")
		http.Error(responseWriter, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	go func(ctx context.Context) { //nolint:gosec,contextcheck // using req ctx would cause work to be canceled too soon
		boxURLs, err := forwarder.postECommerceBoxes(ctx, neoneServerURL, authHeaderValue, boxMap)
		if err != nil {
			log.Err(err).Msg("Failed to post eCommerce Boxes to NE:ONE Server")

			return
		}

		log.Debug().Strs("URLs", boxURLs).Msg("Posted Boxes to NE:ONE Server")
	}(context.Background())

	log.Debug().Msg("Started background process to post data to NE:ONE Server")
	responseWriter.WriteHeader(http.StatusAccepted)
}

func (forwarder *NeoneDataForwarder) postECommerceBoxes(
	ctx context.Context,
	neoneServerURL string,
	auth string,
	boxes BoxMap,
) ([]string, error) {
	boxURLs := make([]string, 0, len(boxes))

	for id, box := range boxes {
		loParcels := loParcelsFromParcelMap(box.Parcels)

		involvedParties := []onerecord.Party{
			onerecord.CargoShipper(
				box.ShipperName,
				new(onerecord.CargoLocation(
					new(onerecord.CargoAddress(
						[]string{
							box.ShipperStreet,
						},
						box.ShipperZipcode,
						box.ShipperCity,
						box.ShipperCountryCode,
					))))),
		}

		boxPiece := onerecord.CargoBoxPiece(string(id), loParcels, involvedParties)

		log.Debug().Str("id", string(id)).Msg("Posting box to NE:ONE Server")

		logisticsObjectURL, err := forwarder.Server.PostLogisticsObject(ctx, neoneServerURL, auth, boxPiece)
		if err != nil {
			return nil, fmt.Errorf("failed to post box to neone server: %w", err)
		}

		boxURLs = append(boxURLs, logisticsObjectURL)
		log.Debug().Str("id", string(id)).Str("URL", logisticsObjectURL).Msg("Posted box to NE:ONE Server")

		idsToAuthorize, err := forwarder.Server.GetLogisticsObjectEmbeddedIDs(
			ctx,
			neoneServerURL,
			auth,
			logisticsObjectURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get embedded IDs for box %s: %w", logisticsObjectURL, err)
		}

		log.Debug().
			Str("box_URL", logisticsObjectURL).
			Strs("embedded_IDs", idsToAuthorize).
			Msg("Retrieved embedded IDs for box from NE:ONE Server")

		accessDelegationULR, err := forwarder.Server.DelegateAccess(ctx,
			neoneServerURL,
			auth,
			"Automatic access delegation for eCommerce box and embedded Pieces, Items and Products",
			idsToAuthorize)
		if err != nil {
			return nil, fmt.Errorf("failed to delegate access for box and embedded objects: %w", err)
		}

		log.Debug().
			Str("access_delegation_URL", accessDelegationULR).
			Str("box_URL", logisticsObjectURL).
			Msg("Posted access delegation for box and embedded objects to NE:ONE Server")
	}

	return boxURLs, nil
}

func loParcelsFromParcelMap(parcels ParcelMap) []onerecord.Piece {
	loParcels := make([]onerecord.Piece, 0, len(parcels))

	for id, parcel := range parcels {
		parcelPiece := onerecord.CargoParcelPiece(string(id), loPiecesFromPieces(parcel.Pieces))

		loParcels = append(loParcels, parcelPiece)
	}

	return loParcels
}

func loPiecesFromPieces(pieces PieceMap) []onerecord.Piece {
	loPieces := make([]onerecord.Piece, 0, len(pieces))

	for id, piece := range pieces {
		consignee := onerecord.CargoConsignee(
			piece.BuyerName,
			new(onerecord.CargoLocation(
				new(onerecord.CargoAddress(
					[]string{
						piece.BuyerStreet,
					},
					piece.BuyerZipcode,
					piece.BuyerCity,
					piece.BuyerCountryCode,
				)))))

		logisticObjectPiece := onerecord.CargoPiece(
			string(id),
			loItemsFromItems(piece.Items),
			[]onerecord.Party{consignee},
		)

		loPieces = append(loPieces, logisticObjectPiece)
	}

	return loPieces
}

func loItemsFromItems(items []Item) []onerecord.Item {
	loItems := make([]onerecord.Item, 0, len(items))

	for _, item := range items {
		loItem := onerecord.CargoItem(onerecord.ItemParams{
			Product: onerecord.CargoProduct(
				item.Product.SkuNumber,
				item.Product.HsCode,
				item.Product.Description,
			),
			ItemQuantity:          item.Quantity,
			UnitPrice:             item.UnitPrice,
			Currency:              item.InvoiceCurrency,
			WeightInKg:            item.Weight,
			InvoiceNumber:         item.InvoiceNumber,
			ProductionCountryCode: item.CountryOfOrigin,
			TargetCountryCode:     item.CountryOfDestination,
		})

		loItems = append(loItems, loItem)
	}

	return loItems
}
