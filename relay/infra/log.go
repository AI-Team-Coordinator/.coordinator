package infra

import "log"

func LogInfo(msg string, kv ...any) {
	log.Printf("[relay] "+msg, kv...)
}

func LogError(msg string, kv ...any) {
	log.Printf("[relay] ERROR "+msg, kv...)
}

func LogWarn(msg string, kv ...any) {
	log.Printf("[relay] WARN "+msg, kv...)
}
