package fans

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestWorker_ShowLive(t *testing.T) {
	accessKey = os.Getenv("access_key")
	if accessKey == "" {
		log.Println("bili_test: 未设置access_key环境变量，退出测试")
		return
	}
	bili = NewBili(accessKey, 10)
	ms, err := bili.GetMedals()
	assert.NoError(t, err)
	w := NewWorker(ms, bili, workerConfig{})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	w.ShowLive(ctx)
}
