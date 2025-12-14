package main

import applicationcli "spyder/internal/application_cli"

func main() {

	appCli := applicationcli.NewAppCli()

	appCli.Run()
}
