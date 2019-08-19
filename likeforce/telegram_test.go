package likeforce

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetButtonText(t *testing.T) {
	f := func(likes int, message string) {
		got := getButtonText(likes, []string{"q", "w", "e", "r"})
		assert.Equal(t, message, got, "likes: %d", likes)
	}

	f(0, "q")
	f(1, "w")
	f(2, "w")
	f(9, "w")
	f(10, "e")
	f(19, "e")
	f(1000, "r")
	f(1001, "r")
}
