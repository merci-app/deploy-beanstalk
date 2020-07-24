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

func Duration(key string, defaultValue time.Duration, min ...time.Duration) time.Duration {
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

		if len(min) == 0 {
			min = append(min, time.Second)
		}

		n, _ := strconv.ParseInt(raw, 10, 64)
		value := format * time.Duration(n)

		if value < min[0] {
			log.Fatalf("%s must be greater than or equal to %v", key, min[0])
		}

		return value
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
