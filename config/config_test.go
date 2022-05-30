package config

import (
	"fmt"
	"testing"
)

func TestFromData_Globals(t *testing.T) {
	yaml := fmt.Sprintf("globals:\n  fiat: EUR")
	c, err := FromData([]byte(yaml))
	if err != nil {
		t.Error(err)
	}
	if c.globals.Fiat != "EUR" {
		t.Errorf("Fiat is not 'ERR' is '%v'", c.globals.Fiat)
	}
}

func TestFromData_Wallets(t *testing.T) {
	yaml := fmt.Sprintf("wallets:\n  - name: test")
	c, err := FromData([]byte(yaml))
	if err != nil {
		t.Error(err)
	}
	wallets := c.GetWallets()
	if len(wallets) != 1 {
		t.Errorf("Wallet size is not 1")
	}
	if wallets[0].Name != "test" {
		t.Errorf("Wallet name is not 'test' is '%v'", wallets[0])
	}
}
