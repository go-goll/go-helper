package httphelper

import (
	"testing"
)

func TestGet(t *testing.T) {
	url := "http://192.168.0.133:30000/livez"

	query := make(map[string][]string)
	query["name"] = []string{"jack", "telsa"}
	query["age"] = []string{"12", "23"}

	req := &ClientRequest{
		Url:   url,
		Query: query,
	}

	resp := Get(req)
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}

	t.Log(resp.Code)
	t.Log(string(resp.Body))
}
