package ocrparser

import (
	"reflect"
	"testing"
)

func TestParsePlanExample(t *testing.T) {
	raw := `WARUNG MAKAN ABC

Nasi Goreng       25.000
Mie Goreng        20.000
Es Teh             5.000
Ayam Goreng       30.000
--------------------------
Subtotal          80.000
Tax                8.000
Service            4.000
Total              92.000`

	got := Parse(raw)

	if got.MerchantName != "WARUNG MAKAN ABC" {
		t.Errorf("MerchantName = %q, want %q", got.MerchantName, "WARUNG MAKAN ABC")
	}
	if len(got.Items) != 4 {
		t.Fatalf("len(Items) = %d, want 4", len(got.Items))
	}
	wantNames := []string{"Nasi Goreng", "Mie Goreng", "Es Teh", "Ayam Goreng"}
	for i, w := range wantNames {
		if got.Items[i].Name != w {
			t.Errorf("Items[%d].Name = %q, want %q", i, got.Items[i].Name, w)
		}
	}
	if got.Items[0].UnitPrice != 25000 {
		t.Errorf("Items[0].UnitPrice = %d, want 25000", got.Items[0].UnitPrice)
	}
	if got.Subtotal != 80000 || got.Tax != 8000 || got.ServiceCharge != 4000 || got.Total != 92000 {
		t.Errorf("summary mismatch: %+v", got)
	}
}

func TestParseQuantityPatterns(t *testing.T) {
	raw := "Es Teh 3x5.000 15.000\nKopi 2 @ 10.000 20.000"
	got := Parse(raw)

	if len(got.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(got.Items))
	}
	if got.Items[0].Quantity != 3 || got.Items[0].TotalPrice != 15000 {
		t.Errorf("Items[0] = %+v, want qty=3 total=15000", got.Items[0])
	}
	if got.Items[1].Quantity != 2 || got.Items[1].UnitPrice != 10000 {
		t.Errorf("Items[1] = %+v, want qty=2 unit=10000", got.Items[1])
	}
}

func TestParseAmountFormats(t *testing.T) {
	raw := "Subtotal Rp1.250.000\nTotal 25000\nDiskon 2,500"
	got := Parse(raw)
	if got.Subtotal != 1250000 {
		t.Errorf("Subtotal = %d, want 1250000", got.Subtotal)
	}
	if got.Total != 25000 {
		t.Errorf("Total = %d, want 25000", got.Total)
	}
	if got.Discount != 2500 {
		t.Errorf("Discount = %d, want 2500", got.Discount)
	}
}

func TestParseEmpty(t *testing.T) {
	got := Parse("")
	if len(got.Items) != 0 || got.MerchantName != "" {
		t.Errorf("expected empty parse, got %+v", got)
	}
}

func TestParseDecorationOnly(t *testing.T) {
	got := Parse("-----\n=====\n* * *")
	if len(got.Items) != 0 {
		t.Errorf("decoration parsed as items: %+v", got.Items)
	}
}

func TestParsedStructFields(t *testing.T) {
	p := Parsed{}
	if reflect.TypeOf(p.Subtotal).Kind() != reflect.Int64 {
		t.Error("Subtotal should be int64")
	}
}
