package env

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func String(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("[Requirements] Env %s is empty", key)
	}

	return value
}

func Duration(key string, defaultValue time.Duration) time.Duration {
	if raw := os.Getenv(key); raw != "" {
		raw = strings.ToLower(strings.TrimSpace(raw))

		format := time.Second
		switch raw[len(raw)-1] {
		case 'm':
			format = time.Minute
			fallthrough
		case 's':
			raw = raw[:len(raw)-1]
		}

		n, _ := strconv.ParseInt(raw, 10, 64)
		if n <= 0 {
			log.Fatalf("%s must be greater than 0", key)
		}

		return format * time.Duration(n)
	}

	return defaultValue
}

func Int(key string, defaultValue int) int {
	if raw := os.Getenv(key); raw != "" {
		n, _ := strconv.Atoi(raw)
		if n <= 0 {
			log.Fatalf("%s must be greater than 0", key)
		}

		return n
	}

	return defaultValue
}
