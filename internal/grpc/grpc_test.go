package grpc

import (
	"testing"
)

func TestSendMessage(t *testing.T) {
	var testData = []byte("test")
	resp, err := SendMessage(testData, 1, 1)
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if resp == "" {
		t.Fatalf("SendMessage failed: empty response")
	}
	t.Logf("SendMessage: %s", resp)
}
