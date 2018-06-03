package kong

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func postWebhooks(k *KongWebhook, eventBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(eventBody))
	w := httptest.NewRecorder()
	w.Code = 500

	k.eventHandler(w, req)

	return w
}

func TestRequestEvent(t *testing.T) {
	var acc testutil.Accumulator
	k := &KongWebhook{Path: "/kong", acc: &acc}
	resp := postWebhooks(k, RequestEventJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST deploy returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"value": 1,
	}

	tags := map[string]string{
		"upstream_uri": "/",
		"request_uri": "/get",
		"request_method": "GET",
		"client": "TODO"
		"response_status": 200,
		"response_size": 434
	}

	acc.AssertContainsTaggedFields(t, "rollbar_webhooks", fields, tags)
}

func TestUnknowItem(t *testing.T) {
	rb := &RollbarWebhook{Path: "/rollbar"}
	resp := postWebhooks(rb, UnknowJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST unknow returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}
}
