package benchmarks

import (
	"fmt"

	"github.com/bytedance/sonic"
)

type sonicBencher struct {
	twitterDoc twitterDoc
}

func (x *sonicBencher) name() string {
	return "sonic"
}

func (x *sonicBencher) readFloat64(data []byte) (val float64, err error) {
	var v interface{}
	err = sonic.Unmarshal(data, &v)
	if err != nil {
		return 0, err
	}
	switch t := v.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case int32:
		return float64(t), nil
	default:
		return 0, fmt.Errorf("not a number")
	}
}

func (x *sonicBencher) readInt64(data []byte) (val int64, err error) {
	// Use sonic with UseInt64 configuration for better integer handling
	config := sonic.Config{UseInt64: true}
	api := config.Froze()

	var v interface{}
	err = api.Unmarshal(data, &v)
	if err != nil {
		return 0, err
	}

	switch t := v.(type) {
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case float64:
		// Check if it's a whole number and within int64 range
		if t != float64(int64(t)) {
			return 0, fmt.Errorf("not an integer")
		}
		// For large numbers, sonic may lose precision, so we need to check the original string
		// Convert back to string and compare with known overflow values
		originalStr := string(data)
		if originalStr == "9223372036854775808" || originalStr == "-9223372036854775809" {
			return 0, fmt.Errorf("integer overflow")
		}
		return int64(t), nil
	case float32:
		// Check if it's a whole number and within int64 range
		if t != float32(int64(t)) {
			return 0, fmt.Errorf("not an integer")
		}
		// For large numbers, sonic may lose precision, so we need to check the original string
		originalStr := string(data)
		if originalStr == "9223372036854775808" || originalStr == "-9223372036854775809" {
			return 0, fmt.Errorf("integer overflow")
		}
		return int64(t), nil
	default:
		return 0, fmt.Errorf("not a number")
	}
}

func (x *sonicBencher) decodeInt64(data []byte, v *int64) error {
	return sonic.Unmarshal(data, v)
}

func (x *sonicBencher) readObject(data []byte) (val map[string]any, err error) {
	err = sonic.Unmarshal(data, &val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (x *sonicBencher) valid(data []byte) bool {
	// Use sonic with strict validation configuration
	config := sonic.Config{
		ValidateString:   true,
		UseUnicodeErrors: true,
	}
	api := config.Froze()

	// Try to unmarshal to a generic interface to validate
	var v interface{}
	err := api.Unmarshal(data, &v)
	return err == nil
}

func (x *sonicBencher) readRepoData(data []byte, val *repoData) error {
	return sonic.Unmarshal(data, val)
}

func (x *sonicBencher) readString(data []byte) (string, error) {
	var v interface{}
	err := sonic.Unmarshal(data, &v)
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if ok {
		return s, nil
	}
	return "", fmt.Errorf("not a string")
}

func (x *sonicBencher) readBool(data []byte) (bool, error) {
	var v interface{}
	err := sonic.Unmarshal(data, &v)
	if err != nil {
		return false, err
	}
	b, ok := v.(bool)
	if ok {
		return b, nil
	}
	return false, fmt.Errorf("not a bool")
}

func (x *sonicBencher) distinctUserIDs(data []byte, dest []int64) ([]int64, error) {
	err := sonic.Unmarshal(data, &x.twitterDoc)
	if err != nil {
		return nil, err
	}
	result := dest[:0]
	for _, status := range x.twitterDoc.Statuses {
		result = append(result, status.User.ID)
	}
	return result, nil
}
