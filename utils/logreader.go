package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
)


type LogFilter struct {
	Component string
	FlowId string
}
//{"comp":"flow","fid":"RMgV2a79wxQMQ2o","level":"info","msg":" Node is loaded and added.","time":"2018-12-24 16:03:29.178"}
type LogEntry struct {
	Component string `json:"comp"`
	FlowId    string `json:"fid"`
	LogLevel  string `json:"level"`
	Msg       string `json:"msg"`
	Time      string `json:"time"`
}


func GetLogs(logFile string,filter *LogFilter,limitLines int) []LogEntry {
	//fmt.Print("Reading log file")
	file, err := os.Open(logFile)

	if err != nil {
		//fmt.Println("Error:"+err.Error())
		return nil
	}
	defer file.Close()

	var limit int64 = 2000000 // read limit 2 MB

	stat, err := os.Stat(logFile)
	bufsize := limit
	start := stat.Size() - limit
	if start <0 {
		start = 0
		bufsize = stat.Size()
	}
	buf := make([]byte, bufsize)

	//fmt.Println("Start position = ",start)
	_, err = file.ReadAt(buf, start)

	if err != nil {
		//fmt.Println("Error2:"+err.Error())
		return nil
	}
	bytesReader := bytes.NewReader(buf)
	scanner := bufio.NewScanner(bytesReader)
	var result []LogEntry

	scanner.Scan() // skipping first line in case if it's incomplete
	lineCounter := 0
	for scanner.Scan() {
		var logEntry LogEntry
		b := scanner.Bytes()
		err := json.Unmarshal(b,&logEntry)
		if err != nil {
			//fmt.Println("Error parsing log entry :"+string(b))
			continue
		}
		if filter.FlowId == "" || filter.FlowId == logEntry.FlowId {
			lineCounter++
			result = append(result, logEntry)
		}else {
			//fmt.Println("ENtry doesn't match ")
		}
	}
	startLine := lineCounter - limitLines
	if startLine < 0 {
		startLine = 0
	}
	return result[startLine:]
}

