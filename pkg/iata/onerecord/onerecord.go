package onerecord

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"chi-deutschland.com/ecommerce-one-record-converter/pkg/iata"
)

// LogisticsObject is an interface for One Record objects which are Logistics
// Objects.
type LogisticsObject interface {
	// IsLogisticsObject is a marker method for One Record objects which are
	// Logistics Objects
	IsLogisticsObject()
}

// LogisticsAgent represents an agent involved in the logistics process.
type LogisticsAgent interface {
	// IsLogisticsAgent is a marker method to indicate that this is a LogisticsAgent.
	IsLogisticsAgent()
}

// Party represents a party involved in the logistics process.
type Party struct {
	PartyDetails LogisticsAgent   `json:"cargo:partyDetails,omitempty"`
	PartyRole    *CodeListElement `json:"cargo:partyRole,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
}

const (
	// HSType is the type for HS codes on Product.
	HSType = "UN Standard International Trade Classification"
)

// Unit codes for units of measurement, such as weight and piece units.
const (
	KilogramUnitCode = "KGM"
	PieceUnitCode    = "H87"

	// ConsigneeCode Consignee code from CodeListReferenceIATAParticipantIdentifier.
	ConsigneeCode = "CNE"

	// ShipperCode Shipper code from CodeListReferenceIATAParticipantIdentifier.
	ShipperCode = "SHP"
)

// CodeListVersion represents the version of a code list, such as a country code
// list or unit code list.
type CodeListVersion string

// CodeListVersion constants for different code lists used in One Record.
const (
	CodeListVersionHSCode    CodeListVersion = "2026"
	CodeListVersionPieceUnit CodeListVersion = "Revision 11e"
)

// CodeListReference represents a reference to a code list, such as a country code
// list or unit code list.
type CodeListReference string

// CodeListReference constants for different code lists used in One Record.
const (
	CodeListReferenceCountry    CodeListReference = "https://vocabulary.uncefact.org/CountryId"
	CodeListReferenceCurrency   CodeListReference = "https://vocabulary.uncefact.org/RevisedCurrencyCode"
	CodeListReferenceHSCode     CodeListReference = "www.tariffnumber.com"
	CodeListReferencePieceUnit  CodeListReference = "https://docs.peppol.eu/poacc/billing/3.0/codelist/UNECERec20/"
	CodeListReferenceWeightUnit CodeListReference = "https://vocabulary.uncefact.org/WeightUnitMeasureCode"

	CodeListReferenceIATACore                  = "https://onerecord.iata.org/ns/coreCodeLists"
	CodeListReferenceIATAParticipantIdentifier = "https://onerecord.iata.org/ns/code-lists/ParticipantIdentifier"
)

// CargoCodeListElement creates a new CodeListElement with the CargoContext.
func CargoCodeListElement(code string, reference CodeListReference, version CodeListVersion) CodeListElement {
	return CodeListElement{
		Context:           new(CargoContext()),
		Type:              "cargo:CodeListElement",
		Code:              code,
		CodeListVersion:   version,
		CodeListReference: reference,
	}
}

// CodeListElement represents a coded value from a code list, such as country or unit codes.
type CodeListElement struct {
	Code              string            `json:"cargo:code,omitempty"`
	CodeListReference CodeListReference `json:"cargo:codeListReference,omitempty"`
	CodeListVersion   CodeListVersion   `json:"cargo:codeListVersion,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
}

// Address represents a physical address, including country and street address lines.
type Address struct {
	CityName           string           `json:"cargo:cityName,omitempty"`
	Country            *CodeListElement `json:"cargo:country,omitempty"`
	StreetAddressLines []string         `json:"cargo:streetAddressLines,omitempty"`
	TextualPostCode    string           `json:"cargo:textualPostalCode,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
}

// CargoAddress creates a new Address with the CargoContext.
func CargoAddress(streetAddressLines []string, textualPostCode, cityName, countryCode string) Address {
	return Address{
		CityName:           cityName,
		Country:            new(CargoCodeListElement(countryCode, CodeListReferenceCountry, "")),
		StreetAddressLines: streetAddressLines,
		TextualPostCode:    textualPostCode,
		Type:               "cargo:Address",
		Context:            new(CargoContext()),
	}
}

// Location represents a location with an address.
type Location struct {
	Address *Address `json:"cargo:Address,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
}

// CargoLocation creates a new Location with the CargoContext.
func CargoLocation(address *Address) Location {
	return Location{
		Address: address,
		Type:    "cargo:Location",
		Context: new(CargoContext()),
	}
}

// Waybill represents a waybill document, such as a Master Air Waybill or a House Air Waybill.
type Waybill struct {
	ArrivalLocation   *Location   `json:"cargo:arrivalLocation,omitempty"`
	DepartureLocation *Location   `json:"cargo:departureLocation,omitempty"`
	WaybillNumber     string      `json:"cargo:waybillNumber,omitempty"`
	WaybillPrefix     string      `json:"cargo:waybillPrefix,omitempty"`
	WaybillType       WaybillType `json:"cargo:waybillType,omitempty"`
	InvolvedParties   []Party     `json:"cargo:involvedParties,omitempty"`
	ShippingRef       string      `json:"cargo:shippingRefNo,omitempty"`
	Shipment          *Shipment   `json:"cargo:shipment,omitempty"`

	// JSON-LD stuff
	Context *Context `json:"@context,omitempty"`
	Type    string   `json:"@type,omitempty"`
}

// CargoMasterWaybill creates a new Master Air Waybill with the given MAWB number
// and shipment.
func CargoMasterWaybill(mawbNumber iata.MawbNumber, shipment *Shipment) Waybill {
	return Waybill{
		WaybillNumber: strconv.Itoa(mawbNumber.Serial()),
		WaybillPrefix: strconv.Itoa(mawbNumber.AirlineCode()),
		WaybillType:   WaybillTypeMaster,
		Shipment:      shipment,
		Context:       new(CargoContext()),
		Type:          "cargo:Waybill",
	}
}

// IsLogisticsObject is a marker method to indicate that Waybill is a
// LogisticsObject.
func (Waybill) IsLogisticsObject() {}

// Context represents the JSON-LD context for One Record objects.
type Context struct {
	Cargo string `json:"cargo,omitempty"`
	API   string `json:"api,omitempty"`
}

// CargoContext returns a Context with the Cargo namespace for One Record
// objects.
func CargoContext() Context {
	return Context{
		Cargo: "https://onerecord.iata.org/ns/cargo#",
	}
}

// Shipment represents a Waybill shipment.
type Shipment struct {
	InvolvedParties []Party `json:"cargo:involvedParties,omitempty"`
	Pieces          []Piece `json:"cargo:pieces,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
}

// CargoShipment creates a new Shipment with the CargoContext.
func CargoShipment(pieces []Piece, parties []Party) Shipment {
	return Shipment{
		InvolvedParties: parties,
		Pieces:          pieces,
		Type:            "cargo:Shipment",
		Context:         new(CargoContext()),
	}
}

// Value represents a value with a numerical value and a unit.
type Value struct {
	NumericalValue string           `json:"cargo:numericalValue,omitempty"`
	Unit           *CodeListElement `json:"cargo:unit,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
}

// OtherIdentifier represents an additional identifier for an object, such as a reference ID.
type OtherIdentifier struct {
	OtherIdentifierType string `json:"cargo:otherIdentifierType,omitempty"`
	TextualValue        string `json:"cargo:textualValue,omitempty"`

	Context *Context `json:"@context,omitempty"`
	Type    string   `json:"@type,omitempty"`
}

// CargoOtherIdentifier creates a new OtherIdentifier with the CargoContext.
func CargoOtherIdentifier(otherIdentifierType, textualValue string) OtherIdentifier {
	return OtherIdentifier{
		Context:             new(CargoContext()),
		Type:                "cargo:OtherIdentifier",
		OtherIdentifierType: otherIdentifierType,
		TextualValue:        textualValue,
	}
}

// Product represents the type of item in a shipment.
type Product struct {
	Description      string            `json:"cargo:description,omitempty"`
	OtherIdentifiers []OtherIdentifier `json:"cargo:otherIdentifiers,omitempty"`
	HsCode           *CodeListElement  `json:"cargo:hsCode,omitempty"`
	HsType           string            `json:"cargo:hsType,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
}

// CargoProduct creates a new Product with the CargoContext.
func CargoProduct(skuNumber, hsCode, description string) Product {
	var otherIdentifiers []OtherIdentifier
	if skuNumber != "" {
		otherIdentifiers = []OtherIdentifier{CargoOtherIdentifier("SKU", skuNumber)}
	}

	return Product{
		Description:      description,
		OtherIdentifiers: otherIdentifiers,
		HsCode:           new(HSCode(hsCode)),
		HsType:           HSType,
		Context:          new(CargoContext()),
		Type:             "cargo:Product",
	}
}

// HSCode creates a new CodeListElement for an HS code.
func HSCode(hsCode string) CodeListElement {
	return CargoCodeListElement(hsCode, CodeListReferenceHSCode, CodeListVersionHSCode)
}

// Item represents a physical item in a shipment.
type Item struct {
	OfProduct         *Product          `json:"cargo:ofProduct,omitempty"`
	ItemQuantity      *Value            `json:"cargo:itemQuantity,omitempty"`
	UnitPrice         *Value            `json:"cargo:unitPrice,omitempty"`
	Weight            *Value            `json:"cargo:weight,omitempty"`
	OtherIdentifiers  []OtherIdentifier `json:"cargo:otherIdentifiers,omitempty"`
	ProductionCountry *CodeListElement  `json:"cargo:productionCountry,omitempty"`
	TargetCountry     *CodeListElement  `json:"cargo:targetCountry,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
	ID      string   `json:"@id,omitempty"`
}

// ItemParams represents the parameters needed to create a new CargoItem.
type ItemParams struct {
	Product               Product
	ItemQuantity          string
	UnitPrice             string
	Currency              string
	WeightInKg            string
	InvoiceNumber         string
	ProductionCountryCode string
	TargetCountryCode     string
}

// CargoItem creates a new Item with the CargoContext.
func CargoItem(params ItemParams) Item {
	var otherIdentifiers []OtherIdentifier
	if params.InvoiceNumber != "" {
		otherIdentifiers = []OtherIdentifier{CargoInvoiceNumber(params.InvoiceNumber)}
	}

	var productionCountry, targetCountry *CodeListElement
	if params.ProductionCountryCode != "" {
		productionCountry = new(CargoCodeListElement(params.ProductionCountryCode, CodeListReferenceCountry, ""))
	}

	if params.TargetCountryCode != "" {
		targetCountry = new(CargoCodeListElement(params.TargetCountryCode, CodeListReferenceCountry, ""))
	}

	price := new(CargoValue(params.UnitPrice, CargoCodeListElement(params.Currency, CodeListReferenceCurrency, "")))

	return Item{
		OfProduct:         new(params.Product),
		ItemQuantity:      new(CargoValue(params.ItemQuantity, PieceUnit())),
		UnitPrice:         price,
		Weight:            new(CargoValue(params.WeightInKg, KilogramWeightUnit())),
		OtherIdentifiers:  otherIdentifiers,
		ProductionCountry: productionCountry,
		TargetCountry:     targetCountry,
		Context:           new(CargoContext()),
		Type:              "cargo:Item",
	}
}

// CargoInvoiceNumber creates a new OtherIdentifier for an invoice number with
// the CargoContext.
func CargoInvoiceNumber(invoiceNumber string) OtherIdentifier {
	return CargoOtherIdentifier("invoice_number", invoiceNumber)
}

// KilogramWeightUnit returns a CodeListElement for the kilogram unit code with
// the CargoContext.
func KilogramWeightUnit() CodeListElement {
	return CargoCodeListElement(KilogramUnitCode, CodeListReferenceWeightUnit, "")
}

// IsLogisticsObject is a marker method to indicate that Item is a
// LogisticsObject.
func (Item) IsLogisticsObject() {}

// NewItemArrayFromIDs creates an array of Item references from an array of IDs.
func NewItemArrayFromIDs(ids []string) []Item {
	items := make([]Item, len(ids))
	for i, id := range ids {
		items[i] = Item{
			ID: id,
		}
	}

	return items
}

// CargoValue creates a new Value with the CargoContext.
func CargoValue(numericalValue string, unitCodeList CodeListElement) Value {
	return Value{
		Context:        new(CargoContext()),
		Type:           "cargo:Value",
		NumericalValue: numericalValue,
		Unit:           new(unitCodeList),
	}
}

// PieceUnit returns a CodeListElement for the piece unit code.
func PieceUnit() CodeListElement {
	return CargoCodeListElement(PieceUnitCode, CodeListReferencePieceUnit, CodeListVersionPieceUnit)
}

// Piece represents an individual piece or virtual grouping of pieces in a shipment.
type Piece struct {
	ContainedItems   []Item            `json:"cargo:containedItems,omitempty"`
	OtherIdentifiers []OtherIdentifier `json:"cargo:otherIdentifiers,omitempty"`
	ContainedPieces  []Piece           `json:"cargo:containedPieces,omitempty"`
	InvolvedParties  []Party           `json:"cargo:involvedParties,omitempty"`

	// JSON-LD stuff
	Context *Context `json:"@context,omitempty"`
	ID      string   `json:"@id,omitempty"`
	Type    string   `json:"@type,omitempty"`
}

// NewPieceReference creates a new Piece reference with the given ID.
func NewPieceReference(id string) Piece {
	return Piece{
		ID: id,
	}
}

// IsLogisticsObject is a marker method to indicate that Piece is a
// LogisticsObject.
func (Piece) IsLogisticsObject() {}

// NewPieceReferenceArray creates an array of Piece references from an array of
// IDs.
func NewPieceReferenceArray(ids []string) []Piece {
	pieces := make([]Piece, len(ids))
	for i, id := range ids {
		pieces[i] = NewPieceReference(id)
	}

	return pieces
}

// CargoPiece creates a new Piece with the CargoContext.
func CargoPiece(referenceID string, containedItemIDs []string, involvedParties []Party) Piece {
	var otherIdentifiers []OtherIdentifier
	if referenceID != "" {
		otherIdentifiers = []OtherIdentifier{ReferenceIDIdentifier(referenceID)}
	}

	containedItems := NewItemArrayFromIDs(containedItemIDs)

	return Piece{
		Context:          new(CargoContext()),
		Type:             "cargo:Piece",
		ContainedItems:   containedItems,
		InvolvedParties:  involvedParties,
		OtherIdentifiers: otherIdentifiers,
	}
}

// ReferenceIDIdentifier creates a new OtherIdentifier for a reference ID with
// the CargoContext.
func ReferenceIDIdentifier(referenceID string) OtherIdentifier {
	return CargoOtherIdentifier("reference_id", referenceID)
}

// CargoParcelPiece creates a new Piece representing a parcel, with the
// CargoContext.
func CargoParcelPiece(containedPieces []Piece, parcelID string) Piece {
	var otherIdentifiers []OtherIdentifier
	if parcelID != "" {
		otherIdentifiers = []OtherIdentifier{ParcelIDIdentifier(parcelID)}
	}

	return Piece{
		Context:          new(CargoContext()),
		Type:             "cargo:Piece",
		ContainedPieces:  containedPieces,
		OtherIdentifiers: otherIdentifiers,
	}
}

// ParcelIDIdentifier creates a new OtherIdentifier for a parcel ID with the
// CargoContext.
func ParcelIDIdentifier(parcelID string) OtherIdentifier {
	return CargoOtherIdentifier("Parcel ID", parcelID)
}

// CargoBoxPiece creates a new Piece representing a box, with the CargoContext.
func CargoBoxPiece(containedPieces []Piece, boxID string) Piece {
	var otherIdentifiers []OtherIdentifier
	if boxID != "" {
		otherIdentifiers = []OtherIdentifier{BoxIDIdentifier(boxID)}
	}

	return Piece{
		Context:          new(CargoContext()),
		Type:             "cargo:Piece",
		ContainedPieces:  containedPieces,
		OtherIdentifiers: otherIdentifiers,
	}
}

// BoxIDIdentifier creates a new OtherIdentifier for a box ID with the
// CargoContext.
func BoxIDIdentifier(boxID string) OtherIdentifier {
	return CargoOtherIdentifier("Box ID", boxID)
}

// WaybillType represents the type of waybill (MASTER, HOUSE, DIRECT).
type WaybillType string

// ErrUnknownWaybillType is an error returned when an unknown waybill type is
// encountered during JSON unmarshalling.
var ErrUnknownWaybillType = errors.New("unknown waybill type")

// UnmarshalJSON implements the json.Unmarshaler interface for WaybillType,
// allowing it to be unmarshalled from a JSON string.
func (t *WaybillType) UnmarshalJSON(data []byte) error {
	var waybill string

	err := json.Unmarshal(data, &waybill)
	if err != nil {
		return err //nolint:wrapcheck // no point in wrapping unmarshalling errors here
	}

	switch waybill {
	case "MASTER":
		*t = WaybillTypeMaster
	case "HOUSE":
		*t = WaybillTypeHouse
	case "DIRECT":
		*t = WaybillTypeDirect
	default:
		return fmt.Errorf("%w: %s", ErrUnknownWaybillType, waybill)
	}

	return nil
}

// WaybillType constants for different types of waybills.
const (
	WaybillTypeMaster WaybillType = "MASTER"
	WaybillTypeHouse  WaybillType = "HOUSE"
	WaybillTypeDirect WaybillType = "DIRECT"
)

// CargoConsignee creates a new Party representing a consignee, with the CargoContext.
func CargoConsignee(name string, location *Location) Party {
	return Party{
		PartyDetails: CargoCompany(name, location),
		PartyRole:    new(CargoConsigneeRole()),
		Type:         "cargo:Party",
		Context:      new(CargoContext()),
	}
}

// Company represents a LogisticsAgent which is a company involved in the
// logistics process, such as a shipper, consignee, or freight forwarder.
type Company struct {
	Name            string    `json:"cargo:name,omitempty"`
	BasedAtLocation *Location `json:"cargo:basedAtLocation,omitempty"`

	Type    string   `json:"@type,omitempty"`
	Context *Context `json:"@context,omitempty"`
}

// IsLogisticsAgent is a marker method to indicate that Company is a
// LogisticsAgent.
func (Company) IsLogisticsAgent() {}

// CargoCompany creates a new Company with the CargoContext.
func CargoCompany(name string, location *Location) Company {
	return Company{
		Name:            name,
		BasedAtLocation: location,
		Type:            "cargo:Company",
		Context:         new(CargoContext()),
	}
}

// CargoConsigneeRole creates a new CodeListElement for the consignee role, with
// the CargoContext.
func CargoConsigneeRole() CodeListElement {
	return CodeListElement{
		Code:              ConsigneeCode,
		CodeListReference: CodeListReferenceIATAParticipantIdentifier,
		CodeListVersion:   CodeListReferenceIATACore,
		Type:              "cargo:CodeListElement",
		Context:           new(CargoContext()),
	}
}

// CargoShipper creates a new Party representing a shipper, with the
// CargoContext.
func CargoShipper(name string, location *Location) Party {
	return Party{
		PartyDetails: CargoCompany(name, location),
		PartyRole:    new(CargoShipperRole()),
		Type:         "cargo:Party",
		Context:      new(CargoContext()),
	}
}

// CargoShipperRole creates a new CodeListElement for the Shipper role, with
// the CargoContext.
func CargoShipperRole() CodeListElement {
	return CodeListElement{
		Code:              ShipperCode,
		CodeListReference: CodeListReferenceIATAParticipantIdentifier,
		CodeListVersion:   CodeListReferenceIATACore,
		Type:              "cargo:CodeListElement",
		Context:           new(CargoContext()),
	}
}

// APIContext returns a Context with the API namespace for One Record objects.
func APIContext() Context {
	return Context{
		API: "https://onerecord.iata.org/ns/api#",
	}
}

// NotificationEventType represents the type of API event.
type NotificationEventType struct {
	ID string `json:"@id,omitempty"`
}

// APIAnyURI represents a URI value in the API namespace.
type APIAnyURI struct {
	Type  string `json:"@type"`
	Value string `json:"@value"`
}

// Notification represents a notification of an API event, such as the creation
// or update of a logistics object.
type Notification struct {
	APIHasEventType           *NotificationEventType `json:"api:hasEventType"`
	APIHasLogisticsObject     LogisticsObject        `json:"api:hasLogisticsObject"`
	APIHasLogisticsObjectType *APIAnyURI             `json:"api:hasLogisticsObjectType"`

	Context *Context `json:"@context"`
	Type    string   `json:"@type"`
	ID      string   `json:"@id"`
}

// APILogisticsObjectCreatedNotificationEventType returns a NotificationEventType
// for the "api:LOGISTICS_OBJECT_CREATED" event.
func APILogisticsObjectCreatedNotificationEventType() NotificationEventType {
	return NotificationEventType{
		ID: "https://onerecord.iata.org/ns/api#LOGISTICS_OBJECT_CREATED",
	}
}

// APIPieceLogisticsObjectType returns an APIAnyURI representing the type of
// logistics object for a CargoPiece.
func APIPieceLogisticsObjectType() APIAnyURI {
	return APIAnyURI{
		Type:  "http://www.w3.org/2001/XMLSchema#anyURI",
		Value: "https://onerecord.iata.org/ns/cargo#Piece",
	}
}

// PieceCreatedNotification creates a new Notification for the creation of a
// Piece logistics object.
func PieceCreatedNotification(pieceID, notificationID string) Notification {
	return Notification{
		APIHasEventType:           new(APILogisticsObjectCreatedNotificationEventType()),
		APIHasLogisticsObject:     NewPieceReference(pieceID),
		APIHasLogisticsObjectType: new(APIPieceLogisticsObjectType()),
		Context:                   new(APIContext()),
		Type:                      "api:Notification",
		ID:                        notificationID,
	}
}
