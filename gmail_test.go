package gmail

import "testing"

func TestInit(t *testing.T) {
	if Init("config_missing.json", "_token2.json") == nil {
		t.Error("bad load config")
	}
	if Init("config_bad.json", "_token2.json") == nil {
		t.Error("load bad config")
	}
	// substitute request function
	Prompt = func(string) (string, error) {
		return "4/A3r-rEOKf4w8a0Y-26Y9wvzHXq5kl8NsO_x9gaf-OAw", nil
	}
	Init("config.json", "_token2.json")
	// restore initialization
	if err := Init("config.json", "token.json"); err != nil {
		t.Fatal(err)
	}

}
