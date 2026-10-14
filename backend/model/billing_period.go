package model

type BillingPeriod struct {
	StartISO  string `json:"start_iso"`
	StartUnix int64  `json:"start_unix"`
	EndUnix   int64  `json:"end_unix"`
}
