package service

import (
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/web/entity"
)

// subURIOverride is the panel's subscription settings round-trip: the settings
// page shows a derived subURI and saves every field back, so the derived value
// gets stored as if the admin had typed it.
func subURIShown(t *testing.T, s *SettingService, host string) (subURI, subJsonURI, subClashURI string) {
	t.Helper()
	raw, err := s.GetDefaultSettings(host)
	if err != nil {
		t.Fatalf("GetDefaultSettings: %v", err)
	}
	got := raw.(map[string]any)
	return got["subURI"].(string), got["subJsonURI"].(string), got["subClashURI"].(string)
}

func saveSubscriptionSettings(t *testing.T, s *SettingService, mutate func(*entity.AllSetting)) {
	t.Helper()
	all, err := s.GetAllSetting()
	if err != nil {
		t.Fatalf("GetAllSetting: %v", err)
	}
	mutate(all)
	if err := s.UpdateAllSetting(all, SecretClears{}); err != nil {
		t.Fatalf("UpdateAllSetting: %v", err)
	}
}

// A stored subURI whose path no longer matches the served subPath produced links
// that 404 on every fetch: the panel kept showing the old path after the admin
// renamed the subscription path, because a derived subURI had been saved back.
func TestSubURIFollowsServedPathAfterRename(t *testing.T) {
	initHappTestDB(t)
	s := &SettingService{}

	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) {
		all.SubEnable = true
		all.SubPath = "/firstpanel/"
		all.SubURI = ""
	})

	derived, _, _ := subURIShown(t, s, "example.com")
	if derived != "http://example.com:2096/firstpanel/" {
		t.Fatalf("derived subURI = %q", derived)
	}

	// The settings page saves back everything it displays, freezing the value.
	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) { all.SubURI = derived })

	// The admin later renames the subscription path.
	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) { all.SubPath = "/secondpanel/" })

	got, _, _ := subURIShown(t, s, "example.com")
	if want := "http://example.com:2096/secondpanel/"; got != want {
		t.Errorf("subURI = %q, want %q — the link the panel hands out must be the path the sub server serves", got, want)
	}
}

// The reverse-proxy case the setting exists for: a path the panel never served
// is an explicit rewrite and must keep winning.
func TestSubURIHonorsReverseProxyPath(t *testing.T) {
	initHappTestDB(t)
	s := &SettingService{}

	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) {
		all.SubEnable = true
		all.SubPath = "/internal-sub/"
		all.SubURI = "https://public.example/rewritten/"
	})

	got, _, _ := subURIShown(t, s, "example.com")
	if want := "https://public.example/rewritten/"; got != want {
		t.Errorf("subURI = %q, want the configured reverse-proxy URL %q", got, want)
	}
}

// A proxy URL that happens to sit on the served path is still the admin's, and
// renaming the path must carry its scheme/host/port along rather than dropping
// the override back to the listen host.
func TestStaleSubURIRebuildsOnConfiguredHost(t *testing.T) {
	initHappTestDB(t)
	s := &SettingService{}

	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) {
		all.SubEnable = true
		all.SubPath = "/live/"
		all.SubURI = "https://sub.example.com:8443/live/"
	})
	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) { all.SubPath = "/renamed/" })

	got, _, _ := subURIShown(t, s, "example.com")
	if want := "https://sub.example.com:8443/renamed/"; got != want {
		t.Errorf("subURI = %q, want %q", got, want)
	}
}

// A malformed stored value can't be parsed, so it keeps the panel's derive
// fallback instead of being rebound to a path it never named.
func TestMalformedSubURIFallsBackToDerivedBase(t *testing.T) {
	initHappTestDB(t)
	s := &SettingService{}

	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) {
		all.SubEnable = true
		all.SubPath = "/live/"
		all.SubURI = "not-a-url"
	})

	all, err := s.GetAllSetting()
	if err != nil {
		t.Fatal(err)
	}
	if all.SubURI != "not-a-url" {
		t.Errorf("subURI = %q, want the stored value untouched", all.SubURI)
	}

	got, _, _ := subURIShown(t, s, "example.com")
	if want := "not-a-url"; got != want {
		t.Errorf("subURI = %q, want %q", got, want)
	}
}

// JSON and Clash carry their own stored URIs, so each must follow its own served
// path rather than the raw subscription's.
func TestFormatURIsFollowTheirOwnServedPaths(t *testing.T) {
	initHappTestDB(t)
	s := &SettingService{}

	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) {
		all.SubEnable = true
		all.SubPath = "/live/"
		all.SubURI = ""
		all.SubJsonEnable = true
		all.SubJsonPath = "/json-live/"
		all.SubJsonURI = "http://example.com:2096/json-live/"
		all.SubClashEnable = true
		all.SubClashPath = "/clash-live/"
		all.SubClashURI = "http://example.com:2096/clash-live/"
	})
	saveSubscriptionSettings(t, s, func(all *entity.AllSetting) {
		all.SubJsonPath = "/json-renamed/"
		all.SubClashPath = "/clash-renamed/"
	})

	_, jsonURI, clashURI := subURIShown(t, s, "example.com")
	if want := "http://example.com:2096/json-renamed/"; jsonURI != want {
		t.Errorf("subJsonURI = %q, want %q", jsonURI, want)
	}
	if want := "http://example.com:2096/clash-renamed/"; clashURI != want {
		t.Errorf("subClashURI = %q, want %q", clashURI, want)
	}
}
