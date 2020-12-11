package sdk_test

import (
	"context"
	"fmt"
	"os"
	"seborama/pcloud/sdk"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	exitVal := 0

	func() {
		initAuthenticatedClient()
		defer logoutClient()

		exitVal = m.Run()
	}()

	os.Exit(exitVal)
}

var pcc *sdk.Client
var lock sync.Mutex

func initAuthenticatedClient() {
	lock.Lock()
	defer lock.Unlock()

	if pcc != nil {
		return
	}

	username := os.Getenv("GO_PCLOUD_USERNAME")
	password := os.Getenv("GO_PCLOUD_PASSWORD")

	if username == "" || password == "" {
		panic("invalid credentials - please see README.md")
	}

	pccTry := sdk.NewClient()

	err := pccTry.Login(
		context.Background(),
		sdk.WithGlobalOptionUsername(username),
		sdk.WithGlobalOptionPassword(password),
		sdk.WithGlobalOptionAuthInactiveExpire(5*time.Minute),
	)
	requireNoError(err)

	pcc = pccTry
}

func logoutClient() {
	lr, err := pcc.Logout(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Auth token deleted:", lr.AuthDeleted)
}

func requireNoError(err error) {
	if err != nil {
		panic(err)
	}
}
