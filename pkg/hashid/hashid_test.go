package hashid

import (
	"regexp"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNormalizer(t *testing.T) {
	testCases := map[string]string{
		"A81758F#   FFE04©E4F5":                                       "a81758f-ffe04ce4f5",
		"A81758FFFE04©E4F5":                                           "a81758fffe04ce4f5",
		"iot.devicetype:s2node":                                       "iotdevicetypes2node",
		"iot.devicetype:s2--node":                                     "iotdevicetypes2-node",
		"1000円=711.56₹":                                               "1000yen=71156indian-rupee",
		"   UPPER   case   ":                                          "upper-case",
		"iot.devicetype:s2-m1-3200":                                   "iotdevicetypes2-m1-3200",
		"1000円711.56₹":                                                "1000yen71156indian-rupee",
		"decentlab-serial/01588":                                      "decentlab-serial/01588",
		"iot.devicetype:unknown-reader-type":                          "iotdevicetypeunknown-reader-type",
		"iot.devicetype:decentlab-dllp8p-co2-sensor":                  "iotdevicetypedecentlab-dllp8p-co2-sensor",
		"IOT.devicetype:decentlab-dl-lp8p-001-US915-co2-sensor":       "iotdevicetypedecentlab-dl-lp8p-001-us915-co2-sensor",
		"4064441d-7ef2-733e-ddcb-003f7965fa07#eui48#A81758FFFE04E4F8": "4064441d-7ef2-733e-ddcb-003f7965fa07eui48a81758fffe04e4f8",
	}

	for input, expected := range testCases {
		output, err := Normalizer(input)
		assert.NoError(t, err)
		assert.Equal(t, expected, output, "Input: %s", input)
	}
}

func TestNormalizerWithSeparator(t *testing.T) {
	testCases := map[string]struct {
		input     string
		separator string
		expected  string
	}{
		"default separator": {
			input:     "A81758F   FFE04©E4F5",
			separator: "-",
			expected:  "a81758f-ffe04ce4f5",
		},
		"character replacement": {
			input:     "1000円=711.56₹",
			separator: "_",
			expected:  "1000yen=71156indian_rupee",
		},
		"underscore separator": {
			input:     "1000円= 711.56₹",
			separator: "_",
			expected:  "1000yen=_71156indian_rupee",
		},
		"empty separator": {
			input:     "   UPPER   case   ",
			separator: "",
			expected:  "upper-case",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			output, err := NormalizerWithSeparator(tc.input, tc.separator)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestNew(t *testing.T) {
	uuidRegex := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[3458][a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)

	testCases := []struct {
		name    string
		input   string
		options []Option
	}{
		{"Default MD5", "example", nil},
		{"SHA1", "EXAMPLE", []Option{WithHashAlgorithm(SHA1)}},
		{"HMAC-SHA256", "Another String", []Option{WithHMACKey([]byte("secret")), WithHashAlgorithm(HMAC_SHA256)}},
		{"HMAC-SHA256", "user@example.com", []Option{WithHMACKey([]byte("hmac-secret")), WithHashAlgorithm(HMAC_SHA256)}},
		{"No normalization", "     leading and trailing    ", []Option{WithNormalization(false)}},
		{"Custom UUID version", "A81758FFFE04E4F5", []Option{WithUUIDVersion(5)}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uuid, err := New(tc.input, tc.options...)
			assert.NoError(t, err)
			assert.True(t, uuidRegex.MatchString(uuid), "Generated UUID %s does not match expected format", uuid)
		})
	}
}

func TestNewOutput(t *testing.T) {
	testStrings := map[string]string{
		"A81758FFFE04E4F5":                                            "ddea575a-d5e2-3114-9267-dbead79c4ab8",
		"a81758fffe04e4f5":                                            "ddea575a-d5e2-3114-9267-dbead79c4ab8",
		"A81758FFFE04©E4F5":                                           "c4c2a132-cb6f-3e0e-b646-3c04195e72e3",
		"A81758FFFE04(c)E4F5":                                         "c4c2a132-cb6f-3e0e-b646-3c04195e72e3",
		"iot.devicetype:s2node":                                       "712a111e-2d38-31ef-8944-be0caafcc408",
		"decentlab-serial/01588":                                      "f851a55d-8445-3ffb-a367-e7ecc984ca10",
		"iot.devicetype:s2-node":                                      "4a3a3a09-df30-3023-bd6b-933bbce48602",
		"iot.devicetype:s2-m1-3200":                                   "155255f6-35ee-32b9-8fd6-b6ba8c8ee7d9",
		"1000円711.56₹":                                                "f1c804d0-42fe-3a82-a4ac-53aed7fec8d0",
		"unknown-managed-switch-type":                                 "712b3748-6d58-354c-be5f-540c5e6b9fe4",
		"iot.devicetype:unknown-reader-type":                          "ce670956-f57a-3bcd-9201-8e46ff070e10",
		"iot.devicetype:decentlab-dllp8p-co2-sensor":                  "8b57238a-5aa3-315a-a7a7-72effc3e3629",
		"IOT.devicetype:decentlab-dl-lp8p-001-US915-co2-sensor":       "05cd6ee6-626f-3fbe-90c5-fcb6557eabf6",
		"4064441d-7ef2-733e-ddcb-003f7965fa07#eui48#A81758FFFE04E4F8": "2543ca04-1c3a-3cee-9fff-a8cccbb090c9",
	}
	for key, val := range testStrings {
		out, err := New(key)
		assert.NoError(t, err)
		assert.Equal(t, val, out)
	}
}

func TestNewWithCustomCharMap(t *testing.T) {
	customCharMap := map[string]string{
		"@": "at",
		"#": "hash",
	}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Custom @ replacement", "test@example.com", "93dc5e95-bead-3d28-ba29-09aa4855264e"},
		{"Custom # replacement", "test#example", "dd7d5ad9-4714-36c1-ba08-305c07546bce"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uuid, err := New(tc.input, WithCustomCharMap(customCharMap))
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, uuid)
		})
	}
}

func TestNormalizerWithCustomCharMap(t *testing.T) {
	customCharMap := map[string]string{
		"@": "at",
		"#": "hash",
	}

	testCases := []struct {
		input    string
		expected string
	}{
		{"test@example.com", "testatexamplecom"},
		{"test#example", "testhashexample"},
	}

	for _, c := range testCases {
		output, err := NormalizerWithCharMap(c.input, customCharMap)
		assert.NoError(t, err)
		assert.Equal(t, c.expected, output, "Input: %s", c.input)
	}
}

func TestConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	concurrent := 100

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := New("test")
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestInvalidConfigurations(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		options []Option
	}{
		{"HMAC without key", "test", []Option{WithHashAlgorithm(HMAC_SHA256)}},
		{"Invalid UUID version", "test", []Option{WithUUIDVersion(10)}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.input, tc.options...)
			assert.Error(t, err)
		})
	}
}

func TestNewUUID(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		options []Option
		wantErr bool
	}{
		{
			name:    "Default settings",
			input:   "test@example.com",
			options: nil,
			wantErr: false,
		},
		{
			name:    "With SHA1",
			input:   "test@example.com",
			options: []Option{WithHashAlgorithm(SHA1)},
			wantErr: false,
		},
		{
			name:    "With HMAC-SHA256",
			input:   "test@example.com",
			options: []Option{WithHMACKey([]byte("secret")), WithHashAlgorithm(HMAC_SHA256)},
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   "",
			options: nil,
			wantErr: false,
		},
		{
			name:    "Invalid configuration",
			input:   "test",
			options: []Option{WithHashAlgorithm(HMAC_SHA256)}, // HMAC without key
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uid, err := NewUUID(tc.input, tc.options...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, uid)

			// Verify the version based on algorithm
			version := uid.Version()
			if len(tc.options) > 0 {
				opts := defaultOptions()
				tc.options[0](&opts)
				switch opts.hashAlgo {
				case SHA1:
					assert.Equal(t, uuid.Version(5), version)
				case HMAC_SHA256:
					assert.Equal(t, uuid.Version(8), version)
				default:
					assert.Equal(t, uuid.Version(3), version)
				}
			} else {
				assert.Equal(t, uuid.Version(3), version)
			}
		})
	}
}

func TestNewShortUUID(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		options []Option
		wantErr bool
	}{
		{
			name:    "Default settings",
			input:   "test@example.com",
			options: nil,
			wantErr: false,
		},
		{
			name:    "With SHA1",
			input:   "test@example.com",
			options: []Option{WithHashAlgorithm(SHA1)},
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   "",
			options: nil,
			wantErr: false,
		},
		{
			name:    "Invalid configuration",
			input:   "test",
			options: []Option{WithHashAlgorithm(HMAC_SHA256)}, // HMAC without key
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shortID, err := NewShortID(tc.input, tc.options...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, shortID)

			// Verify we can parse it back
			uid, err := ParseShortID(shortID)
			assert.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, uid)
		})
	}
}

func TestParseShortUUID(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid short UUID",
			input:   "2M9JFnemic9bLiXnT8AHun",
			wantErr: false,
		},
		{
			name:    "Invalid characters",
			input:   "invalid-uuid-format!!!",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uid, err := ParseShortID(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, uid)
		})
	}
}

func TestUUIDRoundTrip(t *testing.T) {
	testInputs := []string{
		"test@example.com",
		"user-123",
		"complex!!@@##string",
		"1234567890",
	}

	for _, input := range testInputs {
		t.Run("Round trip: "+input, func(t *testing.T) {
			// Generate short UUID
			shortID, err := NewShortID(input)
			assert.NoError(t, err)

			// Parse it back to UUID
			uid, err := ParseShortID(shortID)
			assert.NoError(t, err)

			// Verify it's valid
			assert.NotEqual(t, uuid.Nil, uid)
		})
	}
}
