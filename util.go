package fans

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

func tsStr() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func sign(body M) {
	v := make(url.Values)
	body["appkey"] = appKey
	for k := range body {
		v.Add(k, fmt.Sprint(body[k]))
	}

	h := md5.New()
	// Encode 方法会对键进行排序
	h.Write([]byte(v.Encode()))
	h.Write([]byte(appSec))
	body["sign"] = hex.EncodeToString(h.Sum(nil))
}

func handResp(resp *http.Response, err error) (*gjson.Result, error) {
	if err != nil {
		return nil, err
	}
	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		return nil, errors.Errorf("请求失败：code=%d, status=%s", statusCode, resp.Status)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var reader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "错误的gzip压缩格式")
		}
	case "deflate":
		reader = flate.NewReader(resp.Body)
	case "br":
		reader = brotli.NewReader(resp.Body)
	default:
		reader = resp.Body
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, errors.Wrap(err, "读取数据失败")
	}
	r := gjson.ParseBytes(buf.Bytes())
	return &r, nil
}

func checkResp(resp *gjson.Result) (*gjson.Result, error) {
	code := resp.Get("code").Int()
	if code != 0 {
		message := resp.Get("message")
		if message.Exists() {
		} else {
			message = resp.Get("msg")
		}
		return nil, errors.Errorf("code:%d, msg:%s", code, message.String())
	}
	data := resp.Get("data")
	return &data, nil
}

func clientSign(body ES) string {
	sb := new(strings.Builder)
	sb.WriteByte('{')
	for i := range body {
		_, _ = fmt.Fprintf(sb, `"%s":"%v"`, body[i].name, body[i].value)
		if i < len(body)-1 {
			sb.WriteByte(',')
		}
	}
	sb.WriteByte('}')
	src := sb.String()
	blakeHash, _ := blake2b.New512(nil)
	hashes := []hash.Hash{sha512.New(), sha3.New512(), sha512.New384(), sha3.New384(), blakeHash}
	data := []byte(src)
	for i := range hashes {
		hashes[i].Write(data)
		temp := hashes[i].Sum(nil)
		data = make([]byte, hex.EncodedLen(len(temp)))
		hex.Encode(data, temp)
	}
	return string(data)
}

const (
	mixedRandStr = 0
	upperRandStr = 1
	lowerRandStr = -1
)

//随机生成长度为 l 的字符串，mode用于指定大小写，等于0为大小写混合，大于0为全大小，小于0为全小写
func randStr(l int, mode int) string {
	var table []byte
	const (
		lower = "abcdefghijklmnopqrstuvwxyz"
		upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digit = "1234567890"
	)
	if mode == 0 {
		table = []byte(lower + upper + digit)
	} else if mode > 0 {
		table = []byte(upper + digit)
	} else {
		table = []byte(lower + digit)
	}

	n := len(table)
	rand.Seed(time.Now().UnixNano())
	//打乱table的顺序
	rand.Shuffle(n, func(i, j int) {
		table[i], table[j] = table[j], table[i]
	})
	if l <= n {
		return string(table[:l])
	} else {
		q, r := l/n, l%n
		temp := make([]byte, 0)
		for ; q != 0; q-- {
			temp = append(temp, table...)
		}
		return string(append(temp, table[:r]...))
	}
}
