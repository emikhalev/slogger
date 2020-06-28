// +build integration

package relp

import (
	"log"
	sl "log/syslog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fsouza/go-dockerclient"

	logger "slogger/syslog/relp"
	syslogHelper "slogger/tests/helper"
)

var (
	testSyslogPortTCP, testSyslogPortRELP     = "11014", "11015"
	testSyslogAddr     = "127.0.0.1:" + testSyslogPortRELP
	testSyslogProtocol = "tcp"
	testSyslogTag      = "test-travel-logger"

	dockerClient    *docker.Client
	dockerContainer *docker.Container
)

func TestMain(m *testing.M) {
	var (
		err error
	)
	dockerClient, dockerContainer, err = syslogHelper.SyslogStart(testSyslogAddr, testSyslogPortTCP, testSyslogPortRELP)
	if err != nil {
		syslogHelper.SyslogStop(dockerClient, dockerContainer)
		log.Fatal(err)
	}

	code := m.Run()
	syslogHelper.SyslogStop(dockerClient, dockerContainer)
	os.Exit(code)
}

func TestDial(t *testing.T) {
	c, err := logger.Dial(testSyslogAddr, sl.LOG_WARNING|sl.LOG_DAEMON, testSyslogTag, 3 * time.Second)
	if err != nil {
		t.Errorf("RELP client.Dial expecting not error, got: %v", err)
	}

	c.Close()
}

func TestWrite(t *testing.T) {
	c, err := logger.Dial(testSyslogAddr, sl.LOG_WARNING|sl.LOG_DAEMON, testSyslogTag, 3 * time.Second)
	if err != nil {
		t.Errorf("RELP client.Dial expecting not error, got: %v", err)
	}

	testMsg := "TEST-RELP-Send-String-Message"
	errCnt := 9
	for i:=0;i<errCnt;i++ {
		if _, err:=c.Write([]byte(testMsg + "-" + strconv.Itoa(i))); err!=nil {
			t.Errorf("cannot send string: %v", err)
		}
	}
	c.Close()
	time.Sleep(500 * time.Millisecond)

	lRecsCnt, err := syslogHelper.SyslogRecordsCount(testMsg)
	if err != nil {
		t.Errorf("cannot grep log: %v", err)
	}

	if lRecsCnt != errCnt {
		t.Errorf("expecting %v log records, got %v", errCnt, lRecsCnt)
	}

}