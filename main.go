package main

import (
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.con/notnulldev/sdm/docker"
	"github.con/notnulldev/sdm/routes"
	"log"
	"os"
	"os/exec"
	"time"
)

func dockerPs() {
	cmd := exec.Command("docker", "ps")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	cmd.Wait()

	if err != nil {
		print("ERROR: ", err.Error())
		return
	}
	print("SUCCESS: ")
}

func getSseString(data string) string {
	return "data: " + data + "\n\n"
}

type LogsFilters struct {
	ContainerName string `json:"containerName"`
}

func getDockerStats(c echo.Context, containerName string) {
	exec.Command("docker", "logs", "")
}

func getStats(c echo.Context) error {
	var resp = c.Response()
	var headers = resp.Header()

	headers.Add("Content-Type", "text/event-stream")
	headers.Add("Cache-Control", "no-cache")
	headers.Add("Connection", "Keep-Alive")

	msgs := make(chan string)

	go docker.DockerStat(func(text string) {
		log.Printf("pushing new message")
		msgs <- text
	})

	for {
		log.Printf("waiting for next message...")
		next := <-msgs
		w, err := resp.Write([]byte(getSseString(next)))
		resp.Flush()
		if err != nil {
			log.Printf("error! [%s]", err.Error())
		} else {
			log.Printf("Successfully sent %v bytes.", w)
		}
		time.Sleep(time.Second * 2)
	}

	return nil
}

func main() {
	app := pocketbase.New()

	dockerCli, err := client.NewClientWithOpts(client.FromEnv)
	appContext := routes.AdminContext{
		app,
		dockerCli,
	}

	if err != nil {
		log.Panicf("Could not create docker client, error: [%s]", err.Error())
	}

	appContext.Pb.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		router := e.Router
		g := router.Group("/app")

		router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{"*"},
		}))

		router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./public"), true))

		g.GET("/logs", appContext.GetLogs)
		g.GET("/containers", appContext.GetContainers)
		g.GET("/containersFull", appContext.GetContainersFull)
		g.GET("/composes", appContext.GetDockerComposes)

		//router.GET("/stats", getStats)
		return nil
	})

	app.Start()
}
