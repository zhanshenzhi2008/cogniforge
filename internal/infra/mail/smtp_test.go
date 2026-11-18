package mail

import "testing"

func TestSMTP_Enabled(t *testing.T) {
	if NewSMTP("", 465, "a@qq.com", "code", "a@qq.com").Enabled() {
		t.Fatal("empty host")
	}
	if NewSMTP("smtp.qq.com", 465, "", "code", "a@qq.com").Enabled() {
		t.Fatal("empty user")
	}
	if !NewSMTP("smtp.qq.com", 465, "a@qq.com", "auth-code", "CogniForge <a@qq.com>").Enabled() {
		t.Fatal("expected enabled")
	}
}

func TestExtractEmail(t *testing.T) {
	if got := extractEmail("CogniForge <a@qq.com>"); got != "a@qq.com" {
		t.Fatalf("got %q", got)
	}
	if got := extractEmail("a@qq.com"); got != "a@qq.com" {
		t.Fatalf("got %q", got)
	}
}
