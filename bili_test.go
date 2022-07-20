package fans

import (
	"github.com/golang/mock/gomock"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	accessKey = ""
	bili      *Bili
	isMock    = false
)

//获取用户信息
func TestBili_UserInfo(t *testing.T) {
	var r Request
	if isMock {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		temp := NewMockRequest(ctrl)
		temp.EXPECT().Get(userInfoApi, gomock.Any(), gomock.Any()).DoAndReturn(
			func(args M) (*gjson.Result, error) {
				k := args["access_key"]
				if k == "" {

				}
				return nil, nil
			})
		r = temp
	}
	tests := []struct {
		name  string
		key   string
		uname string
	}{
		{"empty key", "", ""},
		{"invalid key", "160cf4cc3dc0c9f90", ""},
		{"valid key", accessKey, "啵啵"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := NewBili(test.key, 0)
			if isMock {
				b.r = r
			}
			err := b.UserInfo()
			if test.uname != "" {
				assert.NoError(t, err)
				assert.NotEqual(t, "", b.u.uname)
			} else {
				assert.ErrorIs(t, err, ErrInvalidKey)
				assert.Equal(t, "", b.u.uname)
			}
		})
	}
}

func TestMain(m *testing.M) {
	accessKey = os.Getenv("access_key")
	bili = NewBili(accessKey, 10)
	if accessKey == "" {
		log.Println("bili_test: 未设置access_key环境变量，执行mock测试")
		isMock = true
		bili.u.accessKey = "mock"
		bili.r = nil
	}
	os.Exit(m.Run())
}

//直播签到
func TestBili_SignIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//ignore
	_, _, err := bili.SignIn()
	assert.NoError(t, err)
}

//获取拥有的粉丝牌
func TestBili_GetMedals(t *testing.T) {
	medals, err := bili.GetMedals()
	assert.NoError(t, err)
	assert.NotEmpty(t, medals)
	t.Logf("获取到粉丝牌数量：%d", len(medals))
	for i := range medals {
		if medals[i].isWear {
			log.Printf("当前佩戴的粉丝牌：%s", medals[i].name)
			break
		}
	}
}

//佩戴粉丝牌
func TestBili_WearMedal(t *testing.T) {
	tests := []struct {
		name string
		id   int
		ok   bool
	}{
		{"empty id", 0, false},
		{"error id", 1234, false},
		{"not has id", 585997, false},
		{"ok id", 378375, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ok {
				assert.NoError(t, bili.WearMedal(test.id))
			} else {
				err := bili.WearMedal(test.id)
				t.Log(err)
				assert.ErrorContains(t, err, "code:")
			}
		})
		//必须等待一会，否则会提示操作频繁
		time.Sleep(500 * time.Millisecond)
	}
}

//点赞直播间
func TestBili_Like(t *testing.T) {
	tests := []struct {
		name string
		id   int
		ok   bool
	}{
		{"empty id", 0, false},
		{"ok id", 22634198, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ok {
				assert.NoError(t, bili.Like(test.id))
			} else {
				err := bili.Like(test.id)
				t.Log(err)
				assert.ErrorContains(t, err, "code")
			}
		})
	}
}

//分享直播间
func TestBili_Share(t *testing.T) {
	tests := []struct {
		name string
		id   int
		ok   bool
	}{
		{"empty id", 0, false},
		{"ok id", 22634198, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ok {
				assert.NoError(t, bili.Share(test.id))
			} else {
				err := bili.Share(test.id)
				t.Log(err)
				assert.ErrorContains(t, err, "code")
			}
		})
	}
}

//发送弹幕
func TestBili_SendChat(t *testing.T) {
	tests := []struct {
		name string
		id   int
		msg  string
		ok   bool
	}{
		{"empty id", 0, "", false},
		{"ban id", 11365, "早上好", false}, //被禁言的直播间
		{"ok id", 22634198, "早上好", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ok {
				assert.NoError(t, bili.SendChat(test.id, test.msg))
			} else {
				err := bili.SendChat(test.id, test.msg)
				t.Log(err)
				assert.ErrorContains(t, err, "code")
			}
		})
		time.Sleep(time.Second)
	}
}

//心跳包
func TestBili_Heartbeat(t *testing.T) {
	/*ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	err := bili.Heartbeat(ctx, 1200909, 39742326)
	assert.NoError(t, err)*/
	//ignore
	t.Log("ignore test: Bili_Heartbeat")
}
