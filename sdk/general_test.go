package sdk_test

import (
	"context"
	"os"
	"seborama/pcloud/sdk"
	"sync"
	"testing"
)

func Test_UserInfo(t *testing.T) {
	getAuth()
}

var pcc *sdk.Client
var lock sync.Mutex

func getAuth() {
	lock.Lock()
	defer lock.Unlock()
	if pcc != nil {
		return
	}

	username := os.Getenv("GO_PCLOUD_USERNAME")
	password := os.Getenv("GO_PCLOUD_PASSWORD")

	if username == "" || password == "" {
		panic("invalid credentials - please README.md")
	}

	pccTry := sdk.NewClient()

	ui, err := pccTry.UserInfo(context.Background(), username, password)
	requireNoError(err)
	if ui.Auth == "" {
		panic("could not obtain auth")
	}

	pcc = pccTry
}

func requireNoError(err error) {
	if err != nil {
		panic(err)
	}
}
