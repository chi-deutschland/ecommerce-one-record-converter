package main

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"chi-deutschland.com/ecommerce-one-record-converter/pkg/iata"
	"github.com/xuri/excelize/v2"
)

// rowData represents the expected Excel row data structure, with struct tags
// indicating the corresponding column header IDs and their indices in the Excel
// file. The struct tags are used for mapping the Excel columns to the struct
// fields during parsing and validation. Incorrect struct tags may lead to
// parsing errors or incorrect data mapping, so they must be defined accurately
// according to the expected Excel format. In particular, column indexes must be
// valid non-negative integers or the validation will panic.
type rowData struct {
	MasterAirWaybillNumber string `columnHeaderID:"MawbNr"             columnIndex:"0"`
	BoxID                  string `columnHeaderID:"BoxID"              columnIndex:"1"`
	ParcelID               string `columnHeaderID:"ParcelID"           columnIndex:"2"`
	ReferenceID            string `columnHeaderID:"ReferenceID"        columnIndex:"3"`
	ShipperName            string `columnHeaderID:"ShipperName"        columnIndex:"4"`
	ShipperStreet          string `columnHeaderID:"ShipperStreet"      columnIndex:"5"`
	ShipperZipcode         string `columnHeaderID:"ShipperZipcode"     columnIndex:"6"`
	ShipperCity            string `columnHeaderID:"ShipperCity"        columnIndex:"7"`
	ShipperCountryCode     string `columnHeaderID:"ShipperCountryCode" columnIndex:"8"`
	BuyerName              string `columnHeaderID:"BuyerName"          columnIndex:"9"`
	BuyerStreet            string `columnHeaderID:"BuyerStreet"        columnIndex:"10"`
	BuyerZipcode           string `columnHeaderID:"BuyerZipcode"       columnIndex:"11"`
	BuyerCity              string `columnHeaderID:"BuyerCity"          columnIndex:"12"`
	BuyerCountryCode       string `columnHeaderID:"BuyerCountryCode"   columnIndex:"13"`
	NumberOfPieces         string `columnHeaderID:"NumberOfPieces"     columnIndex:"14"`
	Quantity               string `columnHeaderID:"Quantity"           columnIndex:"15"`
	TotalWeight            string `columnHeaderID:"TotalWeight"        columnIndex:"16"`
	ItemHSCode             string `columnHeaderID:"ItemHSCode"         columnIndex:"17"`
	SKUNumber              string `columnHeaderID:"SKUNumber"          columnIndex:"18"`
	GoodsDescription       string `columnHeaderID:"GoodsDescription"   columnIndex:"19"`
	InvoiceDate            string `columnHeaderID:"InvoiceDate"        columnIndex:"20"`
	InvoiceNumber          string `columnHeaderID:"InvoiceNumber"      columnIndex:"21"`
	InvoiceCurrency        string `columnHeaderID:"InvoiceCurrency"    columnIndex:"22"`
	TotalValue             string `columnHeaderID:"TotalValue"         columnIndex:"23"`
	UnitPrice              string `columnHeaderID:"UnitPrice"          columnIndex:"24"`
	ProductWeight          string `columnHeaderID:"ProductWeight"      columnIndex:"25"`
	CountryOfOrigin        string `columnHeaderID:"CountryOfOrigin"    columnIndex:"26"`
}

const (
	columnHeaderNameKey = "columnHeaderID"
	columnIndexKey      = "columnIndex"
)

type HeaderValidationError struct {
	MissingColumns    []string
	UnexpectedColumns []string
}

func newHeaderValidationError(missingColumns, unexpectedColumns []string) *HeaderValidationError {
	return &HeaderValidationError{
		MissingColumns:    missingColumns,
		UnexpectedColumns: unexpectedColumns,
	}
}

func (e *HeaderValidationError) Error() string {
	var parts []string

	if len(e.MissingColumns) > 0 {
		parts = append(parts, fmt.Sprintf("missing columns: %v", e.MissingColumns))
	}

	if len(e.UnexpectedColumns) > 0 {
		parts = append(parts, fmt.Sprintf("unexpected columns: %v", e.UnexpectedColumns))
	}

	return "invalid headers: " + strings.Join(parts, ", ")
}

func validateHeaders(headers []string) error {
	var missingColumns []string

	var unexpectedColumns []string

	var expected rowData

	rowValue := reflect.ValueOf(&expected).Elem()
	fieldCount := rowValue.Type().NumField()

	for fieldIndex := range fieldCount {
		fieldValue := rowValue.Field(fieldIndex)
		fieldType := rowValue.Type().Field(fieldIndex)

		result := validateHeadersForField(fieldValue, fieldType, headers)

		if result.missingColumn != "" {
			missingColumns = append(missingColumns, result.missingColumn)
		}

		if result.unexpectedColumn != "" {
			unexpectedColumns = append(unexpectedColumns, result.unexpectedColumn)
		}
	}

	if len(missingColumns) > 0 || len(unexpectedColumns) > 0 {
		return newHeaderValidationError(missingColumns, unexpectedColumns)
	}

	return nil
}

type columnCheckResult struct {
	missingColumn    string
	unexpectedColumn string
}

func validateHeadersForField(
	fieldValue reflect.Value,
	fieldType reflect.StructField,
	headers []string,
) columnCheckResult {
	if fieldValue.Kind() != reflect.String || !fieldValue.CanSet() {
		return columnCheckResult{} // skip non-string and internal fields
	}

	columnHeaderName, ok := fieldType.Tag.Lookup(columnHeaderNameKey)
	if !ok {
		return columnCheckResult{} // skip fields without a column header name tag
	}

	rawColumnIndex, ok := fieldType.Tag.Lookup(columnIndexKey)
	if !ok {
		return columnCheckResult{} // skip fields without a column index tag
	}

	columnIndex, err := strconv.Atoi(rawColumnIndex)
	if err != nil || columnIndex < 0 {
		// This is a programming error, not a user input error, so a panic is
		// appropriate. Returning or logging would allow the error to go unnoticed and
		// lead to incorrect processing of the Excel file.
		panic("invalid column index in rowData struct tag for field " + fieldType.Name)
	}

	if columnIndex > len(headers) {
		return columnCheckResult{
			missingColumn: columnHeaderName,
		}
	}

	// normalize header and expected header name by disregarding capitalization and underscores for comparison
	normalizedHeader := strings.ReplaceAll(strings.ToLower(headers[columnIndex]), "_", "")
	normalizedExpectedHeader := strings.ReplaceAll(strings.ToLower(columnHeaderName), "_", "")

	if normalizedHeader != normalizedExpectedHeader {
		return columnCheckResult{
			unexpectedColumn: headers[columnIndex],
		}
	}

	return columnCheckResult{}
}

var errInputDataTooShort = errors.New("input data has fewer columns than expected")

func parseRowData(raw []string) (rowData, error) {
	var parsed rowData

	rowValue := reflect.ValueOf(&parsed).Elem()
	fieldCount := rowValue.Type().NumField()

	for fieldIndex := range fieldCount {
		fieldType := rowValue.Type().Field(fieldIndex)
		fieldValue := rowValue.Field(fieldIndex)

		err := setFieldValue(fieldValue, fieldType, raw)
		if err != nil {
			return rowData{}, fmt.Errorf("failed to set field value for field %s: %w", fieldType.Name, err)
		}
	}

	return parsed, nil
}

func setFieldValue(fieldValue reflect.Value, fieldType reflect.StructField, raw []string) error {
	rawColumnIndex, ok := fieldType.Tag.Lookup(columnIndexKey)
	if !ok {
		return nil // skip fields without a column number tag
	}

	if fieldValue.Kind() != reflect.String || !fieldValue.CanSet() {
		return nil // skip non-string and internal fields
	}

	columnIndex, err := strconv.Atoi(rawColumnIndex)
	if err != nil || columnIndex < 0 {
		// This is a programming error, not a user input error, so a panic is
		// appropriate. Returning or logging would allow the error to go unnoticed and
		// lead to incorrect processing of the Excel file.
		panic("invalid column index in rowData struct tag for field " + fieldType.Name)
	}

	if columnIndex >= len(raw) {
		return fmt.Errorf("%w: expected at least %d columns but got %d",
			errInputDataTooShort, columnIndex+1, len(raw))
	}

	fieldValue.SetString(raw[columnIndex])

	return nil
}

// MawbMap is a map of master air waybill numbers to their corresponding master
// air waybill data. It is used to group data by unique master air waybill
// numbers.
type MawbMap map[iata.MawbNumber]MasterAirWaybill

// MasterAirWaybill represents the data associated with a master air waybill,
// which is the top level entity in the data hierarchy.
type MasterAirWaybill struct {
	Boxes              BoxMap
	ShipperName        string
	ShipperStreet      string
	ShipperZipcode     string
	ShipperCity        string
	ShipperCountryCode string
}

func newEmptyMawb(data rowData) MasterAirWaybill {
	return MasterAirWaybill{
		Boxes:              make(map[BoxID]Box),
		ShipperName:        data.ShipperName,
		ShipperStreet:      data.ShipperStreet,
		ShipperZipcode:     data.ShipperZipcode,
		ShipperCity:        data.ShipperCity,
		ShipperCountryCode: data.ShipperCountryCode,
	}
}

// BoxID is a unique identifier for a box within a master air waybill.
type BoxID string

// BoxMap is a map of box IDs to their corresponding box data. It is used to
// group data by unique box IDs within a master air waybill.
type BoxMap map[BoxID]Box

// Box represents the data associated with a box, which is a container for
// parcels.
type Box struct {
	Parcels ParcelMap
}

// ParcelMap is a map of parcel IDs to their corresponding parcel data. It is
// used to group data by unique parcel IDs within a box.
type ParcelMap map[ParcelID]Parcel

// ParcelID is a unique identifier for a parcel within a box.
type ParcelID string

// Parcel represents the data associated with a parcel, which is a container for
// pieces.
type Parcel struct {
	Pieces PieceMap
}

// PieceMap is a map of reference IDs to their corresponding piece data.
type PieceMap map[ReferenceID]Piece

// ReferenceID is a unique identifier for a piece within a parcel.
type ReferenceID string

// Piece represents the data associated with a piece, which is a grouping for
// items.
type Piece struct {
	BuyerName        string
	BuyerStreet      string
	BuyerZipcode     string
	BuyerCity        string
	BuyerCountryCode string
	Items            []Item
}

// Item represents the data associated with an item, which is the physical good
// being shipped.
type Item struct {
	Product              Product
	UnitPrice            string
	InvoiceCurrency      string
	Quantity             string
	Weight               string
	InvoiceNumber        string
	CountryOfOrigin      string
	CountryOfDestination string
}

type Product struct {
	Description string
	HsCode      string
	SkuNumber   string
}

var errFailedToGetSheetList = errors.New("failed to get sheet list from excel file")

func ParseExcelData(file *excelize.File) (MawbMap, error) {
	sheetList := file.GetSheetList()
	if len(sheetList) == 0 {
		return nil, errFailedToGetSheetList
	}

	firstSheet := sheetList[0]

	rows, err := file.Rows(firstSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from excel file: %w", err)
	}

	mawbMap, err := excelRowsToOneRecord(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to convert excel to one record: %w", err)
	}

	return mawbMap, nil
}

var errNoRows = errors.New("no rows in excel")

func excelRowsToOneRecord(rows *excelize.Rows) (MawbMap, error) {
	mawbs := make(MawbMap)

	rowNumber := 1

	if !rows.Next() { // read header line
		return nil, errNoRows
	}

	headers, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to read row headers: %w", err)
	}

	err = validateHeaders(headers)
	if err != nil {
		return nil, fmt.Errorf("header validation failed: %w", err)
	}

	for rows.Next() {
		rowNumber++

		err := processRow(rows, mawbs)
		if err != nil {
			return nil, fmt.Errorf("failed to process row %d: %w", rowNumber, err)
		}
	}

	return mawbs, nil
}

func processRow(rows *excelize.Rows, mawbs MawbMap) error {
	data, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to read row column data: %w", err)
	}

	parsed, err := parseRowData(data)
	if err != nil {
		return fmt.Errorf("failed to parse row data: %w", err)
	}

	err = updateMawbMap(mawbs, parsed)
	if err != nil {
		return fmt.Errorf("failed to update mawb map: %w", err)
	}

	return nil
}

// ErrShipperMismatch indicates that data for a MAWB contains multiple shippers,
// which is not supported by the current implementation.
var ErrShipperMismatch = errors.New("not implemented for multiple shippers per MAWB")

func updateMawbMap(mawbs MawbMap, data rowData) error {
	mawbNumber, err := iata.ParseMawb(data.MasterAirWaybillNumber)
	if err != nil {
		return fmt.Errorf("failed to parse mawb number: %w", err)
	}

	_, exists := mawbs[mawbNumber]
	if !exists {
		mawbs[mawbNumber] = newEmptyMawb(data)
	}

	if !shipperMatchesExistingMawb(mawbs[mawbNumber], data) {
		return ErrShipperMismatch
	}

	rowBoxId := BoxID(data.BoxID)

	_, exists = mawbs[mawbNumber].Boxes[rowBoxId]
	if !exists {
		mawbs[mawbNumber].Boxes[rowBoxId] = newEmptyBox()
	}

	rowParcelID := ParcelID(data.ParcelID)
	rowReferenceID := ReferenceID(data.ReferenceID)

	if data.NumberOfPieces == "0" {
		return appendItemToExistingPiece(mawbs, mawbNumber, rowBoxId, rowParcelID, rowReferenceID, data)
	}

	_, exists = mawbs[mawbNumber].Boxes[rowBoxId].Parcels[rowParcelID]
	if !exists {
		mawbs[mawbNumber].Boxes[rowBoxId].Parcels[rowParcelID] = Parcel{
			Pieces: map[ReferenceID]Piece{
				rowReferenceID: newReferenceIDPiece(data),
			},
		}

		return nil
	}

	_, exists = mawbs[mawbNumber].Boxes[rowBoxId].Parcels[rowParcelID].Pieces[rowReferenceID]
	if exists {
		return fmt.Errorf("piece with reference ID %s already exists for parcel %s, box %s and MAWB %s",
			rowReferenceID, rowParcelID, rowBoxId, mawbNumber)
	}

	mawbs[mawbNumber].Boxes[rowBoxId].Parcels[rowParcelID].Pieces[rowReferenceID] = newReferenceIDPiece(data)

	return nil
}

func newReferenceIDPiece(data rowData) Piece {
	return Piece{
		BuyerName:        data.BuyerName,
		BuyerStreet:      data.BuyerStreet,
		BuyerZipcode:     data.BuyerZipcode,
		BuyerCity:        data.BuyerCity,
		BuyerCountryCode: data.BuyerCountryCode,
		Items:            []Item{newItem(data)},
	}
}

func newItem(data rowData) Item {
	return Item{
		Product: Product{
			Description: data.GoodsDescription,
			HsCode:      data.ItemHSCode,
			SkuNumber:   data.SKUNumber,
		},
		UnitPrice:            data.UnitPrice,
		InvoiceCurrency:      data.InvoiceCurrency,
		Quantity:             data.Quantity,
		Weight:               data.TotalWeight,
		InvoiceNumber:        data.InvoiceNumber,
		CountryOfOrigin:      data.CountryOfOrigin,
		CountryOfDestination: data.BuyerCountryCode,
	}
}

func newEmptyBox() Box {
	return Box{
		Parcels: make(map[ParcelID]Parcel),
	}
}

func shipperMatchesExistingMawb(waybill MasterAirWaybill, data rowData) bool {
	return waybill.ShipperName == data.ShipperName &&
		waybill.ShipperStreet == data.ShipperStreet &&
		waybill.ShipperZipcode == data.ShipperZipcode &&
		waybill.ShipperCity == data.ShipperCity &&
		waybill.ShipperCountryCode == data.ShipperCountryCode
}

var (
	ErrParcelNotFound           = errors.New("parcel not found")
	ErrReferenceIDPieceNotFound = errors.New("reference ID piece not found")
)

func appendItemToExistingPiece(
	mawbs MawbMap,
	rowMawb iata.MawbNumber,
	rowBoxID BoxID,
	rowParcelID ParcelID,
	rowReferenceID ReferenceID,
	data rowData,
) error {
	_, exists := mawbs[rowMawb].Boxes[rowBoxID].Parcels[rowParcelID]
	if !exists {
		return fmt.Errorf("%w: parcel with ID %s not found on box with ID %s in MAWB %s",
			ErrParcelNotFound, rowParcelID, rowBoxID, rowMawb)
	}

	piece, exists := mawbs[rowMawb].Boxes[rowBoxID].Parcels[rowParcelID].Pieces[rowReferenceID]
	if !exists {
		return fmt.Errorf("%w: piece with reference ID %s not found on parcel with ID %s, box with ID %s in MAWB %s",
			ErrReferenceIDPieceNotFound, rowReferenceID, rowParcelID, rowBoxID, rowMawb)
	}

	// Buyer should already be set from the first item with the same reference ID.
	// newItem(data)
	piece.Items = append(piece.Items, newItem(data))

	mawbs[rowMawb].Boxes[rowBoxID].Parcels[rowParcelID].Pieces[rowReferenceID] = piece

	return nil
}
