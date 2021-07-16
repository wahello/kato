// KATO, Application Management Platform
// Copyright (C) 2021 Gridworkz Co., Ltd.

// Permission is hereby granted, free of charge, to any person obtaining a copy of this 
// software and associated documentation files (the "Software"), to deal in the Software
// without restriction, including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons 
// to whom the Software is furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all copies or 
// substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, 
// INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
// PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
// FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package probe

import (
	"context"
	"fmt"
	"testing"

	"github.com/gridworkz/kato/node/nodem/service"
)

func TestHttpProbe_Check(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	resultChannel := make(chan *service.HealthStatus, 10)
	httpProbe := HttpProbe{
		Name:         "builder",
		Address:      "127.0.0.1:9090/health",
		Ctx:          ctx,
		Cancel:       cancel,
		ResultsChan:  resultChannel,
		TimeInterval: 3,
	}
	go httpProbe.Check()
	for {
		result := <-resultChannel
		fmt.Println(result.Name, result.Status, result.Info)
	}
}
