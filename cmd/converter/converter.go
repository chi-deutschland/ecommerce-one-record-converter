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
	var boxURLs []string

	for id, box := range boxes {
		parcelURLs, err := forwarder.postECommerceParcels(ctx, neoneServerURL, auth, box.Parcels)
		if err != nil {
			return nil, err // returned err should have enough context already
		}

		parcelPieces := onerecord.NewPieceReferenceArray(parcelURLs)

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

		boxPiece := onerecord.CargoBoxPiece(string(id), parcelPieces, involvedParties)

		log.Debug().Str("id", string(id)).Msg("Posting box to NE:ONE Server")

		logisticsObjectURL, err := forwarder.Server.PostLogisticsObject(ctx, neoneServerURL, auth, boxPiece)
		if err != nil {
			return nil, fmt.Errorf("failed to post box to neone server: %w", err)
		}

		boxURLs = append(boxURLs, logisticsObjectURL)
		log.Debug().Str("URL", logisticsObjectURL).Msg("Box posted to NE:ONE Server")

		err = forwarder.Server.PostLogisticsObjectCreationNotification(ctx, neoneServerURL, auth, logisticsObjectURL)
		if err != nil {
			log.Err(err).Msg("Failed to post notification for created box")

			continue // box was created successfully, so continue despite failure to notify.
		}

		log.Debug().Str("URL", logisticsObjectURL).Msg("Posted notification for created box")
	}

	accessDelegationULR, err := forwarder.Server.DelegateAccess(ctx,
		neoneServerURL,
		auth,
		"Automatic access delegation for eCommerce boxes",
		boxURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to delegate access for boxes: %w", err)
	}

	log.Debug().
		Str("access_delegation_URL", accessDelegationULR).
		Strs("URLs", boxURLs).
		Msg("Posted access delegation for boxes to NE:ONE Server")

	return boxURLs, nil
}

func (forwarder *NeoneDataForwarder) postECommerceParcels(
	ctx context.Context,
	neoneServerURL string,
	auth string,
	parcels ParcelMap,
) ([]string, error) {
	var parcelURLs []string

	for id, parcel := range parcels {
		log.Debug().Str("id", string(id)).Msg("Posting parcel to NE:ONE Server")

		pieceURLs, err := forwarder.postECommercePieces(ctx, neoneServerURL, auth, parcel.Pieces)
		if err != nil {
			return nil, err // returned err should have enough context already
		}

		pieceReferences := onerecord.NewPieceReferenceArray(pieceURLs)
		parcelPiece := onerecord.CargoParcelPiece(pieceReferences, string(id))

		logisticsObjectURL, err := forwarder.Server.PostLogisticsObject(ctx, neoneServerURL, auth, parcelPiece)
		if err != nil {
			return nil, fmt.Errorf("failed to post parcel to neone server: %w", err)
		}

		parcelURLs = append(parcelURLs, logisticsObjectURL)
		log.Debug().Str("URL", logisticsObjectURL).Msg("Parcel posted to NE:ONE Server")
	}

	accessDelegationULR, err := forwarder.Server.DelegateAccess(ctx,
		neoneServerURL,
		auth,
		"Automatic access delegation for eCommerce parcels",
		parcelURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to delegate access for parcels: %w", err)
	}

	log.Debug().
		Str("access_delegation_URL", accessDelegationULR).
		Strs("URLs", parcelURLs).
		Msg("Posted access delegation for parcels to NE:ONE Server")

	return parcelURLs, nil
}

func (forwarder *NeoneDataForwarder) postECommercePieces(
	ctx context.Context,
	neoneServerURL string,
	auth string,
	pieces PieceMap,
) ([]string, error) {
	var pieceURLs []string

	for id, piece := range pieces {
		log.Debug().Str("id", string(id)).Msg("Posting piece to NE:ONE Server")

		itemURLs, err := forwarder.postECommerceItems(ctx, neoneServerURL, auth, piece.Items)
		if err != nil {
			return nil, err // returned err should have enough context already
		}

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

		logisticObjectPiece := onerecord.CargoPiece(string(id), itemURLs, []onerecord.Party{consignee})

		logisticsObjectURL, err := forwarder.Server.PostLogisticsObject(ctx,
			neoneServerURL,
			auth,
			logisticObjectPiece)
		if err != nil {
			return nil, fmt.Errorf("failed to post piece to neone server: %w", err)
		}

		pieceURLs = append(pieceURLs, logisticsObjectURL)
		log.Debug().Str("URL", logisticsObjectURL).Msg("Piece posted to NE:ONE Server")
	}

	accessDelegationULR, err := forwarder.Server.DelegateAccess(ctx,
		neoneServerURL,
		auth,
		"Automatic access delegation for eCommerce pieces",
		pieceURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to delegate access for pieces: %w", err)
	}

	log.Debug().
		Str("access_delegation_URL", accessDelegationULR).
		Strs("URLs", pieceURLs).
		Msg("Posted access delegation for pieces to NE:ONE Server")

	return pieceURLs, nil
}

func (forwarder *NeoneDataForwarder) postECommerceItems(
	ctx context.Context,
	neoneServerURL string,
	auth string,
	items []Item,
) ([]string, error) {
	var productURLs []string
	var itemURLs []string

	for _, item := range items {
		log.Debug().
			Str("SKU", item.Product.SkuNumber).
			Str("HSCODE", item.Product.HsCode).
			Msg("Posting product to NE:ONE Server")

		productURL, err := forwarder.Server.PostLogisticsObject(ctx,
			neoneServerURL,
			auth,
			onerecord.CargoProduct(item.Product.SkuNumber, item.Product.HsCode, item.Product.Description),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to post item to neone server: %w", err)
		}

		productURLs = append(productURLs, productURL)
		log.Debug().Str("URL", productURL).Msg("Product posted to NE:ONE Server")

		logisticObjectItem := onerecord.CargoItem(onerecord.ItemParams{
			Product:               onerecord.NewProductReference(productURL),
			ItemQuantity:          item.Quantity,
			UnitPrice:             item.UnitPrice,
			Currency:              item.InvoiceCurrency,
			WeightInKg:            item.Weight,
			InvoiceNumber:         item.InvoiceNumber,
			ProductionCountryCode: item.CountryOfOrigin,
			TargetCountryCode:     item.CountryOfDestination,
		})

		log.Debug().Str("unit_price", item.UnitPrice).Msg("Posting item to NE:ONE Server")

		logisticsObjectURL, err := forwarder.Server.PostLogisticsObject(ctx,
			neoneServerURL,
			auth,
			logisticObjectItem)
		if err != nil {
			return nil, fmt.Errorf("failed to post item to neone server: %w", err)
		}

		itemURLs = append(itemURLs, logisticsObjectURL)
		log.Debug().Str("URL", logisticsObjectURL).Msg("Item posted to NE:ONE Server")
	}

	itemProductURLs := append(itemURLs, productURLs...)

	accessDelegationULR, err := forwarder.Server.DelegateAccess(ctx,
		neoneServerURL,
		auth,
		"Automatic access delegation for eCommerce items and products",
		itemProductURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to delegate access for items and products: %w", err)
	}

	log.Debug().
		Str("access_delegation_URL", accessDelegationULR).
		Strs("URLs", itemProductURLs).
		Msg("Posted access delegation for items and products to NE:ONE Server")

	return itemURLs, nil
}
