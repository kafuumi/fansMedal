package fans

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	appKey = "4409e2ce8ffd12b8"
	appSec = "59b43e04ad6965f34319062b478f83dd"
)
const (
	retryTimes = 3
)

const (
	userInfoApi  = "https://app.bilibili.com/x/v2/account/mine"
	signInApi    = "https://api.live.bilibili.com/rc/v1/Sign/doSign"
	getMedalsApi = "https://api.live.bilibili.com/xlive/app-ucenter/v1/fansMedal/panel"
	wearMedals   = "https://api.live.bilibili.com/xlive/app-ucenter/v1/fansMedal/wear"
	likeApi      = "https://api.live.bilibili.com/xlive/web-ucenter/v1/interact/likeInteract"
	shareApi     = "https://api.live.bilibili.com/xlive/app-room/v1/index/TrigerInteract"
	sendChatApi  = "https://api.live.bilibili.com/xlive/app-room/v1/dM/sendmsg"
	heartbeatApi = "https://live-trace.bilibili.com/xlive/data-interface/v1/heartbeat/mobileHeartBeat"
)

var (
	ErrInvalidKey = errors.New("无效的access key")
)

type User struct {
	accessKey string
	uname     string
	uid       int64
}

type Bili struct {
	r          Request
	u          User
	HeartCount int //发送心跳包的最大次数
}

// Medal 粉丝牌
type Medal struct {
	medalId    int    //粉丝牌id
	level      int    //粉丝牌等级
	todayFeed  int    //今日亲密度
	limit      int    //单日最大亲密度
	roomId     int    //房间号
	targetId   int64  //主播uid
	anchorName string //主播昵称
	name       string //粉丝牌名称
	isWear     bool   //当前是否佩戴该粉丝牌
}

func (m Medal) Level() int {
	return m.level
}

func (m Medal) TargetId() int64 {
	return m.targetId
}

func (m Medal) AnchorName() string {
	return m.anchorName
}

func (m Medal) Name() string {
	return m.name
}

func NewBili(key string, count int) *Bili {
	return &Bili{
		u:          User{accessKey: key},
		r:          NewR(3),
		HeartCount: count,
	}
}

func (b *Bili) helpResp(resp *gjson.Result, err error) (*gjson.Result, error) {
	if err != nil {
		return nil, err
	}
	resp, err = checkResp(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (b *Bili) helpParam(param M) M {
	if param == nil {
		param = make(M)
	}
	param["access_key"] = b.u.accessKey
	param["actionKey"] = "appkey"
	param["ts"] = tsStr()
	sign(param)
	return param
}

// UserInfo 通过获取用户信息验证来是否是有效的access_key
func (b *Bili) UserInfo() error {
	resp, err := b.helpResp(b.r.Get(userInfoApi, b.helpParam(nil), nil))
	if err != nil {
		return err
	}
	uid, uname := resp.Get("mid").Int(), resp.Get("name").String()
	if uid == 0 {
		return ErrInvalidKey
	}
	b.u.uname = uname
	b.u.uid = uid
	return nil
}

// SignIn 直播签到
func (b *Bili) SignIn() (hadDays int, allDays int, err error) {
	resp, err := b.r.Get(signInApi, b.helpParam(nil), nil)
	if err != nil {
		return
	}
	code := resp.Get("code").Int()
	//101140 为重复签到
	if code != 0 && code != 1011040 {
		_, err = b.helpResp(resp, err)
		err = errors.Wrap(err, "签到失败")
		return
	}
	//只有签到成功才会有相应的数据返回
	if code == 0 {
		hadDays = int(resp.Get("data.hadSignDays").Int())
		allDays = int(resp.Get("data.allDays").Int())
	}
	return
}

func (b *Bili) GetMedals() ([]Medal, error) {
	param := M{
		"page":      0,
		"page_size": 50,
	}
	page := 0
	//集合去重
	set := make(Set[int])
	medals := make([]Medal, 0)
	parseMedal := func(src *gjson.Result) {
		medalInfo := src.Get("medal")
		anchorInfo := src.Get("anchor_info")
		roomInfo := src.Get("room_info")
		var medal Medal
		medal.medalId = int(medalInfo.Get("medal_id").Int())
		if set.Contains(medal.medalId) {
			return
		}
		medal.level = int(medalInfo.Get("level").Int())
		medal.todayFeed = int(medalInfo.Get("today_feed").Int())
		medal.limit = int(medalInfo.Get("day_limit").Int())

		medal.roomId = int(roomInfo.Get("room_id").Int())
		medal.targetId = medalInfo.Get("target_id").Int()
		medal.anchorName = anchorInfo.Get("nick_name").String()
		medal.name = medalInfo.Get("medal_name").String()
		medal.isWear = medalInfo.Get("wearing_status").Int() == 1

		set.Add(medal.medalId)
		medals = append(medals, medal)
	}
	var err error
	var resp *gjson.Result
	for {
		param = b.helpParam(param)
		resp, err = b.helpResp(b.r.Get(getMedalsApi, param, nil))
		if err != nil {
			break
		}
		list := resp.Get("list").Array()
		specialList := resp.Get("special_list").Array()
		if len(list) == 0 && len(specialList) == 0 {
			break
		}
		for i := range list {
			parseMedal(&(list[i]))
		}
		for i := range specialList {
			parseMedal(&(specialList[i]))
		}
		page++
		param["page"] = page
		delete(param, "sign")
	}
	return medals, err
}

// WearMedal 佩戴指定粉丝牌，id为粉丝牌id
func (b *Bili) WearMedal(id int) error {
	param := b.helpParam(M{
		"medal_id": id,
		"platform": "android",
		"type":     1,
		"version":  0,
	})
	_, err := b.helpResp(b.r.Post(wearMedals, param, nil))
	return err
}

// Like 点赞指定的直播间，id为直播间的真实房间号，
func (b *Bili) Like(id int) error {
	param := b.helpParam(M{
		"roomid": id,
	})
	_, err := b.helpResp(b.r.Post(likeApi, param, nil))
	return err
}

// Share 分享直播间，新规则里，分享直播间不加亲密度了
func (b *Bili) Share(id int) error {
	param := b.helpParam(M{
		"interact_type": 3,
		"roomid":        id,
	})
	_, err := b.helpResp(b.r.Post(shareApi, param, nil))
	return err
}

// SendChat 发送弹幕
func (b *Bili) SendChat(id int, msg string) error {
	param := b.helpParam(nil)
	body := url.Values{}
	body.Add("cid", strconv.Itoa(id))
	body.Add("msg", msg)
	body.Add("rnd", tsStr())
	body.Add("color", "16777215")
	body.Add("fontsize", "25")
	_, err := b.helpResp(b.r.Post(sendChatApi, param, strings.NewReader(body.Encode()),
		E{"Content-Type", applicationForm}))
	return err
}

func (b *Bili) heart(id int, uid int64, now time.Time, uuids [2]string) error {
	param := ES{
		{"platform", "android"},
		{"uuid", uuids[0]},
		{"buvid", randStr(37, upperRandStr)},
		{"seq_id", 1},
		{"room_id", id},
		{"parent_id", 6}, //TODO 不使用固定值
		{"area_id", 283},
		{"timestamp", now.Unix() - 60},
		{"secret_key", "axoaadsffcazxksectbbb"},
		{"watch_time", 60},
		{"up_id", uid},
		{"up_level", 40},
		{"jump_from", "30000"},
		{"gu_id", randStr(43, upperRandStr)},
		{"play_type", 0},
		{"play_url", ""},
		{"s_time", 0},
		{"data_behavior_id", ""},
		{"data_source_id", ""},
		{"up_session", fmt.Sprintf("l:one:live:record:%d:%d", id, now.Unix()-88888)},
		{"visit_id", randStr(32, upperRandStr)},
		{"watch_status", "%7B%22pk_id%22%3A0%2C%22screen_status%22%3A1%7D"},
		{"click_id", uuids[1]},
		{"session_id", ""},
		{"player_type", 0},
		{"client_ts", now.Unix()},
	}
	signStr := clientSign(param)
	body := param.Map()
	body["client_sign"] = signStr
	body = b.helpParam(body)
	_, err := b.helpResp(b.r.Post(heartbeatApi, body, nil))
	return err
}

// Heartbeat 发送心跳包
func (b *Bili) Heartbeat(ctx context.Context, id int, uid int64) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	uuids := [2]string{uuid.NewString(), uuid.NewString()}
	if err := b.heart(id, uid, time.Now(), uuids); err != nil {
		return err
	}
	count := 1
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			var err error
			for retry := retryTimes; retry != 0; retry-- {
				if err = b.heart(id, uid, now, uuids); err == nil {
					break
				}
			}
			if err != nil {
				return err
			}
			count++
			if count >= b.HeartCount {
				return err
			}
		}
	}

}
