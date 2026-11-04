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

func TestTransSensorState(t *testing.T) {
	orig := GetLang()
	defer SetLang(orig)

	SetLang("en")
	if got := Trans("Change sensor state:%s,%s,%s"); got != "Change sensor state:%s,%s,%s" {
		t.Errorf("Trans(en) = %q, want %q", got, "Change sensor state:%s,%s,%s")
	}

	SetLang("ja")
	if got := Trans("Change sensor state:%s,%s,%s"); got != "センサー状態変化:%s,%s,%s" {
		t.Errorf("Trans(ja) = %q, want %q", got, "センサー状態変化:%s,%s,%s")
	}
}
