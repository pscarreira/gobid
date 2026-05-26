package validator

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator interface {
	Valid(context.Context) Evaluator
}

type Evaluator map[string]string

func (e *Evaluator) AddFieldError(key, message string) {
	if *e == nil {
		*e = make(map[string]string)
	}
	if _, exists := (*e)[key]; !exists {
		(*e)[key] = message
	}
}

func (e *Evaluator) CheckField(ok bool, key, message string) {
	if !ok {
		e.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, max int) bool {
	return len(value) <= max
}

func MinChars(value string, min int) bool {
	return len(value) >= min
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func IsEmail(value string) bool {
	return Matches(value, EmailRX)
}

func Positive(value float64) bool {
	return value > 0
}

func Future(t time.Time) bool {
	return t.After(time.Now())
}

func FutureWithMinDuration(t time.Time, min time.Duration) bool {
	nowUTC := time.Now().UTC()
	minimumTimeUTC := nowUTC.Add(min)

	slog.Info(
		"Auction end validation window (UTC)",
		"now_utc", nowUTC.Format(time.RFC3339),
		"minimum_end_utc", minimumTimeUTC.Format(time.RFC3339),
		"candidate_end_utc", t.UTC().Format(time.RFC3339),
	)

	return t.UTC().After(minimumTimeUTC)
}
