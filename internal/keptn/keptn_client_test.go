package keptn

import "testing"

// Test that unsupported metrics return an error
func TestGetUnsupportedSLI(t *testing.T) {
	customQueries := NewEmptyCustomQueries()
	got, err := customQueries.GetQueryByNameOrDefault("foobar")
	if got != "" {
		t.Errorf("dh.getTimeseriesConfig() returned (\"%s\"), expected(\"\")", got)
	}

	// TODO 2021-09-02: create own error and check type & property name below instead of error message
	expected := "unsupported SLI foobar"

	if err == nil {
		t.Errorf("dh.getTimeseriesConfig() did not return an error")
	} else {
		if err.Error() != expected {
			t.Errorf("dh.getTimeseriesConfig() returned error %s, expected %s", err.Error(), expected)
		}
	}
}
