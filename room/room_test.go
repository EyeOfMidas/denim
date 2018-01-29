package room

import (
	"os"
	"reflect"
	"testing"

	"github.com/dotariel/denim/bluejeans"
	"github.com/emersion/go-vcard"
)

var wd, _ = os.Getwd()

func TestResolveSource(t *testing.T) {
	tmpDir := setup()

	testCases := []struct {
		description string
		env         map[string]string
		expected    string
	}{
		{
			description: "empty all around",
			env:         map[string]string{"DENIM_ROOMS": "", "DENIM_HOME": "", "HOME": ""},
			expected:    "",
		},
		{
			description: "default to $HOME",
			env:         map[string]string{"DENIM_ROOMS": "", "DENIM_HOME": "", "HOME": tmpDir.UserHome},
			expected:    tmpDir.UserHome + "/.denim/rooms",
		},
		{
			description: "override with $DENIM_HOME",
			env:         map[string]string{"DENIM_ROOMS": "", "DENIM_HOME": tmpDir.AppHome, "HOME": tmpDir.UserHome},
			expected:    tmpDir.AppHome + "/rooms",
		},
		{
			description: "override with $DENIM_ROOMS file",
			env:         map[string]string{"DENIM_ROOMS": tmpDir.AppHome + "/rooms", "DENIM_HOME": tmpDir.AppHome, "HOME": tmpDir.UserHome},
			expected:    tmpDir.AppHome + "/rooms",
		},
		{
			description: "override with $DENIM_ROOMS url",
			env:         map[string]string{"DENIM_ROOMS": "http://localhost:8080/rooms", "DENIM_HOME": tmpDir.AppHome, "HOME": tmpDir.UserHome},
			expected:    "http://localhost:8080/rooms",
		},
	}

	for _, tt := range testCases {

		for k, v := range tt.env {
			os.Setenv(k, v)
		}

		if actual := resolveSource(); actual != tt.expected {
			t.Errorf("'%v' failed; wanted: %v, but got: %v", tt.description, tt.expected, actual)
		}
	}

	teardown(tmpDir)
}

func TestLoad(t *testing.T) {
	tmp := setup()

	testCases := []struct {
		description string
		input       string
		expected    int
	}{
		{description: "bad file", input: "FOO\r\nBAR\r\n", expected: 0},
		{description: "single", input: "ABC 12345\n", expected: 1},
		{description: "extra columns", input: "MORE THAN TWO COLUMNS\n", expected: 1},
		{description: "multiple", input: "ABC 12345\nXYZ 9823", expected: 2},
		{description: "empty lines", input: "\nABC 12345\n\nXYZ 9823", expected: 2},
	}

	for _, tt := range testCases {
		f := touch(tmp.Root + "/rooms") // Create a local file for use
		os.Setenv("DENIM_ROOMS", f.Name())
		f.WriteString(tt.input)

		Load()

		if actual := len(rooms); actual != tt.expected {
			t.Errorf("'%v' failed; wanted: %v, but got: %v", tt.description, tt.expected, actual)
		}
	}

	teardown(tmp)
}

func TestFind(t *testing.T) {
	rooms = []Room{
		{Meeting: bluejeans.New("12345"), Name: "foo"},
		{Meeting: bluejeans.New("67890"), Name: "bar"},
	}

	testCases := []struct {
		input    string
		error    bool
		expected bool
	}{
		{input: "foo", error: false, expected: true},
		{input: "Foo", error: false, expected: true},
		{input: "bar", error: false, expected: true},
		{input: "baz", error: true, expected: false},
	}

	for _, tt := range testCases {
		actual, err := Find(tt.input)

		if (err != nil) != tt.error {
			t.Errorf("expected error mismatch; wanted: %v, but got: %v", tt.error, err != nil)
		}

		if (actual != nil) != tt.expected {
			t.Errorf("failed expectation; wanted: %v, but got: %v", tt.expected, actual)
		}
	}
}

func TestExport(t *testing.T) {
	tmpDir := setup()

	testCases := []struct {
		description string
		input       []Room
		prefix      string
		expected    string
	}{
		{
			description: "single entry without prefix",
			input: []Room{
				{Meeting: bluejeans.New("12345"), Name: "foo_1"},
			},
			prefix:   "",
			expected: wd + "/fixtures/single-noprefix.vcf",
		},
		{
			description: "single entry with prefix",
			input: []Room{
				{Meeting: bluejeans.New("12345"), Name: "foo_1"},
			},
			prefix:   "foo-",
			expected: wd + "/fixtures/single-prefix.vcf",
		},
		{
			description: "multiple entries",
			input: []Room{
				{Meeting: bluejeans.New("12345"), Name: "foo_1"},
				{Meeting: bluejeans.New("12345"), Name: "bar_1"},
			},
			prefix:   "foo-",
			expected: wd + "/fixtures/multiple.vcf",
		},
	}

	for _, tt := range testCases {
		rooms = tt.input
		f, err := Export(tmpDir.Root+"/rooms.vcf", tt.prefix)

		if err != nil {
			panic(err)
		}

		expFile, _ := os.Open(tt.expected)
		defer expFile.Close()

		actFile, _ := os.Open(f.Name())
		defer actFile.Close()

		actual, _ := vcard.NewDecoder(actFile).Decode()
		expected, _ := vcard.NewDecoder(expFile).Decode()

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("'%s' failed; wanted:%v, but got:%v", tt.description, expected, actual)
		}
	}

	teardown(tmpDir)
}

func TestIsURL(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{input: "", expected: false},
		{input: "/foo", expected: false},
		{input: "http://foo.co/bar", expected: true},
		{input: "https://foo.co/bar", expected: true},
	}

	for _, tt := range testCases {
		if actual := isURL(tt.input); actual != tt.expected {
			t.Errorf("'%v' failed; wanted:%v, but got:%v", tt.input, tt.expected, actual)
		}
	}
}
