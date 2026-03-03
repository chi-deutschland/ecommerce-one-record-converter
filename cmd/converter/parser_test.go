package main_test

import (
	"bytes"
	_ "embed"
	"testing"

	main "chi-deutschland.com/ecommerce-one-record-converter/cmd/converter"
	"chi-deutschland.com/ecommerce-one-record-converter/pkg/iata"
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

	mawbMap, err := main.ParseExcelData(file)
	if err != nil {
		t.Fatalf("parseExcelData returned error: %v", err)
	}

	expectedMAWB, err := iata.ParseMawb("112-39023246")
	if err != nil {
		t.Fatalf("failed to parse expected MAWB: %v", err)
	}

	_, ok := mawbMap[expectedMAWB]
	if !ok {
		t.Errorf("expected MAWB %s not found in result", expectedMAWB)
	}

	expectedBoxID := main.BoxID("340434780839197000")

	_, ok = mawbMap[expectedMAWB].Boxes[expectedBoxID]
	if !ok {
		t.Errorf("expected Box ID %s not found in MAWB %s", expectedBoxID, expectedMAWB)
	}

	expectedParcelID := main.ParcelID("340434763452009000")

	_, ok = mawbMap[expectedMAWB].Boxes[expectedBoxID].Parcels[expectedParcelID]
	if !ok {
		t.Errorf("expected Parcel ID %s not found in Box ID %s of MAWB %s",
			expectedParcelID, expectedBoxID, expectedMAWB)
	}

	expectedReferenceID := main.ReferenceID("BG-251006920E6LY7SX")

	_, ok = mawbMap[expectedMAWB].Boxes[expectedBoxID].Parcels[expectedParcelID].Pieces[expectedReferenceID]
	if !ok {
		t.Errorf("expected Reference ID %s not found in Parcel ID %s of Box ID %s in MAWB %s",
			expectedReferenceID, expectedParcelID, expectedBoxID, expectedMAWB)
	}

	const expectedBuyerCity = "Sigmaringen"

	buyerCity := mawbMap[expectedMAWB].Boxes[expectedBoxID].Parcels[expectedParcelID].Pieces[expectedReferenceID].
		BuyerCity

	if buyerCity != expectedBuyerCity {
		t.Errorf("expected Buyer City %s, got %s", expectedBuyerCity, buyerCity)
	}

	const expectedItemLen = 11

	itemLen := len(mawbMap[expectedMAWB].Boxes[expectedBoxID].Parcels[expectedParcelID].Pieces[expectedReferenceID].
		Items)

	if itemLen != expectedItemLen {
		t.Errorf("expected number of items %d, got %d", expectedItemLen, itemLen)
	}
}
