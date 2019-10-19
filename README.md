# urltester-bot

Simple telegram bot that monitors URLs and alerts the alert owner on status changes.

The configuration is done by the users who got permissions from the administrators, so at least one administrator must be configured to grant and/or revoke permissions when need.

# Bot usage (from Telegram)

Bot have three different set of commands that are executable based on the current user permissions.

`Basic help` for anonymous users. `User help` for users with granted access permissions. And `Admin help`, well for admins.

```

HELP
/history - shows last commands sent to the bot
/start - starts the bot asking for permissions from an administrator
/help - this help

USER COMMANDS
/summary - shows current status of monitors subscribed to
/monitors - shows current defined monitors
/newmonitor <method> <url> <statuscode> <interval> <private> - creates a new monitor with basic monitor settings
/remove <id> - removes a monitor
/setinterval <id> <interval> - updates the interval between monitor calls
/settext <id> <text> - updates the expected text to be found on every monitor request
/settimeout <id> <timeout> - updates the expected timeout to be found on every monitor request
/setstatuscode <id> <statuscode> - updates the expected HTTP status code for the monitor
/subscribe <id> - subscrives to monitor status
/unsubscribe <id> - unsubscribes from monitor status
/test <method> <url> <statuscode> - test a URL and its HTTP status
/testfull <method> <url> <statuscode> - test a URL and its HTTP status, returning its data

ADMIN COMMANDS
/users - returns users and its current status
/grant <id> - grants access to a user
/revoke <id> - revokes access to a user

```

## Monitor settings

- `ID`. This is set by the service on monitor creation.
- `Method`. HTTP method to use for connect the URL (GET, POST, PUT and OPTIONS are allowed).
- `URL`. Full path to the resource to monitor. Ex: `https://www.example.com/what/to/monitor.html`
- `StatusCode`. Expected Status Code the resource must return on every request.
- `Interval`. Amount of time between requests.
- `Private`. You can set the monitor can be private to you or not.
- `Text`. The text you expect to find on every request or set the monitor as failure.
- `Timeout`. Maximun timeout for the requests before set monitor as failure.


# Basic usage (service)

The bot have two different required options to run, Telegram Token and at least one administrator.

## How-To get a Telegram Bot API Token.

Open your Telegram client and search for the bot `BotFather`.

To create a new token/bot you need type `/newbot` and answer these questions ( Botfather guide you):

- `/newbot` to start the wizard.
- Name for the bot. Anything must fit but some rules could be.
- Username for the bot. It must end in `_bot`.

You will have something like: 
```
Done! Congratulations on your new bot. You will find it at t.me/demobotexample_bot. You can now add a description, about section and profile picture for your bot, see /help for a list of commands. By the way, when you've finished creating your cool bot, ping our Bot Support if you want a better username for it. Just make sure the bot is fully operational before you do this.

Use this token to access the HTTP API:
`123456789:ASDFGasdfgzxcvZXCVqwerQEWR`
Keep your token secure and store it safely, it can be used by anyone to control your bot.

For a description of the Bot API, see this page: https://core.telegram.org/bots/api
````

Your api key is `123456789:ASDFGasdfgzxcvZXCVqwerQEWR`, the one that the bot expects to run with the name and username you select on the wizard.

## How-To know your own Telegram User ID.

There is a bot that helps, search for `userinfobot`. As soon as you enter in the channel or send the command `/start` it will return your details (username, id, first and last name and your current language).


# Start the service

Service allows you to use environment variables or options to set its configuration:

- `--telegram-token` or env var `$TELEGRAM_TOKEN` is the token for our bot
- `--db-path` or env var `$DB_PATH` sets the location for the database where to store the running data. By default it stores on the same directory where it starts `./urltester.db`
- `--admins` or env var `$ADMINS` is the array of administrators of the application. Note that using flags you must set it for every administrator.

Example:

```
# using flags
./urltester-bot --telegram-token 123456789:ASDFGasdfgzxcvZXCVqwerQEWR --admins 12345678 --admins 23456789 --admins 34567890 --db-path /opt/urltester/db/urltester.db

# using environment variables
TELEGRAM_TOKEN="123456789:ASDFGasdfgzxcvZXCVqwerQEWR" ADMINS="12345678,23456789,34567890" DB_PATH="/opt/urltester/db/urltester.db" ./urltester-bot 
```


## Get some help

```
./urltester-bot help
NAME:
   urltester - Service that schedules URL monitoring by user request.
All the configuration is done by the final user itself by using the telegram bot.

USAGE:
    [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --telegram-token value  Telegram API token for the bot to work. [$TELEGRAM_TOKEN]
   --db-path value         Database path where to store user requests and statuses. (default: "./urltester.db") [$DB_PATH]
   --admins value          Telegram UserIDs that can administer bot accesses. [$ADMINS]
   --help, -h              show help
   --version, -v           print the version
```



# Changes:

0.3.0:

- Problems now explain the failure
- Markdown messages updated

0.2.1:

- /about command

0.2.0:

- Full rewrite of command properties
- Full rewrite of payload gathering
- Full rewrite of help command
- Added ability of search for a text
- Added ability of monitor for timeout


0.1.0: 

- Administrators with /users, /grant, /revoke new commands
- Permissions requests

0.0.6:

- User subscriptions

0.0.5:

- Added timeline events
- Notify on status change

0.0.4:

- Notify on every monitor check

0.0.3:

- Accesory commands
- /help
- /summany

0.0.2:

- /monitors

0.0.1:

- Monitor add / remove
- Monitor HTTP status on interval request
