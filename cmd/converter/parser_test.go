package main_test

import (
	"bytes"
	_ "embed"
	"testing"

	main "chi-deutschland.com/ecommerce-one-record-converter/cmd/converter"
	"github.com/xuri/excelize/v2"
)

//go:embed testdata/fake_ecommerce_data.xlsx
var fakeEcommerceData []byte

func TestParseExcelData_EmbeddedFile(t *testing.T) {
	file, err := excelize.OpenReader(bytes.NewReader(fakeEcommerceData))
	if err != nil {
		t.Fatalf("failed to open embedded excel file: %v", err)
	}

	defer func() {
		_ = file.Close()
	}()

	boxes, err := main.ParseExcelData(file)
	if err != nil {
		t.Fatalf("parseExcelData returned error: %v", err)
	}

	expectedBoxID := main.BoxID("340434780839197000")

	_, ok := boxes[expectedBoxID]
	if !ok {
		t.Errorf("expected Box ID %s not found in boxes", expectedBoxID)
	}

	expectedParcelID := main.ParcelID("340434763452009000")

	_, ok = boxes[expectedBoxID].Parcels[expectedParcelID]
	if !ok {
		t.Errorf("expected Parcel ID %s not found in Box ID %s", expectedParcelID, expectedBoxID)
	}

	expectedReferenceID := main.ReferenceID("BG-251006920E6LY7SX")

	_, ok = boxes[expectedBoxID].Parcels[expectedParcelID].Pieces[expectedReferenceID]
	if !ok {
		t.Errorf("expected Reference ID %s not found in Parcel ID %s of Box ID %s",
			expectedReferenceID, expectedParcelID, expectedBoxID)
	}

	const expectedBuyerCity = "Sigmaringen"

	buyerCity := boxes[expectedBoxID].Parcels[expectedParcelID].Pieces[expectedReferenceID].BuyerCity

	if buyerCity != expectedBuyerCity {
		t.Errorf("expected Buyer City %s, got %s", expectedBuyerCity, buyerCity)
	}

	const expectedItemLen = 11

	itemLen := len(boxes[expectedBoxID].Parcels[expectedParcelID].Pieces[expectedReferenceID].Items)

	if itemLen != expectedItemLen {
		t.Errorf("expected number of items %d, got %d", expectedItemLen, itemLen)
	}
}
