package hashid

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
)

// HashAlgorithm captures the supported hasing algorithms
type HashAlgorithm string

const (
	MD5 HashAlgorithm = "md5"
	// SHA1 is a hashing algorithm that produces a 256-bit digest.
	SHA1 HashAlgorithm = "sha1"
	// SHA256 is a hashing algorithm that produces a 256-bit digest.
	SHA256      HashAlgorithm = "sha256"
	HMAC_SHA256 HashAlgorithm = "hmac"
)

type options struct {
	hashAlgo    HashAlgorithm
	normalize   bool
	normalizer  func(string) (string, error)
	uuidVersion int
	hmacKey     []byte
	charMap     map[string]string
}

// Option configures the behavior of the New function. It allows you to set
// different hashing algorithms, normalization rules, or UUID versions.
type Option func(*options)

func defaultOptions() options {
	return options{
		hashAlgo:    MD5,
		normalize:   true,
		normalizer:  Normalizer,
		uuidVersion: 3,
		hmacKey:     nil,
		charMap:     nil,
	}
}

// WithHashAlgorithm sets the hashing algorithm for generating the UUID.
//
// Supported algorithms: MD5, SHA1, SHA256, HMAC-SHA256.
func WithHashAlgorithm(algo HashAlgorithm) Option {
	return func(o *options) {
		o.hashAlgo = algo
		switch algo {
		case SHA1:
			o.uuidVersion = 5
		case HMAC_SHA256:
			o.uuidVersion = 8
		default:
			o.uuidVersion = 3
		}
	}
}

// WithCustomNormalizer sets the function used to normalize
// the input strings. This step is crucial to ensure that
// different string representations of the same entity
// return the same hash, for example if your input is
// an username or email then you ensure that lower/upper
// and other characters make no difference.
func WithCustomNormalizer(normalizer func(string) (string, error)) Option {
	return func(o *options) {
		o.normalizer = normalizer
	}
}

// WithHMACKey will set the HMAC key used to hash
// the input strings.
func WithHMACKey(key []byte) Option {
	return func(o *options) {
		o.hmacKey = key
		o.hashAlgo = HMAC_SHA256
	}
}

// WithNormalization will set wether we normalize
// input strings or not. Default `true`
func WithNormalization(normalize bool) Option {
	return func(o *options) {
		o.normalize = normalize
	}
}

// WithUUIDVersion allows explicitly setting the UUID version
func WithUUIDVersion(version int) Option {
	return func(o *options) {
		o.uuidVersion = version
	}
}

// WithCustomCharMap allows you to provide a custom character map
// for normalization. The character map replaces specific characters
// in the input string with mapped values, enabling support for
// custom transformations during normalization.
//
// Example usage:
//
//	customMap := map[string]string{"Æ": "AE", "ß": "ss"}
//	id, _ := hashid.New("input", hashid.WithCustomCharMap(customMap))
func WithCustomCharMap(mapping map[string]string) Option {
	return func(o *options) {
		o.charMap = mapping
	}
}

// groupID := fmt.Sprintf("group:%s", message.ID.String())
// gid, err := hashid.NewUUID(groupID)
// func NewUUIDSptf(input, )

func NewUUID(input string, opts ...Option) (uuid.UUID, error) {
	id, err := New(input, opts...)
	if err != nil {
		return uuid.Nil, err
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}

func NewShortID(input string, opts ...Option) (string, error) {
	uid, err := NewUUID(input, opts...)
	if err != nil {
		return "", err
	}
	enc := shortuuid.DefaultEncoder.Encode(uid)
	return enc, nil
}

func ParseShortID(sid string) (uuid.UUID, error) {
	uid, err := shortuuid.DefaultEncoder.Decode(sid)
	if err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}

// New generates a UUID from the provided input string,
// as long as the normalization and hashing options remain
// the same so will the ouptut.
// The algorithm used is defined by the options passed.
//
// Example:
//
//	id, err := hashid.New("user@example.com",
//		   hashid.WithHashAlgorithm(hashid.SHA256))
//	if err != nil {
//	  log.Fatal(err)
//	}
//	fmt.Println(id)
func New(input string, opts ...Option) (string, error) {
	config := defaultOptions()

	for _, opt := range opts {
		opt(&config)
	}

	if config.hashAlgo == HMAC_SHA256 && config.hmacKey == nil {
		return "", fmt.Errorf("HMAC key is required when using HMAC_SHA256")
	}

	uuidVersion := 0
	switch config.uuidVersion {
	case 3, 5, 8:
		uuidVersion = config.uuidVersion
	case 0:
		uuidVersion = 3
	default:
		return "", fmt.Errorf("UUID version should be one of 3, 5, 8")
	}

	var normalizer func(string) (string, error)
	if config.charMap != nil {
		normalizer = func(input string) (string, error) {
			return NormalizerWithCharMap(input, config.charMap)
		}
	} else {
		normalizer = config.normalizer
	}

	var err error

	if config.normalize {
		input, err = normalizer(input)
	}

	if err != nil {
		return "", fmt.Errorf("normalization error: %w", err)
	}

	hasher, err := getHasher(config.hashAlgo, config.hmacKey)
	if err != nil {
		return "", err
	}
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)

	return formatUUID(hash, uuidVersion), nil
}

func getHasher(algo HashAlgorithm, key []byte) (hash.Hash, error) {
	switch algo {
	case SHA1:
		return sha1.New(), nil
	case SHA256:
		return sha256.New(), nil
	case HMAC_SHA256:
		if key == nil {
			return nil, fmt.Errorf("HMAC key is required when using HMAC_SHA256")
		}
		return hmac.New(sha256.New, key), nil
	default:
		return md5.New(), nil
	}
}

// Valid UUID version 3, following the format
// xxxxxxxx-xxxx-3xxx-yxxx-xxxxxxxxxxxx
// Where:
// - The 13th character is always 3 (indicating a version 3 UUID).
// - The 17th character is one of 8, 9, a, or b.
func formatUUID(hash []byte, version int) string {
	var versionByte uint8
	if version == 8 {
		// For version 8, we set the version but don't modify any other bits
		// as per RFC 9562, allowing for custom formats
		versionByte = hash[6]&0x0F | 0x80
	} else {
		// For other versions (3, 5), use traditional version formatting
		versionByte = hash[6]&0x0F | uint8(version<<4)
	}

	// Set the variant to RFC 4122 (the 2 most significant bits should be 10)
	variantByte := hash[8]&0x3F | 0x80

	// Return the UUID formatted string
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		hash[0:4],
		hash[4:6],
		[]byte{versionByte, hash[7]},
		[]byte{variantByte, hash[9]},
		hash[10:16],
	)
}
