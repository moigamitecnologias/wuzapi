package main

import "testing"

func TestDevicePropsOsForWhatsApp(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"meucliente.co", "Chrome (Mac OS)"},
		{"", "Chrome (Mac OS)"},
		{"Chrome (Linux)", "Chrome (Mac OS)"},
		{"Firefox (Windows)", "Chrome (Mac OS)"},
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
		{"meucliente.co", "Chrome (Mac OS)"},
		{"meuagente.co", "Chrome (Mac OS)"},
		{"", "Chrome (Mac OS)"},
		{"   ", "Chrome (Mac OS)"},
		{"Chrome (Linux)", "Chrome (Mac OS)"},
		{"Firefox (Windows)", "Chrome (Mac OS)"},
		{"  Chrome (macOS)  ", "Chrome (Mac OS)"},
	}
	for _, tc := range cases {
		if got := pairPhoneDisplayNameFromOSName(tc.in); got != tc.want {
			t.Errorf("pairPhoneDisplayNameFromOSName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPairPhoneClientDisplayName(t *testing.T) {
	t.Parallel()
	if got := pairPhoneClientDisplayName(); got != "Chrome (Mac OS)" {
		t.Errorf("pairPhoneClientDisplayName() = %q, want Chrome (Mac OS)", got)
	}
}
