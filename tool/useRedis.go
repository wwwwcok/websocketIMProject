package tool

import (
	"strconv"
	"time"
)

func CacheStoreMsg(bytes []byte, userid int64, args ...any) error {
	msg := BaseEncode(bytes)
	err := Rediscli.LPush(strconv.FormatInt(userid, 10), msg).Err()
	if err != nil {
		return err
	}

	err = Rediscli.Expire(strconv.FormatInt(userid, 10), time.Hour*24*7).Err()
	if err != nil {
		return err
	}
	return nil
}
