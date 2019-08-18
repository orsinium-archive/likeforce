package likeforce

import (
	"fmt"

	"github.com/joomcode/errorx"
	"github.com/orsinium/likeforce/likeforce/storage"
)

// ByteCount to make human-readable rating
func ByteCount(count int) string {
	const unit = 1000
	if count < unit {
		return fmt.Sprintf("%d", count)
	}
	div, exp := int64(unit), 0
	for n := count / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(count)/float64(div), "kMGTPE"[exp])
}

// UserStat to get human-readable message with user stat
func UserStat(user storage.User) (string, error) {
	posts, err := user.Posts()
	if err != nil {
		return "", errorx.Decorate(err, "cannot get user posts")
	}
	rating, err := user.Rating().Get()
	if err != nil {
		return "", errorx.Decorate(err, "cannot get user rating")
	}
	if posts == nil {
		return "First blood!", nil
	}
	const tmpl = "*user stat:*\n\nposts: `%s`\nrating: `%s`"
	return fmt.Sprintf(tmpl, ByteCount(len(posts)), ByteCount(rating)), nil
}
