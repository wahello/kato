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

package main

import (
	"os"
	"sort"

	"github.com/sirupsen/logrus"

	version "github.com/gridworkz/kato/cmd"
	"github.com/gridworkz/kato/cmd/init-probe/cmd"
	"github.com/urfave/cli"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		version.ShowVersion("init-probe")
	}
	App := cli.NewApp()
	App.Version = version.GetVersion()
	App.Flags = []cli.Flag{}
	App.Commands = cmd.GetCmds()
	sort.Sort(cli.FlagsByName(App.Flags))
	sort.Sort(cli.CommandsByName(App.Commands))
	if err := App.Run(os.Args); err != nil {
		logrus.Errorf("probe cmd run failure. %s", err.Error())
		os.Exit(1)
	}
}