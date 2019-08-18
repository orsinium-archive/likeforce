package storage

import "fmt"

const prefix = "likeforce"

func makeKeyPosts(chat int64) string {
	return fmt.Sprintf("%s:%d:posts:set", prefix, chat)
}

func makeKeyPostLikes(chat int64, post int) string {
	return fmt.Sprintf("%s:chat-%d:post-%d:likes:int", prefix, chat, post)
}

func makeKeyUserLikes(chat int64, user int) string {
	return fmt.Sprintf("%s:chat-%d:user-%d:set", prefix, chat, user)
}

func makeValuePost(chat int64, post int) string {
	return fmt.Sprintf("%d:%d", chat, post)
}
