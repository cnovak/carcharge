package carcharge

import "testing"

func TestGetEnergy(t *testing.T) {
	result := GetEnergy()
	if result == "" {
		t.Fatalf(`GetEnergy = %q`, result)
	}

}
