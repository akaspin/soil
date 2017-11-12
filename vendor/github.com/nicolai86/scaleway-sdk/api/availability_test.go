package api

import (
	"testing"
)

func TestScalewayAPI_GetServerAvailabilities(t *testing.T) {
	if client == nil {
		t.Skip("skipping GetServerAvailabilities due to missing client credentials")
	}
	availabilities, err := client.GetServerAvailabilities()
	if err != nil {
		t.Errorf("failed to get server availabilities: %v", err.Error())
	}
	if len(availabilities.CommercialTypes()) == 0 {
		t.Errorf("Expected commercial types, but got none")
	}
}
