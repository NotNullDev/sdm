package docker

import (
	"bufio"
	"os/exec"
)

type dockerStatCallback func(text string)

func DockerStat(callback dockerStatCallback) {
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
