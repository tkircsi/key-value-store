package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestJSON(t *testing.T) {
	header := http.Header{}
	headerKey := "Content-Type"
	headerValue := "application/json; charset=utf-8"
	header.Add(headerKey, headerValue)

	testCases := []struct {
		in     map[string]string
		header http.Header
		out    string
	}{
		{map[string]string{"hello": "world"}, header, `{"hello":"world"}`},
		{map[string]string{"a": "b"}, header, `{"a":"b"}`},
		//{map[string]string{";>*>": "!%%%/+"}, header, `{";>*>":"!%%%/+"}`},
		//{map[string]string{"aa:bb": "\"J:"}, header, `{"aa:bb":""J:"}`},
		//{make(chan bool), header, `{"error":"json: unsupported type: chan bool"}`},
	}

	for _, test := range testCases {
		recorder := httptest.NewRecorder()
		JSON(recorder, test.in)

		response := recorder.Result()
		defer response.Body.Close()

		got, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("error reading response body: %s", err)
		}

		if string(got) != test.out {
			t.Errorf("body: got %q, expected %q", string(got), test.out)
		}

		if contentType := response.Header.Get(headerKey); contentType != headerValue {
			t.Errorf("header:got %q, expected %q", contentType, headerValue)
		}
	}
}

func TestGet(t *testing.T) {
	makeStorage(t)
	defer cleanupStorage(t)

	// Generate test data file
	kvStore := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key4": "value4",
	}
	encodedStore := map[string]string{}
	for key, value := range kvStore {
		ek := base64.URLEncoding.EncodeToString([]byte(key))
		ev := base64.URLEncoding.EncodeToString([]byte(value))
		encodedStore[ek] = ev
	}
	fileContents, _ := json.Marshal(encodedStore)
	os.WriteFile(StoragePath+"/data.json", fileContents, 0644)

	// Test cases
	testCases := []struct {
		in  string
		out string
		err error
	}{
		{"key1", "value1", nil},
		{"key2", "value2", nil},
		{"key3", "", nil},
	}

	for _, test := range testCases {
		got, err := Get(context.Background(), test.in)
		if err != test.err {
			t.Errorf("unexpected error: %s", err)
		}

		if got != test.out {
			t.Errorf("key: %q, got %q, expected %q", test.in, got, test.out)
		}
	}
}

func TestGetSetDelete(t *testing.T) {
	makeStorage(t)
	defer cleanupStorage(t)
	ctx := context.Background()

	key := "key"
	value := "value"

	if out, err := Get(ctx, key); err != nil || out != "" {
		t.Fatalf("first Get returned unexpected result, out: %q, error: %s", out, err)
	}

	if err := Set(ctx, key, value); err != nil {
		t.Fatalf("Set returned unexpected error: %s", err)
	}

	if out, err := Get(ctx, key); err != nil || out != value {
		t.Fatalf("second Get returned unexpected result, out: %q, error: %s", out, err)
	}

	if err := Delete(ctx, key); err != nil {
		t.Fatalf("Delete returned unexpected error: %s", err)
	}

	if out, err := Get(ctx, key); err != nil || out != "" {
		t.Fatalf("Third Get returned unexpected result, out: %q, error: %s", out, err)
	}
}

func BenchmarkGet(b *testing.B) {
	makeStorage(b)
	defer cleanupStorage(b)

	b.ResetTimer()
	// Generate test data file
	kvStore := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key4": "value4",
	}
	encodedStore := map[string]string{}
	for key, value := range kvStore {
		ek := base64.URLEncoding.EncodeToString([]byte(key))
		ev := base64.URLEncoding.EncodeToString([]byte(value))
		encodedStore[ek] = ev
	}
	fileContents, _ := json.Marshal(encodedStore)
	os.WriteFile(StoragePath+"/data.json", fileContents, 0644)

	// Test cases
	testCases := []struct {
		in  string
		out string
		err error
	}{
		{"key1", "value1", nil},
		{"key2", "value2", nil},
		{"key3", "", nil},
	}

	for _, test := range testCases {
		got, err := Get(context.Background(), test.in)
		if err != test.err {
			b.Errorf("unexpected error: %s", err)
		}

		if got != test.out {
			b.Errorf("key: %q, got %q, expected %q", test.in, got, test.out)
		}
	}
}

func makeStorage(tb testing.TB) {
	err := os.Mkdir("./testdata", 0755)
	if err != nil && os.IsExist(err) {
		tb.Fatalf("couldn't create directory testdata: %s", err)
	}
	StoragePath = "./testdata"
}

func cleanupStorage(tb testing.TB) {
	if err := os.RemoveAll(StoragePath); err != nil {
		tb.Errorf("failed to delete storage path: %s", StoragePath)
	}
	StoragePath = DefaultStoragePath
}
