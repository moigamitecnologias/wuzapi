package main

import "testing"

func TestDevicePropsOsForWhatsApp(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"meucliente.co", "Chrome (meucliente.co)"},
		{"", "Chrome (Linux)"},
		{"Chrome (Linux)", "Chrome (Linux)"},
	}
	for _, tc := range cases {
		if got := devicePropsOsForWhatsApp(tc.in); got != tc.want {
			t.Errorf("devicePropsOsForWhatsApp(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPairPhoneDisplayNameFromOSName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"meucliente.co", "Chrome (Linux)"},
		{"meuagente.co", "Chrome (Linux)"},
		{"", "Chrome (Linux)"},
		{"   ", "Chrome (Linux)"},
		{"Chrome (Linux)", "Chrome (Linux)"},
		{"Firefox (Windows)", "Firefox (Windows)"},
		{"  Chrome (macOS)  ", "Chrome (macOS)"},
	}
	for _, tc := range cases {
		if got := pairPhoneDisplayNameFromOSName(tc.in); got != tc.want {
			t.Errorf("pairPhoneDisplayNameFromOSName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
