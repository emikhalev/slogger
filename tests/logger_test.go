// +build integration

package tests

import (
	"context"
	"log"
	"os"
	"slogger/syslog"
	"testing"
	"time"

	"github.com/fsouza/go-dockerclient"

	logger "slogger"
	syslogHelper "slogger/tests/helper"
)

var (
	testSyslogPortTCP, testSyslogPortRELP = "11011", "11012"
	testSyslogAddrTCP, testSyslogAddrRELP = "127.0.0.1:" + testSyslogPortTCP, "127.0.0.1:" + testSyslogPortRELP
	testSyslogTag                         = "test-travel-logger"

	dockerClient    *docker.Client
	dockerContainer *docker.Container
)

func TestMain(m *testing.M) {
	var (
		err error
	)
	dockerClient, dockerContainer, err = syslogHelper.SyslogStart(testSyslogAddrTCP, testSyslogPortTCP, testSyslogPortRELP)
	if err != nil {
		syslogHelper.SyslogStop(dockerClient, dockerContainer)
		log.Fatal(err)
	}

	code := m.Run()
	syslogHelper.SyslogStop(dockerClient, dockerContainer)
	os.Exit(code)
}

func TestInitLog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	l, err := logger.New(context.Background(), syslog.SyslogProtocolRELP, testSyslogAddrRELP, testSyslogTag, 32, 1100*time.Millisecond, 128)
	if err != nil {
		t.Errorf("cannot init log: %v", err)
	}
	l.Close()
}

func TestCloseLog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()
	l, err := logger.New(ctx, syslog.SyslogProtocolRELP, testSyslogAddrRELP, testSyslogTag, 32, 1100*time.Millisecond, 128)
	if err != nil {
		t.Errorf("cannot init log: %v", err)
	}

	errCnt := 4
	for i := 0; i < errCnt; i++ {
		l.Err(ctx, "TestCloseLog-Message")
	}
	l.Close()

	lRecsCnt, err := syslogHelper.SyslogRecordsCount("TestCloseLog-Message")
	if err != nil {
		t.Errorf("cannot grep log: %v", err)
	}
	if lRecsCnt != errCnt {
		t.Errorf("expecting %v log records, got %v", errCnt, lRecsCnt)
	}
}

func TestSyslogCrashRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	var (
		err                            error
		lRecsCnt1, lRecsCnt2, lRecsCnt int
	)

	ctx := context.Background()
	l, err := logger.New(ctx, syslog.SyslogProtocolRELP, testSyslogAddrRELP, testSyslogTag, 32, 1100*time.Millisecond, 128)
	if err != nil {
		t.Errorf("cannot init log: %v", err)
	}

	errCnt := 16
	for i := 0; i < errCnt; i++ {
		l.Err(ctx, "TestSyslogCrashRestore-Message")
	}

	lRecsCnt1, err = syslogHelper.SyslogRecordsCount("TestSyslogCrashRestore-Message")
	if err != nil {
		t.Errorf("cannot grep log: %v", err)
	}

	syslogHelper.SyslogStop(dockerClient, dockerContainer)

	time.Sleep(1000 * time.Millisecond)

	dockerClient, dockerContainer, err = syslogHelper.SyslogStart(testSyslogAddrTCP, testSyslogPortTCP, testSyslogPortRELP)
	if err != nil {
		syslogHelper.SyslogStop(dockerClient, dockerContainer)
		t.Errorf("cannot start container: %v", err)
	}

	time.Sleep(1000 * time.Millisecond)

	l.Close()

	lRecsCnt2, err = syslogHelper.SyslogRecordsCount("TestSyslogCrashRestore-Message")
	if err != nil {
		t.Errorf("cannot grep log: %v", err)
	}

	lRecsCnt = lRecsCnt1 + lRecsCnt2

	if lRecsCnt != errCnt {
		t.Errorf("expecting %v log records, got %v", errCnt, lRecsCnt)
	}
}

func TestRELP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()
	l, err := logger.New(ctx, syslog.SyslogProtocolRELP, testSyslogAddrRELP, testSyslogTag, 32, 1100*time.Millisecond, 128)
	if err != nil {
		t.Errorf("cannot init log: %v", err)
	}

	errCnt := 8
	mes := "TestRELP-Message"
	for i := 0; i < errCnt; i++ {
		l.Err(ctx, mes)
	}
	l.Close()

	lRecsCnt, err := syslogHelper.SyslogRecordsCount(mes)
	if err != nil {
		t.Errorf("cannot grep log: %v", err)
	}
	if lRecsCnt != errCnt {
		t.Errorf("expecting %v log records, got %v", errCnt, lRecsCnt)
	}
}
