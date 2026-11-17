package mail

import "testing"

func TestResend_Enabled(t *testing.T) {
	if NewResend("", "a@b.com").Enabled() {
		t.Fatal("empty key should disable")
	}
	if NewResend("re_x", "").Enabled() {
		t.Fatal("empty from should disable")
	}
	if !NewResend("re_x", "CogniForge <noreply@example.com>").Enabled() {
		t.Fatal("expected enabled")
	}
	if (Nop{}).Enabled() {
		t.Fatal("nop should be disabled")
	}
}
