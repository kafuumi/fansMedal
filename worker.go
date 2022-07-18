package fans

import (
	"context"
	"log"
	"os"
	"sync"
	"time"
)

var (
	logger = log.New(os.Stdout, "[log]", log.LstdFlags)
)

type workerConfig struct {
	likeCD int  //点赞cd
	chatCD int  //发弹幕cd
	isWear bool //是否带上粉丝牌发弹幕
}

type Worker struct {
	medals []Medal //需要处理的粉丝牌任务
	bili   *Bili
	workerConfig
}

func NewWorker(medal []Medal, b *Bili, cfg workerConfig) *Worker {
	return &Worker{
		medals:       medal,
		bili:         b,
		workerConfig: cfg,
	}
}

func CreateWorker(cfg Config) *Worker {
	b := NewBili(cfg.AccessKey, 1)
	err := b.UserInfo()
	if err != nil {
		logError(err, "登录失败")
		return nil
	}
	logger.Printf("用户信息：%s", b.u.uname)
	m, err := b.GetMedals()
	if err != nil {
		logError(err, "获取粉丝牌失败")
		return nil
	}
	lSet := make(Set[int64])
	lSet.Add(cfg.List...)

	medals := make([]Medal, 0)
	for i := range m {
		if (cfg.Type && lSet.Contains(m[i].targetId)) ||
			(!cfg.Type && !lSet.Contains(m[i].targetId)) {
			count := m[i].limit - m[i].todayFeed
			if m[i].level < 20 && count > 0 {
				medals = append(medals, m[i])
				hc := count / 100 * 5
				if count > b.HeartCount {
					b.HeartCount = hc + 1
				}
				logger.Printf("粉丝牌：%s 加入任务队列\n", m[i].name)
			}
		}
	}
	wCfg := workerConfig{
		likeCD: cfg.LikeCD,
		chatCD: cfg.ChatCD,
		isWear: cfg.IsWear,
	}
	return NewWorker(medals, b, wCfg)
}

func logError(err error, msg string, args ...interface{}) {
	if err != nil {
		logger.Printf(msg, args...)
		logger.Printf(" %v\n", err)
	}
}

func do(ctx context.Context, d time.Duration, f func(now time.Time) bool) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if con := f(now); !con {
				return
			}
		}
	}
}

func (w *Worker) DoLike(ctx context.Context) {
	i := 0
	do(ctx, time.Duration(w.likeCD)*time.Second, func(now time.Time) bool {
		var m Medal
		m, i = w.medals[i], i+1
		err := w.bili.Like(m.roomId)
		logError(err, "点赞失败：name=%s", m.name)
		if i >= len(w.medals) {
			return false
		}
		return true
	})
	logger.Println("点赞任务完成")
}

func (w *Worker) DoChat(ctx context.Context) {
	i := 0
	do(ctx, time.Duration(w.chatCD)*time.Second, func(now time.Time) bool {
		var m Medal
		m, i = w.medals[i], i+1
		if w.isWear {
			err := w.bili.WearMedal(m.medalId)
			logError(err, "佩戴粉丝牌：%s 失败", m.name)
		}
		err := w.bili.SendChat(m.roomId, "可爱捏 啵啵")
		logError(err, "发送弹幕失败：name=%s", m.name)
		if i >= len(w.medals) {
			return false
		}
		return true
	})
	logger.Println("弹幕任务完成")
}

func (w *Worker) ShowLive(ctx context.Context) {
	var wg sync.WaitGroup
	for i := range w.medals {
		wg.Add(1)
		go func(m Medal) {
			defer wg.Done()
			err := w.bili.Heartbeat(ctx, m.roomId, m.targetId)
			logError(err, "发送心跳包失败：name=%s", m.name)
		}(w.medals[i])
	}
	wg.Wait()
}

func (w *Worker) DoSign(_ context.Context) {
	hadDays, allDay, err := w.bili.SignIn()
	if err != nil {
		logError(err, "签到失败！")
		return
	}
	if allDay == 0 {
		logger.Printf("重复签到！")
		return
	}
	logger.Printf("签到成功：%d/%d", hadDays, allDay)
}

func (w *Worker) Start(ctx context.Context, exit chan struct{}) {
	if len(w.medals) == 0 {
		logger.Println("无任务，可能是无粉丝牌或者均处理完成")
		exit <- struct{}{}
		return
	}
	works := []func(context.Context){
		w.DoSign, w.DoLike, w.DoChat, w.ShowLive,
	}
	var wg sync.WaitGroup
	wg.Add(len(works))

	for i := range works {
		go func(i int) {
			defer wg.Done()
			works[i](ctx)
		}(i)
	}
	go func() {
		wg.Wait()
		exit <- struct{}{}
	}()
}