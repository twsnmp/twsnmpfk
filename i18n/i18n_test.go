package i18n

import (
	"testing"
)

func TestDetectOSLang(t *testing.T) {
	l := detectOSLang()
	if l == "" {
		t.Errorf("detectOSLang() returned empty string")
	}
	t.Logf("detected OS language: %s", l)
}
