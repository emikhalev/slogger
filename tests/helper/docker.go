package helper

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var (
	dockerPortTCP  docker.Port = "514"
	dockerPortRELP docker.Port = "1601"
	dockerImage                = "rsyslog/syslog_appliance_alpine"

	dockerClient    *docker.Client
	dockerContainer *docker.Container

	hostname = ""
)

func SyslogStart(testSyslogAddr, testSyslogPortTCP, testSyslogPortRELP string) (*docker.Client, *docker.Container, error) {
	dockerClient, dockerContainer = dockerSetup(testSyslogPortTCP, testSyslogPortRELP)
	if err := dockerClient.StartContainer(dockerContainer.ID, &docker.HostConfig{
		PortBindings: map[docker.Port][]docker.PortBinding{
			dockerPortTCP:  {{HostPort: testSyslogPortTCP}},
			dockerPortRELP: {{HostPort: testSyslogPortRELP}},
		},
	}); err != nil {
		return dockerClient, dockerContainer, fmt.Errorf("cannot start Docker container: %s", err)
	}
	if err := dockerWaitStarted(dockerClient, dockerContainer.ID, 5*time.Second); err != nil {
		return dockerClient, dockerContainer, fmt.Errorf("container not started: %v", err)
	}

	if err := dockerWaitReachable(testSyslogAddr, 5*time.Second); err != nil {
		return dockerClient, dockerContainer, fmt.Errorf("container not reachable: %v", err)
	}

	if hn, err := os.Hostname(); err != nil {
		SyslogStop(dockerClient, dockerContainer)
		return nil, nil, fmt.Errorf("cannot get hostname: %v", err)
	} else {
		hostname = hn
	}

	return dockerClient, dockerContainer, nil
}

func SyslogStop(client *docker.Client, container *docker.Container) {
	if client == nil || container == nil {
		return
	}
	if err := client.StopContainer(container.ID, 10); err != nil {
		log.Printf("Cannot create Docker container: %s", err)
	}
	if err := client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    container.ID,
		Force: true,
	}); err != nil {
		log.Fatalf("cannot remove container: %s", err)
	}
}

func SyslogRecordsCount(grepPattern string) (int, error) {
	logFile := fmt.Sprintf("/logs/hosts/%s/messages.log", hostname)
	o, err := dockerExec("grep", grepPattern, logFile)
	if err != nil {
		return -1, fmt.Errorf("cannot grep log: %v", err)
	}
	lRecs := strings.Split(o, "\r")
	return len(lRecs) - 1, nil
}

func dockerExec(cmd ...string) (string, error) {
	ce, err := dockerClient.CreateExec(docker.CreateExecOptions{
		Container:    dockerContainer.ID,
		AttachStdout: true,
		Tty:          true,
		Cmd:          cmd,
	})
	if err != nil {
		return "", err
	}

	b := new(bytes.Buffer)
	err = dockerClient.StartExec(ce.ID, docker.StartExecOptions{
		Detach:       false,
		Tty:          true,
		OutputStream: b,
		RawTerminal:  true,
	})
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func dockerSetup(testSyslogPortTCP, testSyslogPortRELP string) (*docker.Client, *docker.Container) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Cannot connect to Docker daemon: %s", err)
	}
	c, err := client.CreateContainer(dockerCreateOptions(dockerImage, testSyslogPortTCP, testSyslogPortRELP))
	if err != nil {
		log.Fatalf("Cannot create Docker container: %s", err)
	}
	return client, c
}

func dockerCreateOptions(dbname, testSyslogPortTCP, testSyslogPortRELP string) docker.CreateContainerOptions {
	ports := make(map[docker.Port]struct{})
	//ports["514"] = struct{}{} // udp
	ports[dockerPortTCP] = struct{}{}  // tcp
	ports[dockerPortRELP] = struct{}{} // relp
	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        dbname,
			ExposedPorts: ports,
		},
		HostConfig: &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				dockerPortTCP:  {{HostIP: "127.0.0.1", HostPort: testSyslogPortTCP}},
				dockerPortRELP: {{HostIP: "127.0.0.1", HostPort: testSyslogPortRELP}},
			},
		},
	}

	return opts
}

func dockerWaitStarted(client *docker.Client, id string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := client.InspectContainer(id)
		if err != nil {
			break
		}
		if c.State.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start container %s for %v", id, maxWait)
}

func dockerWaitReachable(addr string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := net.Dial("tcp", addr)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot connect %v for %v", addr, maxWait)
}
