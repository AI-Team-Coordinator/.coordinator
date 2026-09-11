package model

type LocaleFile struct {
	Version      int    `json:"version"`
	ChatLanguage string `json:"chat_language"`
	DocsLanguage string `json:"docs_language"`
}
