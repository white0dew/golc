package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"io/ioutil"
	"time"
)

func main() {
	config := openai.DefaultConfig("sk-peisUrRs7gPLZKPk3c758475E6604f87B427Df3f4f34Cd45")
	config.BaseURL = "https://35.nekoapi.com/v1"
	client := openai.NewClientWithConfig(config)

	startTime := time.Now()
	res, err := client.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Input:          "今天给大家分享一条非常神奇的Prompt（指令），",
		ResponseFormat: openai.SpeechResponseFormatOpus,
		Voice:          openai.VoiceAlloy,
	})
	fmt.Println(time.Since(startTime))

	//audio, err := io.ReadAll(res)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	fmt.Println(err)
	// 创建buffer来读取数据
	buffer := make([]byte, 4096) // 以适合你数据的大小调整这个大小

	for {
		// 从音频流读取数据
		_, err := res.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// 音频流结束
				break
			}
			// 其他错误处理
			break
		}
		fmt.Println("receive")
	}

	//d, err := mp3.NewDecoder(bytes.NewBuffer(buffer))
	//if err != nil {
	//	return
	//}

	//c, ready, err := oto.NewContext(d.SampleRate(), 2, 2)
	//if err != nil {
	//	return
	//}
	//<-ready
	//
	//p := c.NewPlayer(d)
	//defer p.Close()
	//p.Play()
	//
	//for {
	//	time.Sleep(time.Second)
	//	if !p.IsPlaying() {
	//		break
	//	}
	//}
}

func ToString(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func ReadAllFromReadCloser(rc io.ReadCloser) (string, error) {
	defer rc.Close() // 确保io.ReadCloser在函数返回时关闭

	data, err := ioutil.ReadAll(rc) // 读取全部内容
	if err != nil {
		return "", err // 如果发生错误，返回错误
	}

	return string(data), nil // 将读取到的字节切片转换为字符串并返回
}
