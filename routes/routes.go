package routes

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.con/notnulldev/sdm/docker"
	"io"
	"strings"
)

type AdminContext struct {
	Pb        *pocketbase.PocketBase
	DockerCli *client.Client
}

func (ctx *AdminContext) GetLogs(c echo.Context) error {
	var resp = c.Response()
	var headers = resp.Header()

	headers.Add("Content-Type", "text/event-stream")
	headers.Add("Cache-Control", "no-cache")
	headers.Add("Connection", "Keep-Alive")

	go docker.DockerStat(func(text string) {
		collection, err := ctx.Pb.Dao().FindCollectionByNameOrId("logs")

		if err != nil {
			fmt.Printf("Could not find collection, error: [%s]", err.Error())
		}

		log := models.NewRecord(collection)

		log.Set("logContent", text)

		err = ctx.Pb.Dao().Save(log)
		if err != nil {
			fmt.Printf("could not save log, error: [%s]", err.Error())
		}
		fmt.Printf("created new log!")
	})

	resp.WriteHeader(200)
	return nil
}

func (ctx *AdminContext) GetContainers(c echo.Context) error {

	containers, err := ctx.DockerCli.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		return err
	}

	var containerNames []string

	for _, container := range containers {
		for _, name := range container.Names {
			containerNames = append(containerNames, name)
		}
	}

	err = c.JSON(200, containerNames)
	if err != nil {
		return err
	}
	return nil
}

func (ctx *AdminContext) GetContainersFull(c echo.Context) error {

	containers, err := ctx.DockerCli.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		return err
	}

	err = c.JSON(200, containers)
	if err != nil {
		return err
	}
	return nil
}

// com.docker.compose.project

func (ctx *AdminContext) GetDockerComposes(c echo.Context) error {
	containers, err := ctx.DockerCli.ContainerList(context.Background(), types.ContainerListOptions{})

	var composes []string

	for _, contaienr := range containers {

		labels := contaienr.Labels

		for key, val := range labels {
			if key == "com.docker.compose.project" {
				var found = false
				for _, compose := range composes {
					if compose == val {
						found = true
						break
					}
				}
				if !found {
					composes = append(composes, val)
				}
			}
		}

	}

	if err != nil {
		return err
	}

	err = c.JSON(200, composes)
	if err != nil {
		return err
	}
	return nil
}

type ContainerLogsRequest struct {
	ContainerNames []string `json:"containerNames"`
	ComposesNames  []string `json:"composesNames"`
}

func (ctx *AdminContext) GetContainerLogs(c echo.Context) error {
	var req = new(ContainerLogsRequest)
	err := c.Bind(req)

	if err != nil {
		return err
	}

	var requestedContainers []types.Container
	containers, err := ctx.DockerCli.ContainerList(context.Background(), types.ContainerListOptions{})

main:
	for _, container := range containers {
		labels := container.Labels

		for _, name := range container.Names {
			for _, reqName := range req.ContainerNames {
				if strings.Contains(name, reqName) {
					requestedContainers = append(requestedContainers, container)
					continue main
				}
			}
		}

		for key, val := range labels {
			if key == "com.docker.compose.project" {
				for _, compose := range req.ComposesNames {
					if compose == val {
						requestedContainers = append(requestedContainers, container)
						break
					}
				}
			}
		}
	}

	if err != nil {
		return err
	}

	var allLogsCombined = make(map[string][]string)

	for _, c := range requestedContainers {
		if len(c.Names) == 0 {
			continue
		}

		var name = c.Names[0]

		logs, err := ctx.DockerCli.ContainerLogs(context.Background(), name, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Since:      "",
			Until:      "",
			Timestamps: true,
			Follow:     false,
			Tail:       "",
			Details:    true,
		})

		logsContent, err := io.ReadAll(logs)

		if err != nil {
			return err
		}

		var logsSplitIntoLines []string

		for _, line := range strings.Split(string(logsContent), "\n") {
			logsSplitIntoLines = append(logsSplitIntoLines, line)
		}

		name = name[1:]

		allLogsCombined[name] = logsSplitIntoLines

		if err != nil {
			return err
		}
	}

	err = c.JSON(200, allLogsCombined)
	if err != nil {
		return err
	}
	return nil
}
