package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceInlineImage(t *testing.T) {
	for name, tc := range map[string]struct {
		in   string
		want string
	}{
		"without image": {
			in:   "loreum ipsum \ninfo\ntext text",
			want: "loreum ipsum \ninfo\ntext text",
		},
		"with start image": {
			in: `![image](ipfs://bafkreicrzy34xp6i2cyl4nkn6u4pofiu6zrwbct7aa47dm77c43u2b26zq)
## TL;DR
### Council recommendation: Positive 🙂
### What`,
			want: `![image](ipfs://bafkreicrzy34xp6i2cyl4nkn6u4pofiu6zrwbct7aa47dm77c43u2b26zq)
## TL;DR
### Council recommendation: Positive 🙂
### What`,
		},
		"with inline image": {
			in: `![image](ipfs://bafkreicrzy34xp6i2cyl4nkn6u4pofiu6zrwbct7aa47dm77c43u2b26zq)
## TL;DR
### Council recommendation: Positive 🙂

## Solution Proposed
![image](ipfs://bafkreidexhqro3q22nnsvzm7svxotoemvp4xyzydbrny2tcgwl54jdk2ea)
Game Maker
* Allow Window Resizing, both horizontal and vertical
![banner-GJ.jpg](ipfs://bafybeigo4xzgfoyb7quofhs7zx7yfxlnqagsyuyi7owznrqrsj2dunqa44)
* Allow Resolution Sizes that match 420p,720p, 1080p, 2K, 4K`,
			want: `![image](ipfs://bafkreicrzy34xp6i2cyl4nkn6u4pofiu6zrwbct7aa47dm77c43u2b26zq)
## TL;DR
### Council recommendation: Positive 🙂

## Solution Proposed

![image](ipfs://bafkreidexhqro3q22nnsvzm7svxotoemvp4xyzydbrny2tcgwl54jdk2ea)

Game Maker
* Allow Window Resizing, both horizontal and vertical

![banner-GJ.jpg](ipfs://bafybeigo4xzgfoyb7quofhs7zx7yfxlnqagsyuyi7owznrqrsj2dunqa44)

* Allow Resolution Sizes that match 420p,720p, 1080p, 2K, 4K`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual := ReplaceInlineImages(tc.in)
			assert.Equal(t, tc.want, actual)
		})
	}
}
