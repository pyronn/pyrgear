package comands

import (
	"testing"
)

func TestName(t *testing.T) {
	p := "./素材库/250318"
	//d, f := path.Split(p)
	ps := getDirectoryLevels(p)
	for _, f := range ps {
		println(f)
	}
	println(ps)

}
