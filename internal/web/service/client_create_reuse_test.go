package service

import (
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func settingsHoldUUID(t *testing.T, inboundSvc *InboundService, inboundId int, uuid string) bool {
	t.Helper()
	ib, err := inboundSvc.GetInbound(inboundId)
	if err != nil {
		t.Fatalf("GetInbound %d: %v", inboundId, err)
	}
	return strings.Contains(ib.Settings, uuid)
}

func TestCreateRepeatKeepsExistingUUID(t *testing.T) {
	setupBulkDB(t)
	svc := &ClientService{}
	inboundSvc := &InboundService{}

	ibA := mkInbound(t, 21001, model.VLESS, `{"clients":[]}`)
	ibB := mkInbound(t, 21002, model.VLESS, `{"clients":[]}`)

	const originalUUID = "aaaaaaaa-1111-2222-3333-444444444444"
	if _, err := svc.Create(inboundSvc, &ClientCreatePayload{
		Client:     model.Client{Email: "repeat@x", ID: originalUUID, SubID: "sub-repeat", Enable: true},
		InboundIds: []int{ibA.Id},
	}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if rec := lookupClientRecord(t, "repeat@x"); rec.UUID != originalUUID {
		t.Fatalf("record UUID after first Create = %q, want %q", rec.UUID, originalUUID)
	}

	if _, err := svc.Create(inboundSvc, &ClientCreatePayload{
		Client:     model.Client{Email: "repeat@x", SubID: "sub-repeat", Enable: true},
		InboundIds: []int{ibB.Id},
	}); err != nil {
		t.Fatalf("repeat Create: %v", err)
	}

	if rec := lookupClientRecord(t, "repeat@x"); rec.UUID != originalUUID {
		t.Fatalf("record UUID after repeat Create = %q, want %q", rec.UUID, originalUUID)
	}
	if !settingsHoldUUID(t, inboundSvc, ibA.Id, originalUUID) {
		t.Fatalf("inbound A settings lost the original UUID")
	}
	if !settingsHoldUUID(t, inboundSvc, ibB.Id, originalUUID) {
		t.Fatalf("inbound B settings did not reuse the original UUID")
	}
}

// With the unique-subId check off by default, two clients may share one subId:
// each Create must succeed and both must land on the inbound.
func TestClientCreateDuplicateSubIDAllowedByDefault(t *testing.T) {
	setupBulkDB(t)
	svc := &ClientService{}
	inboundSvc := &InboundService{}

	ib := mkInbound(t, 21003, model.VLESS, `{"clients":[]}`)
	const shared = "sub-shared-create"
	for i, email := range []string{"c1@x", "c2@x"} {
		if _, err := svc.Create(inboundSvc, &ClientCreatePayload{
			Client:     model.Client{Email: email, SubID: shared, Enable: true},
			InboundIds: []int{ib.Id},
		}); err != nil {
			t.Fatalf("Create %d with shared subId should succeed by default: %v", i, err)
		}
	}
	if emails := settingsClientEmails(t, ib.Id); len(emails) != 2 {
		t.Fatalf("expected both clients on the inbound, got %v", emails)
	}
}

func TestClientCreateDuplicateSubIDRejectedWhenEnabled(t *testing.T) {
	setupBulkDB(t)
	svc := &ClientService{}
	inboundSvc := &InboundService{}
	if err := (&SettingService{}).SetCheckUniqueSubId(true); err != nil {
		t.Fatalf("enable unique subId check: %v", err)
	}

	ib := mkInbound(t, 21004, model.VLESS, `{"clients":[]}`)
	const shared = "sub-shared-create"
	if _, err := svc.Create(inboundSvc, &ClientCreatePayload{
		Client:     model.Client{Email: "c1@x", SubID: shared, Enable: true},
		InboundIds: []int{ib.Id},
	}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if _, err := svc.Create(inboundSvc, &ClientCreatePayload{
		Client:     model.Client{Email: "c2@x", SubID: shared, Enable: true},
		InboundIds: []int{ib.Id},
	}); err == nil {
		t.Fatal("Create with colliding subId succeeded, want error")
	}
}
