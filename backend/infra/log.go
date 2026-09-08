package infra

import "log"

func LogInfo(msg string, kv ...any) {
	log.Printf("[coordinator] "+msg, kv...)
}

func LogWarn(msg string, kv ...any) {
	log.Printf("[coordinator] WARN "+msg, kv...)
}

func LogError(msg string, kv ...any) {
	log.Printf("[coordinator] ERROR "+msg, kv...)
}
