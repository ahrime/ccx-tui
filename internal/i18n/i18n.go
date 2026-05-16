package i18n

type I18n struct {
	locale string
	msgs   map[string]string
}

var locales = map[string]map[string]string{
	"zh-CN": zhCN,
	"en":    enUS,
}

func New(locale string) *I18n {
	msgs, ok := locales[locale]
	if !ok {
		msgs = locales["en"]
	}
	return &I18n{locale: locale, msgs: msgs}
}

func (i *I18n) T(key string) string {
	if v, ok := i.msgs[key]; ok {
		return v
	}
	if v, ok := locales["en"][key]; ok {
		return v
	}
	return key
}

func (i *I18n) Locale() string {
	return i.locale
}
