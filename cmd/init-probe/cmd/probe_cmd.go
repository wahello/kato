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

package cmd

import (
	"github.com/gridworkz/kato/cmd/init-probe/healthy"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

//GetCmds get init probe cmds
func GetCmds() cli.Commands {
	var commands cli.Commands
	commands = append(commands, cli.Command{
		Name:  "probe",
		Usage: "probe dependent service endpoint whether ready",
		Action: func(ctx *cli.Context) error {
			logrus.Info("start probe all dependent service")
			controller, err := healthy.NewDependServiceHealthController()
			if err != nil {
				logrus.Fatalf("new probe controller failure %s", err.Error())
				return err
			}
			controller.Check()
			return nil
		},
	})
	return commands
}
