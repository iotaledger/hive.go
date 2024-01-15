package serix

import (
	"strconv"
	"strings"

	"github.com/iotaledger/hive.go/ierrors"
)

func parseStructTagValue(name string, keyValue []string, currentPart string) (string, error) {
	if len(keyValue) != 2 {
		return "", ierrors.Errorf("incorrect %s tag format: %s", name, currentPart)
	}

	return keyValue[1], nil
}

func parseStructTagValueUint(name string, keyValue []string, currentPart string) (uint, error) {
	value, err := parseStructTagValue(name, keyValue, currentPart)
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, ierrors.Wrapf(err, "failed to parse %s %s", name, currentPart)
	}

	return uint(result), nil
}

func parseLengthPrefixType(prefixTypeRaw string) (LengthPrefixType, error) {
	switch prefixTypeRaw {
	case "byte", "uint8":
		return LengthPrefixTypeAsByte, nil
	case "uint16":
		return LengthPrefixTypeAsUint16, nil
	case "uint32":
		return LengthPrefixTypeAsUint32, nil
	case "uint64":
		return LengthPrefixTypeAsUint64, nil
	default:
		return LengthPrefixTypeAsByte, ierrors.Wrapf(ErrUnknownLengthPrefixType, "%s", prefixTypeRaw)
	}
}

func parseStructTagValuePrefixType(name string, keyValue []string, currentPart string) (LengthPrefixType, error) {
	value, err := parseStructTagValue(name, keyValue, currentPart)
	if err != nil {
		return 0, err
	}

	lengthPrefixType, err := parseLengthPrefixType(value)
	if err != nil {
		return 0, ierrors.Wrapf(err, "failed to parse %s %s", name, currentPart)
	}

	return lengthPrefixType, nil
}

type TagSettings struct {
	position   int
	isOptional bool
	inlined    bool
	omitEmpty  bool
	ts         TypeSettings
}

func (ts TagSettings) Position() int {
	return ts.position
}

func (ts TagSettings) IsOptional() bool {
	return ts.isOptional
}

func (ts TagSettings) Inlined() bool {
	return ts.inlined
}

func (ts TagSettings) OmitEmpty() bool {
	return ts.omitEmpty
}

func (ts TagSettings) TypeSettings() TypeSettings {
	return ts.ts
}

// ParseSerixSettings parses the given struct tag and returns the settings.
func ParseSerixSettings(tag string, serixPosition int) (TagSettings, error) {
	settings := TagSettings{}
	settings.position = serixPosition

	if tag == "" {
		// empty struct tags are allowed
		return settings, nil
	}

	parts := strings.Split(tag, ",")
	keyPart := parts[0]

	if strings.ContainsAny(keyPart, "=") {
		return TagSettings{}, ierrors.Errorf("incorrect struct tag format: %s, must start with the field key or \",\"", tag)
	}

	if keyPart != "" {
		settings.ts = settings.ts.WithFieldKey(keyPart)
	}

	parts = parts[1:]
	seenParts := map[string]struct{}{}
	for _, currentPart := range parts {
		if _, ok := seenParts[currentPart]; ok {
			return TagSettings{}, ierrors.Errorf("duplicated tag part: %s", currentPart)
		}
		keyValue := strings.Split(currentPart, "=")
		partName := keyValue[0]

		switch partName {
		case "optional":
			settings.isOptional = true

		case "inlined":
			settings.inlined = true

		case "omitempty":
			settings.omitEmpty = true

		case "description":
			value, err := parseStructTagValue("description", keyValue, currentPart)
			if err != nil {
				return TagSettings{}, err
			}
			settings.ts = settings.ts.WithDescription(value)

		case "maxByteSize":
			value, err := parseStructTagValueUint("maxByteSize", keyValue, currentPart)
			if err != nil {
				return TagSettings{}, err
			}
			settings.ts = settings.ts.WithMaxByteSize(value)

		case "lenPrefix":
			value, err := parseStructTagValuePrefixType("lenPrefix", keyValue, currentPart)
			if err != nil {
				return TagSettings{}, err
			}
			settings.ts = settings.ts.WithLengthPrefixType(value)

		case "minLen":
			value, err := parseStructTagValueUint("minLen", keyValue, currentPart)
			if err != nil {
				return TagSettings{}, err
			}
			settings.ts = settings.ts.WithMinLen(value)

		case "maxLen":
			value, err := parseStructTagValueUint("maxLen", keyValue, currentPart)
			if err != nil {
				return TagSettings{}, err
			}
			settings.ts = settings.ts.WithMaxLen(value)

		default:
			return TagSettings{}, ierrors.Errorf("unknown tag part: %s", currentPart)
		}

		seenParts[partName] = struct{}{}
	}

	return settings, nil
}
