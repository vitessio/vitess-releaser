/*
Copyright 2023 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package releaser

import (
	"context"
	"time"
)

func UnwrapCtx(ctx context.Context) *Context {
	if ctx, ok := ctx.(*Context); ok {
		return ctx
	}
	panic("context should be of type releaser.Context")
}

type Context struct {
	VitessRepo   string
	MajorRelease string
}

func (c Context) Deadline() (deadline time.Time, ok bool) {
	// TODO implement me
	panic("implement me")
}

func (c Context) Done() <-chan struct{} {
	// TODO implement me
	panic("implement me")
}

func (c Context) Err() error {
	// TODO implement me
	panic("implement me")
}

func (c Context) Value(key any) any {
	// TODO implement me
	panic("implement me")
}

