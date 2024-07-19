// Package backend : 裏方の処理
package backend

import (
	"context"
	"sync"
)

var (
	versionCheckState int
	versionNum        string
	dspath            string
)

func Start(ctx context.Context, dsp, vn string, wg *sync.WaitGroup) error {
	dspath = dsp
	versionNum = vn
	wg.Add(1)
	go monitor(ctx, wg)
	wg.Add(1)
	go mapBackend(ctx, wg)
	wg.Add(1)
	go aiBackend(ctx, wg)
	wg.Add(1)
	go networkBackend(ctx, wg)
	return nil
}

func IsLatest() bool {
	return versionCheckState != 2
}
