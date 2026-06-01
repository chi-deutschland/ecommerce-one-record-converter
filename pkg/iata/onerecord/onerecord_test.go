package onerecord_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"

	"chi-deutschland.com/ecommerce-one-record-converter/pkg/iata/onerecord"
)

var (
	//go:embed testdata/CargoAddress.json
	CargoAddressJSON []byte

	//go:embed testdata/CargoCodeListElement.json
	CargoCodeListElementJSON []byte

	//go:embed testdata/CargoContext.json
	CargoContextJSON []byte

	//go:embed testdata/CargoInvoiceNumber.json
	CargoInvoiceNumberJSON []byte

	//go:embed testdata/CargoItem.json
	CargoItemJSON []byte

	//go:embed testdata/CargoLocation.json
	CargoLocationJSON []byte

	//go:embed testdata/CargoOtherIdentifier.json
	CargoOtherIdentifierJSON []byte

	//go:embed testdata/CargoParcelPiece.json
	CargoParcelPieceJSON []byte

	//go:embed testdata/CargoPiece.json
	CargoPieceJSON []byte

	//go:embed testdata/CargoProduct.json
	CargoProductJSON []byte

	//go:embed testdata/CargoShipment.json
	CargoShipmentJSON []byte

	//go:embed testdata/CargoValue.json
	CargoValueJSON []byte

	//go:embed testdata/HSCode.json
	HSCodeJSON []byte

	//go:embed testdata/KilogramWeightUnit.json
	KilogramWeightUnitJSON []byte

	//go:embed testdata/NewPieceReference.json
	NewPieceReferenceJSON []byte

	//go:embed testdata/PieceUnit.json
	PieceUnitJSON []byte

	//go:embed testdata/ReferenceIDIdentifier.json
	ReferenceIDIdentifierJSON []byte

	//go:embed testdata/PieceCreatedNotification.json
	PieceCreatedNotificationJSON []byte
)

// TestOneRecordFunctions is a table-driven test for all one record functions.
func TestOneRecordFunctions(t *testing.T) {
	const germanyCountryCode = "DE"

	tests := []struct {
		name     string
		gotFn    func() (any, error)
		expected []byte
	}{
		{
			name: "CargoContext",
			gotFn: func() (any, error) {
				return onerecord.CargoContext(), nil
			},
			expected: CargoContextJSON,
		},
		{
			name: "CargoShipment",
			gotFn: func() (any, error) {
				pieces := []onerecord.Piece{{ID: "piece1"}, {ID: "piece2"}}
				parties := []onerecord.Party{{Type: "cargo:Party"}}

				return onerecord.CargoShipment(pieces, parties), nil
			},
			expected: CargoShipmentJSON,
		},
		{
			name: "CargoOtherIdentifier",
			gotFn: func() (any, error) {
				return onerecord.CargoOtherIdentifier("testType", "testValue"), nil
			},
			expected: CargoOtherIdentifierJSON,
		},
		{
			name: "CargoProduct",
			gotFn: func() (any, error) {
				return onerecord.CargoProduct("SKU123", "1234.56", "Test Product"), nil
			},
			expected: CargoProductJSON,
		},
		{
			name: "HSCode",
			gotFn: func() (any, error) {
				return onerecord.HSCode("1234.56"), nil
			},
			expected: HSCodeJSON,
		},
		{
			name: "CargoItem",
			gotFn: func() (any, error) {
				params := onerecord.ItemParams{
					Product:               onerecord.CargoProduct("SKU123", "1234.56", "Test Product"),
					ItemQuantity:          "10",
					UnitPrice:             "100.50",
					Currency:              "USD",
					WeightInKg:            "5.5",
					InvoiceNumber:         "INV-001",
					ProductionCountryCode: germanyCountryCode,
					TargetCountryCode:     "US",
				}

				return onerecord.CargoItem(params), nil
			},
			expected: CargoItemJSON,
		},
		{
			name: "CargoInvoiceNumber",
			gotFn: func() (any, error) {
				return onerecord.CargoInvoiceNumber("INV-002"), nil
			},
			expected: CargoInvoiceNumberJSON,
		},
		{
			name: "KilogramWeightUnit",
			gotFn: func() (any, error) {
				return onerecord.KilogramWeightUnit(), nil
			},
			expected: KilogramWeightUnitJSON,
		},
		{
			name: "CargoValue",
			gotFn: func() (any, error) {
				unit := onerecord.KilogramWeightUnit()

				return onerecord.CargoValue("42.0", unit), nil
			},
			expected: CargoValueJSON,
		},
		{
			name: "PieceUnit",
			gotFn: func() (any, error) {
				return onerecord.PieceUnit(), nil
			},
			expected: PieceUnitJSON,
		},
		{
			name: "NewPieceReference",
			gotFn: func() (any, error) {
				return onerecord.NewPieceReference("piece-ref-1"), nil
			},
			expected: NewPieceReferenceJSON,
		},
		{
			name: "CargoPiece",
			gotFn: func() (any, error) {
				return onerecord.CargoPiece("ref-123",
					[]string{"item1", "item2"},
					[]onerecord.Party{{Type: "cargo:Party"}}), nil
			},
			expected: CargoPieceJSON,
		},
		{
			name: "ReferenceIDIdentifier",
			gotFn: func() (any, error) {
				return onerecord.ReferenceIDIdentifier("ref-456"), nil
			},
			expected: ReferenceIDIdentifierJSON,
		},
		{
			name: "CargoParcelPiece",
			gotFn: func() (any, error) {
				pieces := []onerecord.Piece{{ID: "p1"}, {ID: "p2"}}

				return onerecord.CargoParcelPiece(pieces, "parcel-789"), nil
			},
			expected: CargoParcelPieceJSON,
		},
		{
			name: "CargoAddress",
			gotFn: func() (any, error) {
				return onerecord.CargoAddress([]string{"Street 1", "Street 2"},
					"12345",
					"Berlin",
					germanyCountryCode), nil
			},
			expected: CargoAddressJSON,
		},
		{
			name: "CargoCodeListElement",
			gotFn: func() (any, error) {
				return onerecord.CargoCodeListElement(germanyCountryCode,
					onerecord.CodeListReferenceCountry,
					"2026"), nil
			},
			expected: CargoCodeListElementJSON,
		},
		{
			name: "CargoLocation",
			gotFn: func() (any, error) {
				return onerecord.CargoLocation(new(onerecord.CargoAddress([]string{"Street 1"},
					"54321",
					"Munich",
					germanyCountryCode))), nil
			},
			expected: CargoLocationJSON,
		},
		{
			name: "PieceCreatedNotification",
			gotFn: func() (any, error) {
				const id = "http://localhost:8080/logistics-objects/6289da4a-6073-4827-8afc-58e97d5fa6bc"

				return onerecord.PieceCreatedNotification(id, id+"_123456789"), nil
			},
			expected: PieceCreatedNotificationJSON,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.gotFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotBytes, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("failed to marshal got value to JSON: %v", err)
			}

			var gotComparer map[string]any

			err = json.Unmarshal(gotBytes, &gotComparer)
			if err != nil {
				t.Fatalf("failed to unmarshal got JSON: %v", err)
			}

			var expectedComparer map[string]any

			err = json.Unmarshal(test.expected, &expectedComparer)
			if err != nil {
				t.Fatalf("failed to unmarshal expected JSON: %v", err)
			}

			if !reflect.DeepEqual(gotComparer, expectedComparer) {
				t.Errorf("got and expected JSON do not match.\nGot: %s\nExpected: %s",
					string(gotBytes), string(test.expected))
			}
		})
	}
}
