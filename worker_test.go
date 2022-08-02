package fans

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestWorker_ShowLive(t *testing.T) {
	accessKey = os.Getenv("access_key")
	if accessKey == "" {
		log.Println("bili_test: 未设置access_key环境变量，退出测试")
		return
	}
	bili = NewBili(accessKey, 10)
	_, err := bili.GetMedals()
	assert.NoError(t, err)

}
