package fans

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name string
		body M
		sign string
	}{
		{"test case", M{
			"id":   "114514",
			"str":  "1919810",
			"test": "いいよ，こいよ",
		}, "b74d668b66fe3f96053a37331dc15fc1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sign(test.body)
			b := test.body
			assert.Contains(t, b, "appkey")
			assert.Equal(t, b["sign"], test.sign)
		})
	}
}

func TestClientSign(t *testing.T) {
	body := ES{
		{"platform", "android"},
		{"uuid", ""},
		{"buvid", "ddd"},
		{"seq_id", 1},
		{"room_id", 1},
		{"parent_id", 6},
		{"area_id", 283},
		{"timestamp", 1657948446 - 60},
		{"secret_key", "axoaadsffcazxksectbbb"},
		{"watch_time", 60},
		{"up_id", 1},
		{"up_level", 40},
		{"jump_from", 30000},
		{"gu_id", "ddd"},
		{"play_type", 0},
		{"play_url", ""},
		{"s_time", 0},
		{"data_behavior_id", ""},
		{"data_source_id", ""},
		{"up_session", fmt.Sprintf("l:one:live:record:%d:%d", 1, 1657948446-88888)},
		{"visit_id", "ddd"},
		{"watch_status", "%7B%22pk_id%22%3A0%2C%22screen_status%22%3A1%7D"},
		{"click_id", ""},
		{"session_id", ""},
		{"player_type", 0},
		{"client_ts", 1657948446},
	}
	r := "e7efea46d6cd1a94097e9a55ee98d8cc3d8a1dc492f40cf8e1d1c52c82f25cf67a1f67bb4840ceb1d794b547587ce51ef1d499bfa8cb0f76f0549077c6e09775"
	assert.Equal(t, r, clientSign(body))
}

func TestRandStr(t *testing.T) {
	tests := []struct {
		name string
		l    int
		mode int
	}{
		{"empty str", 0, mixedRandStr},
		{"mixed str", 10, mixedRandStr},
		{"lower str", 10, lowerRandStr},
		{"upper str", 10, upperRandStr},
		{"gather than table len 1", 72, mixedRandStr},
		{"gather than table len 2", 100, mixedRandStr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := randStr(test.l, test.mode)
			assert.Equal(t, test.l, len(s))
			if test.mode > 0 {
				assert.Equal(t, strings.ToUpper(s), s)
			} else if test.mode < 0 {
				assert.Equal(t, strings.ToLower(s), s)
			}
		})
	}
}
