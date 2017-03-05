package middleware

import (
	"testing"
)

func TestNatsURL(t *testing.T) {
	// Arrange
	cases := []struct{
		Settings MiddlewareSettings
		Expected string
	}{
		{
			MiddlewareSettings{
				User: "batman",
				Pass: "secret!",
				Host: "123.456.123.987",
				Port: 34000,
			},
			"nats://batman:secret!@123.456.123.987:34000",
		},
		{
			MiddlewareSettings{
				User: "",
				Pass: "",
				Host: "123.456.123.987",
				Port: 123,
			},
			"nats://123.456.123.987:123",
		},
		{
			MiddlewareSettings{
				User: "",
				Pass: "password",
				Host: "8.456.75.987",
				Port: 23157,
			},
			"nats://8.456.75.987:23157",
		},
	}

	for _, tst := range cases {
		// Act
		uri := natsUrl(tst.Settings)

		// Assert
		if uri != tst.Expected {
			t.Error("Expected", tst.Expected, "got", uri)
		}
	}
}
