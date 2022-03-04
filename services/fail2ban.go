package services

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/ttruong15/monitorLog"
)

type fail2ban struct {
	banIPs map[string]int
}

func NewFail2banService() monitorLog.Processor {
	return &fail2ban{
		banIPs: make(map[string]int),
	}
}

func (f *fail2ban) Processing() error {

	fileLocation, ok := monitorLog.AvailableServices["fail2ban"]
	if !ok {
		log.Println("missing fail2ban log file")
	}

	log.Println("I am call", fileLocation)

	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewReader(file)
	for {
		line, _, err := scanner.ReadLine()
		if err == io.EOF {
			break
		}

		lineStr := string(line)
		if !strings.Contains(lineStr, "fail2ban.actions") {
			continue
		}

		// re := regexp.MustCompile(".* file2ban.actions .* (Ban) ([0-9]+?) .*")
		re := regexp.MustCompile("Ban (\\d+\\.\\d+\\.\\d+\\.\\d+)")
		match := re.FindStringSubmatch(lineStr)

		if len(match) > 0 {
			foundIP := match[1]
			_, ok := f.banIPs[foundIP]
			if ok {
				f.banIPs[foundIP] += 1
			} else {
				f.banIPs[foundIP] = 1
			}
		}
	}

	return nil
}

func (f *fail2ban) GenerateReport() string {
	reportString := ""

	for ip, count := range f.banIPs {
		reportString += fmt.Sprintf("%-16v => %4v\n", ip, count)
	}

	return reportString
}
