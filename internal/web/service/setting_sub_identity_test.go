package service

import "testing"

func TestSubShowIdentityOnAllLinksDefaultsAndPersists(t *testing.T) {
	setupSettingTestDB(t)
	s := &SettingService{}

	// New fork default: identity tokens stay on every subscription link.
	if got, err := s.GetSubShowIdentityOnAllLinks(); err != nil || !got {
		t.Fatalf("missing setting = %t, %v; want true, nil", got, err)
	}
	settings, err := s.GetAllSetting()
	if err != nil {
		t.Fatal(err)
	}
	if !settings.SubShowIdentityOnAllLinks {
		t.Fatal("GetAllSetting returned false for a missing setting")
	}

	settings.SubShowIdentityOnAllLinks = false
	if err := s.UpdateAllSetting(settings, SecretClears{}); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetSubShowIdentityOnAllLinks(); err != nil || got {
		t.Fatalf("persisted setting = %t, %v; want false, nil", got, err)
	}

	settings, err = s.GetAllSetting()
	if err != nil {
		t.Fatal(err)
	}
	settings.SubShowIdentityOnAllLinks = true
	if err := s.UpdateAllSetting(settings, SecretClears{}); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetSubShowIdentityOnAllLinks(); err != nil || !got {
		t.Fatalf("persisted setting = %t, %v; want true, nil", got, err)
	}
}
