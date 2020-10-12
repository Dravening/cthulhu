package main

import (
	"fmt"
	"log"
	"os"

	"cthulhu/clog"
	"cthulhu/logic"
	"cthulhu/utils"

	"github.com/urfave/cli/v2"
)

func main() {
	sshInfo := utils.SSHInfo{}
	app := cli.NewApp()
	app.Name = "Cthulhu"
	app.Usage = `Cthulhu is an automatic tool which can easily manage network devices.
    For example:
    ./cthulhu -C "show clock" -A "192.168.2.200:22" -U test -P test`
	app.Version = "1.2.0"
	app.Authors = []*cli.Author{
		{
			Name: "Draven",
		},
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name: "address", Aliases: []string{"A"},
			Value:       "192.168.2.200:22;192.168.2.201:22",
			Usage:       "ip address  --  such as '192.168.1.1:22;192.168.1.2:22'",
			Destination: &sshInfo.AddrStr,
		},
		&cli.StringFlag{
			Name: "command", Aliases: []string{"C"},
			Value:       "show clock",
			Usage:       "commands  --  such as 'show clock;config t;do show run;exit'",
			Destination: &sshInfo.Cmd,
		},
		&cli.StringFlag{
			Name: "user", Aliases: []string{"U"},
			Value:       "admin",
			Usage:       "the account to sign in.  --  such as 'admin'",
			Destination: &sshInfo.User,
		},
		&cli.StringFlag{
			Name: "password", Aliases: []string{"P"},
			Value:       "admin",
			Usage:       "provide password to sign in.  --  such as 'admin'",
			Destination: &sshInfo.Password,
		},
		&cli.StringFlag{
			Name: "key", Aliases: []string{"K"},
			Value:       "",
			Usage:       "provide Private key to sign in(if it's necessary).",
			Destination: &sshInfo.Key,
		},
		&cli.StringFlag{
			Name: "ciphers", Aliases: []string{"CI"},
			Value:       "",
			Usage:       "provide ciphers(if it's necessary).  --  such as 'aes128-ctr;arcfour256'",
			Destination: &sshInfo.Ciphers,
		},
		&cli.Int64Flag{
			Name: "timeout", Aliases: []string{"T"},
			Value:       30,
			Usage:       "will be timeout if there's no response.  --  such as '30'",
			Destination: &sshInfo.Timeout,
		},
		&cli.IntFlag{
			Name: "routineLimit", Aliases: []string{"RL"},
			Value:       30,
			Usage:       "pause establish connection if there's too many goroutine.  --  such as '30'",
			Destination: &sshInfo.RoutineLimit,
		},
		&cli.StringFlag{
			Name: "protocol", Aliases: []string{"PR"},
			Value:       "ssh",
			Usage:       "choose ssh or telnet.  --  such as 'telnet'",
			Destination: &sshInfo.Protocol,
		},
	}
	app.Action = func(c *cli.Context) error {
		if sshInfo.Protocol == "ssh" {
			devices := logic.NewDevices(sshInfo)
			_ = logic.MultiEstablish(devices)
			return nil
		} else if sshInfo.Protocol == "telnet" {
			//to do...
		} else {
			fmt.Println("wrong protocol,protocol should be 'ssh' or 'telnet',please check.")
			return nil
		}
		return nil
	}
	app.Commands = []*cli.Command{
		{
			Name: "serve", Aliases: []string{"S"},
			Usage:  "start an rpc server",
			Action: rpcHandler,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "client", Aliases: []string{""},
					Value: "127.0.0.1:8880",
					Usage: "rpc server address",
				},
			},
		},
	}
	cli.HelpFlag = &cli.BoolFlag{
		Name: "help", Aliases: []string{"H"},
		Usage: "show options",
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name: "version", Aliases: []string{"V"},
		Usage: "print version",
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func rpcHandler(c *cli.Context) error {
	// init logger
	if err := clog.InitLogger(); err != nil {
		return err
	}
	// init rpc
	logic.StartServer()
	return nil
}
