package i18n

import "testing"

func TestTranslate(t *testing.T) {
	zh := New("zh-CN")
	en := New("en")
	if zh.T("tab.overview") == "" {
		t.Error("zh-CN tab.overview should not be empty")
	}
	if en.T("tab.overview") == "" {
		t.Error("en tab.overview should not be empty")
	}
	if zh.T("tab.overview") == en.T("tab.overview") {
		t.Error("zh-CN and en should differ for tab.overview")
	}
}

func TestFallback(t *testing.T) {
	i := New("unknown")
	if i.T("tab.overview") == "" {
		t.Error("should fallback to en for unknown locale")
	}
}

func TestMissingKey(t *testing.T) {
	i := New("en")
	if i.T("nonexistent.key") != "nonexistent.key" {
		t.Error("missing key should return key itself")
	}
}
