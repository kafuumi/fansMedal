package fans

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestM_Param(t *testing.T) {
	tests := []struct {
		name string
		body M
		want string
	}{
		{"test case", M{
			"id":   "114514",
			"str":  "1919810",
			"test": "いいよ，こいよ",
		}, "id=114514&str=1919810&test=%E3%81%84%E3%81%84%E3%82%88%EF%BC%8C%E3%81%93%E3%81%84%E3%82%88"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.body.Param())
		})
	}
}

func TestES_Map(t *testing.T) {
	want := M{
		"str":       "test",
		"emptyStr":  "",
		"int":       0,
		"boolTrue":  true,
		"boolFalse": false,
	}
	src := ES{
		{"str", "test"},
		{"emptyStr", ""},
		{"int", 0},
		{"boolTrue", true},
		{"boolFalse", false},
	}
	r := src.Map()
	assert.Equal(t, want, r)
}
