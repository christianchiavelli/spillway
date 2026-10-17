package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRefusesAnUnknownCommand(t *testing.T) {
	err := run(t.Context(), []string{"serv"}, func(string) string { return "" }, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), `unknown command "serv"`) {
		t.Errorf("run(serv) = %v", err)
	}
}

func TestServeListsEveryBadSettingBeforeStarting(t *testing.T) {
	env := map[string]string{"LOG_FORMAT": "xml"}
	err := run(t.Context(), []string{"serve"}, func(k string) string { return env[k] }, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL: required") || !strings.Contains(err.Error(), "LOG_FORMAT") {
		t.Errorf("run(serve) = %v", err)
	}
}

func TestHealthcheckFailsWhenNothingAnswers(t *testing.T) {
	env := map[string]string{"ADMIN_ADDR": "127.0.0.1:1"}
	if err := run(t.Context(), []string{"healthcheck"}, func(k string) string { return env[k] }, &bytes.Buffer{}); err == nil {
		t.Error("healthcheck passed with no server")
	}
}
