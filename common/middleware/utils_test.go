package middleware

import (
	"testing"
)

func TestDecrypt(t *testing.T) {
	// Arrange
	tst_key, _ := formKey(`!@#$%^&*()_+?><":{}|OADm+-*/23870adw~|\`)
	cases := []struct{
		PlainText string
	}{
		{ // Valid
			"special secret text",
		},
		{ // Valid (symbols should be fine)
			`!@#$%^secret&*()_+?><":{}|OADm+-*text/23870adw~|\`,
		},
	}

	for _, tst := range cases {
		// Act
		cypher, err := encrypt([]byte(tst.PlainText), &tst_key)
		if err != nil {
			t.Error(err)
		}
		plain, err := decrypt(cypher, &tst_key)
		plaintxt := string(plain)

		// Assert
		if plaintxt != tst.PlainText {
			t.Error("Expected", tst.PlainText, "got", plaintxt)
		}
	}
}

func TestFormKey(t *testing.T) {
	// Arrange
	required_len := 32
	cases := []struct{
		Input string
		ErrRaised bool
	}{
		{ // Valid
			"abcdefghijklmnopqrstuvwxyz_0123456789",
			false,
		},
		{ // Valid (symbols should be fine)
			`!@#$%^&*()_+?><":{}|OADm+-*/23870adw~|\`,
			false,
		},
		{ // Invalid (key too short)
			"abc",
			true,
		},
	}

	for _, tst := range cases {
		// Act
		key, err := formKey(tst.Input)

		// Assert
		if err != nil && !tst.ErrRaised {
			t.Error(err)
		}
		if len(key) != required_len {
			t.Error("Key len expected to be", required_len, "got", len(key))
		}
	}
}