package api

import (
	"log"
	"os"
	"testing"
)

var client *ScalewayAPI

func TestMain(m *testing.M) {
	if os.Getenv("SCALEWAY_ORGANIZATION") != "" && os.Getenv("SCALEWAY_ACCESS_KEY") != "" {
		c, err := New(os.Getenv("SCALEWAY_ORGANIZATION"), os.Getenv("SCALEWAY_ACCESS_KEY"), "par1")
		if err != nil {
			log.Printf("Unable to create scaleway client")
			os.Exit(1)
		}
		client = c
	}
	code := m.Run()
	os.Exit(code)
}
