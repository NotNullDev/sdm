package main

import (
	"bufio"
	"github.com/labstack/echo/v4"
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

type dockerStatCallback func(text string)

func dockerStat(callback dockerStatCallback) {
	cmd := exec.Command("docker", "stats")

	out, _ := cmd.StdoutPipe()

	cmd.Start()

	scanner := bufio.NewScanner(out)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		newLine := scanner.Text()
		callback(newLine)
	}

	cmd.Wait()

}
func getSseStroging(data string) string {
	return "data: " + data + "\n\n"
}
func main() {
	router := echo.New()

	router.File("/", "./index.html")

	router.GET("/sse", func(c echo.Context) error {

		var resp = c.Response()
		var headers = resp.Header()

		headers.Add("Content-Type", "text/event-stream")
		headers.Add("Cache-Control", "no-cache")
		headers.Add("Connection", "Keep-Alive")

		for {
			w, err := resp.Write([]byte(getSseStroging("hi!")))
			resp.Flush()
			if err != nil {
				log.Printf("error! [%s]", err.Error())
			} else {
				log.Printf("Successfully sent %v bytes.", w)
			}
			time.Sleep(time.Second * 2)
		}

	})

	router.Start(":9000")

}
