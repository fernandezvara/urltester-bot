package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {

	app := cli.App{}
	app.Name = "URLTesterBot"
	app.Version = fmt.Sprintf("%s (%s)", Version, Commit)
	app.Usage = "Service that schedules URL monitoring by user request. \nAll the configuration is done by the final user itself by using the telegram bot."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:     "telegram-token",
			Usage:    "(required) Telegram API token for the bot to work.",
			EnvVar:   "TELEGRAM_TOKEN",
			Required: true,
		},
		cli.IntSliceFlag{
			Name:     "admins",
			Usage:    "(required) Telegram UserIDs that can administer bot accesses.",
			EnvVar:   "ADMINS",
			Required: true,
		},
		cli.StringFlag{
			Name:   "db-path",
			Usage:  "Database path where to store user requests and statuses.",
			EnvVar: "DB_PATH",
			Value:  "./urltester.db",
		},
	}

	app.Action = startAPI

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func startAPI(c *cli.Context) error {

	tester := urlTester{
		token:  c.GlobalString("telegram-token"),
		dbpath: c.GlobalString("db-path"),
		admins: c.GlobalIntSlice("admins"),
	}

	return tester.botstart()

}
