package util

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"time"
)

func GetAggregateHash(bucket time.Time, action api.Action, filters ...string) string {
	ret := bucket.Format("2006-01-02T15:04:05") + action.String()
	for i := range filters {
		ret += filters[i]
	}
	h := sha256.New()
	h.Write([]byte(ret))
	return hex.EncodeToString(h.Sum(nil))
}
