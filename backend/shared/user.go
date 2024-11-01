package shared

import (
	"strings"

	"github.com/btcsuite/btcutil/base58"

	"go.uber.org/zap"
)

const PROVIDER_SUBJECT_SEPARATOR = "]:=:["

func EncodeUserGitName(logger *zap.Logger, provider, subject string) string {
	encoded := base58.Encode([]byte(provider + PROVIDER_SUBJECT_SEPARATOR + subject))
	return encoded
}

func DecodeUserGitName(logger *zap.Logger, encoded string) (string, string) {
	decoded := string(base58.Decode(encoded))
	parts := strings.Split(string(decoded), PROVIDER_SUBJECT_SEPARATOR)
	if len(parts) != 2 {
		logger.Error("Failed to decode user git name")
		return "", ""
	}
	return parts[0], parts[1]
}
