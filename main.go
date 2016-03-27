package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/hashicorp/vault/api"
	"io/ioutil"
	"os"
)

func init() {

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

//func generatePermToken(vault *api.Client, token string) string {
//
//	vault.SetToken(token)
//
//	secrets, err := vault.Logical().Read("consul/creds/secrets")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	permToken := secrets.Data["token"]
//	log.WithFields(log.Fields{
//		"perm token": permToken,
//	}).Info("Create perm token")
//	return permToken
//}

func generatePermTokenReal(vault *api.Client, token string, policy string) string {
	vault.SetToken(token)

	request := &api.TokenCreateRequest{
		Policies: []string{policy},
	}

	permTokenSecret, err := vault.Auth().Token().Create(request)

	if err != nil {
		log.Fatal(err)
	}

	permToken := permTokenSecret.Auth.ClientToken
	log.WithFields(log.Fields{
		"perm token": permToken,
	}).Info("Create perm token")
	return permToken
}

func generateTempToken(vault *api.Client, token string) string {

	vault.SetToken(token)

	request := &api.TokenCreateRequest{
		NumUses: 2,
	}

	tempTokenSecret, err := vault.Auth().Token().Create(request)
	if err != nil {
		log.Fatal(err)
	}

	tmpToken := tempTokenSecret.Auth.ClientToken
	log.WithFields(log.Fields{
		"temp token": tmpToken,
	}).Info("Create temp token")
	return tmpToken
}

func main() {

	config := api.DefaultConfig()
	config.Address = "http://localhost:8200"

	vault, err := api.NewClient(config)
	if err != nil {
		panic(err)
	}

	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	ch := make(chan *docker.APIEvents)
	_ = client.AddEventListener(ch)

	for {
		event := <-ch
		gotWorker := false
		for !gotWorker {
			select {
			case _ = <-ch:

				log.WithFields(log.Fields{
					"event": event.Action,
				}).Info()

				switch event.Action {
				case "create":
					log.Info("Create event")
					permToken := generatePermTokenReal(vault, "79da82c8-7e62-f40e-55b1-2d9f113555b3", event.From)
					tmpToken := generateTempToken(vault, "79da82c8-7e62-f40e-55b1-2d9f113555b3")

					vault.ClearToken()
					log.Info("Clear token")
					vault.SetToken(tmpToken)
					log.Info("Set tmpToken")

					res, err := vault.Logical().Write("cubbyhole/perm", map[string]interface{}{"token": permToken})
					if err != nil {
						log.Fatal(err)
					}
					log.WithFields(log.Fields{
						"Res": res,
					}).Info("Write perm token in cubbyhole")

					d1 := []byte(tmpToken)
					_ = ioutil.WriteFile(fmt.Sprintf("/tmp/temp_%s", event.From), d1, 0644)

				case "start":
					log.Info("Start event")

				case "die":
					log.Info("Die event")
					log.Info("Removing file")

					os.Remove(fmt.Sprintf("/tmp/temp_%s", event.From))
					if err != nil {
						log.Fatal(err)
					}
				}
				gotWorker = true
			}
		}
	}
}
