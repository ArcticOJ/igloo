package cache

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ArcticOJ/igloo/v0/config"
	"github.com/pierrec/lz4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	defaultKey = "igloo::case_data[%s]"
	hashKey    = "%d|%s"
)

var (
	c *redis.Client
	// group for fetching data
	g singleflight.Group
	// group for adding data to cache
	g2 singleflight.Group
)

func init() {
	c = redis.NewClient(&redis.Options{
		Addr: "192.168.31.212:6380",
	})
}

func Get(ctx context.Context, id string, num uint16, t string) (string, error) {
	f, e, _ := g.Do(fmt.Sprintf("%s/%d|%s", id, num, t), func() (interface{}, error) {
		_, e, _ := g2.Do(id, func() (interface{}, error) {
			r, _e := c.Exists(ctx, fmt.Sprintf(defaultKey, id)).Result()
			if r == 1 && _e == nil {
				return nil, nil
			}
			return nil, load(ctx, id)
		})
		if e != nil {
			return "", e
		}
		return c.HGet(ctx, fmt.Sprintf(defaultKey, id), fmt.Sprintf(hashKey, num, t)).Result()
	})
	return f.(string), e
}

func parse(p string) (id string, num uint16, t string, e error) {
	parts := strings.Split(p, "/")
	if len(parts) != 3 {
		e = errors.New("invalid path")
		return
	}
	id = strings.ToLower(parts[0])
	if n, _e := strconv.ParseUint(parts[1], 10, 16); _e == nil {
		num = uint16(n)
	} else {
		e = _e
		return
	}
	t = strings.ToLower(strings.TrimLeft(path.Ext(parts[2]), "."))
	if !(t == "inp" || t == "out") {
		e = errors.New("invalid type")
		return
	}
	return
}

func load(ctx context.Context, id string) error {
	f, e := os.Open(path.Join(config.Config.Storage.Problems, fmt.Sprintf("%s.tar.lz4", id)))
	if e != nil {
		return e
	}
	defer f.Close()
	t := tar.NewReader(lz4.NewReader(f))
	for h, e := t.Next(); e == nil; h, e = t.Next() {
		if h.Typeflag == tar.TypeReg {
			_id, _num, _type, err := parse(h.Name)
			if err != nil {
				continue
			}
			b := bytes.NewBuffer(nil)
			enc := hex.NewEncoder(b)
			if _, _e := io.Copy(enc, t); _e != nil {
				continue
			}
			s := b.String()
			if s == "" {
				continue
			}
			c.HSetNX(ctx, fmt.Sprintf(defaultKey, _id), fmt.Sprintf(hashKey, _num, _type), s).Result()
		}
	}
	return nil
}
