package main

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNothing(t *testing.T) {
	spend, _ := hex.DecodeString("cc1d00d799a1d626efb6fe93eda827312e1b07e25598d692703580a02e9f150f")
	spendkey := newKeyPair()
	copy(spendkey.Priv[:], spend)
	spendkey.Pub = spendkey.Priv.PubKey()

	view, _ := hex.DecodeString("9edadf7c3a39db4d6948842913b6a95b90647306d6fa270e53758eb67e82ff05")
	viewkey := newKeyPair()
	copy(viewkey.Priv[:], view)
	viewkey.Pub = viewkey.Priv.PubKey()

	w := wallet{SpendKey: spendkey, ViewKey: viewkey}
	if w.Address() != "Sumoo5LoLuANaPkyT7gHtuF1YtAETkUpLRTJeNzwCFX48Ptd48A2KaSazqxinmwjB2EVxRw9CakSWdQKigoBG2bRAPSqwoDDcb9" {
		fmt.Printf("Bad Address: %s\n", w.Address())
		t.Fail()
	}
}
