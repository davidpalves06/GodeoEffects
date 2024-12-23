package videoeffects

import (
	"bufio"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

func is_Format_Supported(format string) bool {
	var supportedFormats = map[string]bool{
		"mp4": true, "avi": true, "mkv": true, "mov": true, "webm": true, "ogv": true, "flv": true, "mpeg": true, "nut": true,
	}
	return supportedFormats[format]
}

func GetFileBytes(file io.Reader) []byte {
	var fileBytesBuffer []byte = make([]byte, 1024*1024)
	var isEoF bool = false
	var fileBytes []byte = make([]byte, 0, 1024*1024)

	for !isEoF {
		bytesRead, err := file.Read(fileBytesBuffer)
		if err == io.EOF {
			isEoF = true
		} else {
			fileBytes = append(fileBytes, fileBytesBuffer[:bytesRead]...)
		}
	}
	return fileBytes
}

func parseDuration(parts []string) int64 {

	hours, _ := time.ParseDuration(parts[0] + "h")
	minutes, _ := time.ParseDuration(parts[1] + "m")

	secParts := strings.Split(parts[2], ".")
	seconds, _ := time.ParseDuration(secParts[0] + "s")
	var milliseconds time.Duration
	if len(secParts) > 1 {
		milliseconds, _ = time.ParseDuration(secParts[1] + "ms")
	}

	duration := hours + minutes + seconds + milliseconds
	return duration.Microseconds()
}

func GetInputVideoDuration(stderrPipe io.ReadCloser) int64 {
	scanner := bufio.NewScanner(stderrPipe)
	var inputVideoDuration int64 = -1
	for scanner.Scan() {
		var cmdOutput = strings.TrimSpace(strings.ToLower(scanner.Text()))
		if strings.Contains(cmdOutput, "duration") {
			duration := strings.Split(strings.Split(cmdOutput, ",")[0], ":")[1:]
			if len(duration) < 3 {
				continue
			}
			parsedDuration := parseDuration(duration)
			inputVideoDuration = parsedDuration
			break
		}
	}
	return inputVideoDuration
}

func SendProgressPercentageThroughChannel(stdoutPipe io.ReadCloser, outputVideoDuration int64, progressChannel chan uint8) {
	var progressPercentage uint8 = 0
	scanner := bufio.NewScanner(stdoutPipe)
	for scanner.Scan() {
		var cmdOutput = strings.TrimSpace(scanner.Text())
		if strings.Contains(cmdOutput, "out_time_ms") {
			us_Output_time, _ := strconv.ParseInt(strings.Split(cmdOutput, "=")[1], 10, 64)
			progressPercentage = uint8(math.Round(float64(us_Output_time) / float64(outputVideoDuration) * 100))
			progressChannel <- progressPercentage
		}
	}

	close(progressChannel)
}
